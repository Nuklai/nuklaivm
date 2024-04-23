// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/timer"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/cmd/nuklai-feed/config"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type FeedContent struct {
	Message string `json:"message"`
	URL     string `json:"url"`
}

type FeedObject struct {
	SubnetID  string `json:"subnetID"`
	ChainID   string `json:"chainID"`
	Address   string `json:"address"`
	TxID      ids.ID `json:"txID"`
	Timestamp int64  `json:"timestamp"`
	Fee       uint64 `json:"fee"`

	Content *FeedContent `json:"content"`
}

type Manager struct {
	log    logging.Logger
	config *config.Config

	ncli     *nrpc.JSONRPCClient
	subnetID ids.ID
	chainID  ids.ID

	l             sync.RWMutex
	t             *timer.Timer
	epochStart    int64
	epochMessages int
	feeAmount     uint64

	f sync.RWMutex
	// TODO: persist this
	feed       []*FeedObject
	cancelFunc context.CancelFunc
}

func New(logger logging.Logger, config *config.Config) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cli := rpc.NewJSONRPCClient(config.NuklaiRPC)
	networkID, subnetID, chainID, err := cli.Network(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	ncli := nrpc.NewJSONRPCClient(config.NuklaiRPC, networkID, chainID)
	m := &Manager{log: logger, config: config, ncli: ncli, subnetID: subnetID, chainID: chainID, feed: []*FeedObject{}, cancelFunc: cancel}
	m.epochStart = time.Now().Unix()
	m.feeAmount = m.config.MinFee
	m.t = timer.NewTimer(m.updateFee)
	m.log.Info("feed initialized",
		zap.Uint32("network ID", networkID),
		zap.String("subnet ID", subnetID.String()),
		zap.String("chain ID", chainID.String()),
		zap.String("address", m.config.Recipient),
		zap.String("fee", utils.FormatBalance(m.feeAmount, nconsts.Decimals)),
	)
	return m, nil
}

// updateFee adjusts the fee based on the message frequency during an epoch
func (m *Manager) updateFee() {
	m.l.Lock()
	defer m.l.Unlock()

	// If time since [epochStart] is within half of the target duration,
	// we attempted to update fee when we just reset during block processing.
	now := time.Now().Unix()
	if now-m.epochStart < m.config.TargetDurationPerEpoch/2 {
		return
	}

	// Decrease fee if there are no messages in this epoch
	if m.feeAmount > m.config.MinFee && m.epochMessages == 0 {
		m.feeAmount -= m.config.FeeDelta
		m.log.Info("decreasing message fee", zap.Uint64("fee", m.feeAmount))
	}
	m.epochMessages = 0
	m.epochStart = time.Now().Unix()
	m.t.SetTimeoutIn(time.Duration(m.config.TargetDurationPerEpoch) * time.Second)
}

