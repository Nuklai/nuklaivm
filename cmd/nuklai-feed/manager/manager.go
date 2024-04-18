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
	Address   string `json:"address"`
	TxID      ids.ID `json:"txID"`
	Timestamp int64  `json:"timestamp"`
	Fee       uint64 `json:"fee"`

	Content *FeedContent `json:"content"`
}

type Manager struct {
	log    logging.Logger
	config *config.Config

	ncli *nrpc.JSONRPCClient

	l             sync.RWMutex
	t             *timer.Timer
	epochStart    int64
	epochMessages int
	feeAmount     uint64

	f sync.RWMutex
	// TODO: persist this
	feed       []*FeedObject
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup // to control the Run method execution
}

func New(logger logging.Logger, config *config.Config) (*Manager, error) {
	// Create a cancellable context at the start of the function.
	ctx, cancel := context.WithCancel(context.Background())

	// Declare err early to make it accessible inside the defer function
	var err error

	// Ensure that the cancel function is called if this function exits
	// after the context is created, but before it is stored in the Manager struct.
	defer func() {
		// Only call cancel if returning with an error,
		// because otherwise, the cancel function will be stored in the Manager struct.
		if err != nil {
			cancel()
		}
	}()

	cli := rpc.NewJSONRPCClient(config.NuklaiRPC)
	networkID, _, chainID, err := cli.Network(ctx)
	if err != nil {
		return nil, err
	}
	ncli := nrpc.NewJSONRPCClient(config.NuklaiRPC, networkID, chainID)
	m := &Manager{log: logger, config: config, ncli: ncli, feed: []*FeedObject{}, cancelFunc: cancel}
	m.epochStart = time.Now().Unix()
	m.feeAmount = m.config.MinFee
	m.log.Info("feed initialized",
		zap.String("address", m.config.Recipient),
		zap.String("fee", utils.FormatBalance(m.feeAmount, nconsts.Decimals)),
	)
	m.t = timer.NewTimer(m.updateFee)
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

// Run processes blocks and messages, updating the internal feed
func (m *Manager) Run(ctx context.Context) error {
	// Start update timer
	m.t.SetTimeoutIn(time.Duration(m.config.TargetDurationPerEpoch) * time.Second)
	go m.t.Dispatch()
	defer m.t.Stop()

	// Continuously listen for new blocks and process transactions directed to the recipient
	parser, err := m.ncli.Parser(ctx)
	if err != nil {
		return err
	}
	recipientAddr, err := m.config.RecipientAddress()
	if err != nil {
		return err
	}

	// Connection loop for robustness
	for ctx.Err() == nil {
		scli, err := rpc.NewWebSocketClient(m.config.NuklaiRPC, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
		if err != nil {
			m.log.Warn("unable to connect to RPC", zap.String("uri", m.config.NuklaiRPC), zap.Error(err))
			time.Sleep(10 * time.Second)
			continue
		}
		if err := scli.RegisterBlocks(); err != nil {
			m.log.Warn("unable to connect to register for blocks", zap.String("uri", m.config.NuklaiRPC), zap.Error(err))
			time.Sleep(10 * time.Second)
			continue
		}
		for ctx.Err() == nil {
			m.log.Info("Listening for blocks...")
			// Listen for blocks
			blk, results, _, err := scli.ListenBlock(ctx, parser)
			if err != nil {
				m.log.Warn("unable to listen for blocks", zap.Error(err))
				continue // Ensure the loop continues or handle reconnection logic here
			}

			// Look for transactions to recipient
			for i, tx := range blk.Txs {
				action, ok := tx.Action.(*actions.Transfer)
				if !ok || action.To != recipientAddr || len(action.Memo) == 0 {
					continue
				}

				result := results[i]
				from := tx.Auth.Actor()
				fromStr := codec.MustAddressBech32(nconsts.HRP, from)
				if !result.Success {
					m.log.Info("incoming message failed on-chain", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Uint64("required", m.feeAmount))
					continue
				}
				if action.Value < m.feeAmount {
					m.log.Info("incoming message did not pay enough", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Uint64("required", m.feeAmount))
					continue
				}

				var c FeedContent
				if err := json.Unmarshal(action.Memo, &c); err != nil {
					m.log.Info("incoming message could not be parsed", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Error(err))
					continue
				}
				if len(c.Message) == 0 {
					m.log.Info("incoming message was empty", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value))
					continue
				}

				// Add to feed
				m.l.Lock()
				m.f.Lock()
				m.feed = append([]*FeedObject{{
					Address:   fromStr,
					TxID:      tx.ID(),
					Timestamp: blk.Tmstmp,
					Fee:       action.Value,
					Content:   &c,
				}}, m.feed...)
				if len(m.feed) > m.config.FeedSize {
					// Trim the feed to prevent unbounded growth
					m.feed[m.config.FeedSize] = nil // prevent memory leak
					m.feed = m.feed[:m.config.FeedSize]
				}
				m.epochMessages++
				if m.epochMessages >= m.config.MessagesPerEpoch {
					m.feeAmount += m.config.FeeDelta
					m.log.Info("increasing message fee", zap.Uint64("fee", m.feeAmount))
					m.epochMessages = 0
					m.epochStart = time.Now().Unix()
					m.t.Cancel()
					m.t.SetTimeoutIn(time.Duration(m.config.TargetDurationPerEpoch) * time.Second)
				}
				m.log.Info("received incoming message", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Uint64("new required", m.feeAmount))
				m.f.Unlock()
				m.l.Unlock()
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Sleep before trying again
		time.Sleep(10 * time.Second)
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
func (m *Manager) GetFeed(context.Context) ([]*FeedObject, error) {
	m.f.RLock()
	defer m.f.RUnlock()

	return slices.Clone(m.feed), nil
}

func (m *Manager) RestartRun(ctx context.Context) {
	if m.cancelFunc != nil {
		m.cancelFunc() // request stopping the current Run
		m.wg.Wait()    // wait for it to finish
	}

	newCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel // update with new cancel func

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.Run(newCtx); err != nil {
			m.log.Error("Error running manager after restart", zap.Error(err))
		}
	}()
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
	networkID, _, chainID, err := cli.Network(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch network details: %w", err)
	}

	// Reassign the newly created clients
	m.ncli = nrpc.NewJSONRPCClient(newNuklaiRPCUrl, networkID, chainID)

	// Reinitialize dependent properties
	m.epochStart = time.Now().Unix()
	m.feeAmount = m.config.MinFee
	m.t = timer.NewTimer(m.updateFee)

	m.log.Info("RPC client has been updated and manager reinitialized",
		zap.String("new RPC URL", newNuklaiRPCUrl),
		zap.Uint32("network ID", networkID),
		zap.String("chain ID", chainID.String()),
		zap.String("address", m.config.Recipient),
		zap.String("fee", utils.FormatBalance(m.feeAmount, nconsts.Decimals)),
	)

	// Restart the Run function safely
	m.RestartRun(ctx)

	return nil
}