func (m *Manager) Run(ctx context.Context) error {
	// Start update timer
	m.t.SetTimeoutIn(time.Duration(m.config.TargetDurationPerEpoch) * time.Second)
	go m.t.Dispatch()
	defer m.t.Stop()

	var scli *rpc.WebSocketClient
	currentRPCURL := m.config.NuklaiRPC

	reconnect := func() error {
		var err error
		if scli != nil {
			scli.Close()
		}
		scli, err = rpc.NewWebSocketClient(m.config.NuklaiRPC, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
		if err != nil {
			m.log.Warn("Failed to connect to RPC", zap.String("uri", m.config.NuklaiRPC), zap.Error(err))
			return fmt.Errorf("failed to connect to RPC: %w", err)
		}
		if err = scli.RegisterBlocks(); err != nil {
			m.log.Warn("failed to register for blocks", zap.String("uri", m.config.NuklaiRPC), zap.Error(err))
			return fmt.Errorf("failed to register for blocks: %w", err)
		}
		return nil
	}

	if err := reconnect(); err != nil {
		m.log.Error("Initial RPC connection failed", zap.Error(err))
		return err
	}

	for ctx.Err() == nil {
		if m.config.NuklaiRPC != currentRPCURL {
			m.log.Info("Detected RPC URL change, reconnecting", zap.String("newURL", m.config.NuklaiRPC))
			if err := reconnect(); err != nil {
				m.log.Error("Reconnection failed", zap.Error(err))
				continue
			}
			currentRPCURL = m.config.NuklaiRPC
		}

		parser, err := m.ncli.Parser(ctx)
		if err != nil {
			m.log.Error("Failed to create parser", zap.Error(err))
			return err
		}

		blk, results, _, err := scli.ListenBlock(ctx, parser)
		if err != nil {
			m.log.Warn("Unable to listen for blocks", zap.Error(err))
			time.Sleep(10 * time.Second)
			continue
		}

		// Look for transactions to recipient
		for i, tx := range blk.Txs {
			action, ok := tx.Action.(*actions.Transfer)
			recipientAddr, err := m.config.RecipientAddress()
			if err != nil {
				return err
			}
			if !ok || action.To != recipientAddr {
				continue
			}

			result := results[i]
			fromStr := codec.MustAddressBech32(nconsts.HRP, tx.Auth.Actor())
			if !result.Success || action.Value < m.feeAmount {
				m.log.Info("Incoming message failed or did not pay enough", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Uint64("required", m.feeAmount))
				continue
			}

			var content FeedContent
			if err := json.Unmarshal(action.Memo, &content); err != nil || len(content.Message) == 0 {
				m.log.Info("Incoming message could not be parsed or was empty", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Error(err))
				continue
			}

			m.l.Lock()
			m.f.Lock()
			m.feed = append([]*FeedObject{{
				SubnetID:  m.subnetID.String(),
				ChainID:   m.chainID.String(),
				Address:   fromStr,
				TxID:      tx.ID(),
				Timestamp: blk.Tmstmp,
				Fee:       action.Value,
				Content:   &content,
			}}, m.feed...)
			if len(m.feed) > m.config.FeedSize {
				m.feed = m.feed[:m.config.FeedSize]
			}
			m.log.Info("Received incoming message", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value))
			m.f.Unlock()
			m.l.Unlock()
		}

		time.Sleep(1 * time.Second)
	}

	return ctx.Err()
}

// GetFeedInfo provides details about the feed and the current fee
func (m *Manager) GetFeedInfo(_ context.Context) (codec.Address, uint64, error) {
	m.l.RLock()
	defer m.l.RUnlock()

	addr, err := m.config.RecipientAddress()
	return addr, m.feeAmount, err
}

// GetFeed returns a copy of the current feed
func (m *Manager) GetFeed(_ context.Context, subnetID, chainID string) ([]*FeedObject, error) {
	m.f.RLock()
	defer m.f.RUnlock()

	var filteredFeed []*FeedObject
	for _, item := range m.feed {
		if item.SubnetID == subnetID && item.ChainID == chainID {
			filteredFeed = append(filteredFeed, item)
		}
	}

	return slices.Clone(filteredFeed), nil
}

// UpdateNuklaiRPC updates the RPC URL and reconnects clients
func (m *Manager) UpdateNuklaiRPC(ctx context.Context, newNuklaiRPCUrl string) error {
	m.l.Lock()
	defer m.l.Unlock()

	m.log.Info("Updating Nuklai RPC URL", zap.String("oldURL", m.config.NuklaiRPC), zap.String("newURL", newNuklaiRPCUrl))

	// Updating the configuration
	m.config.NuklaiRPC = newNuklaiRPCUrl

	// Re-initialize RPC clients
	cli := rpc.NewJSONRPCClient(newNuklaiRPCUrl)
	networkID, subnetID, chainID, err := cli.Network(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch network details: %w", err)
	}

	// Reassign the newly created clients
	m.ncli = nrpc.NewJSONRPCClient(newNuklaiRPCUrl, networkID, chainID)

	// Reinitialize dependent properties
	m.subnetID = subnetID
	m.chainID = chainID
	m.epochStart = time.Now().Unix()
	m.feeAmount = m.config.MinFee
	m.t = timer.NewTimer(m.updateFee)

	m.log.Info("RPC client has been updated and manager reinitialized",
		zap.String("new RPC URL", newNuklaiRPCUrl),
		zap.Uint32("network ID", networkID),
		zap.String("subnet ID", subnetID.String()),
		zap.String("chain ID", chainID.String()),
		zap.String("address", m.config.Recipient),
		zap.String("fee", utils.FormatBalance(m.feeAmount, nconsts.Decimals)),
	)

	return nil
}
