// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/auth"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	brpc "github.com/nuklai/nuklaivm/rpc"
)

type Handler struct {
	h *cli.Handler
}

func NewHandler(h *cli.Handler) *Handler {
	return &Handler{h}
}

func (h *Handler) Root() *cli.Handler {
	return h.h
}

func (h *Handler) DefaultActor() (
	ids.ID, *cli.PrivateKey, chain.AuthFactory,
	*rpc.JSONRPCClient, *brpc.JSONRPCClient, *rpc.WebSocketClient, error,
) {
	addr, priv, err := h.h.GetDefaultKey(true)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	var factory chain.AuthFactory
	switch addr[0] {
	case consts.ED25519ID:
		factory = auth.NewED25519Factory(ed25519.PrivateKey(priv))
	case consts.SECP256R1ID:
		factory = auth.NewSECP256R1Factory(secp256r1.PrivateKey(priv))
	case consts.BLSID:
		p, err := bls.PrivateKeyFromBytes(priv)
		if err != nil {
			return ids.Empty, nil, nil, nil, nil, nil, err
		}
		factory = auth.NewBLSFactory(p)
	default:
		return ids.Empty, nil, nil, nil, nil, nil, ErrInvalidAddress
	}
	chainID, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	jcli := rpc.NewJSONRPCClient(uris[0])
	networkID, _, _, err := jcli.Network(context.TODO())
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	ws, err := rpc.NewWebSocketClient(uris[0], rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	// For [defaultActor], we always send requests to the first returned URI.
	return chainID, &cli.PrivateKey{
			Address: addr,
			Bytes:   priv,
		}, factory, jcli,
		brpc.NewJSONRPCClient(
			uris[0],
			networkID,
			chainID,
		), ws, nil
}

func (h *Handler) DefaultNuklaiVMJSONRPCClient(checkAllChains bool) ([]*brpc.JSONRPCClient, error) {
	clients := make([]*brpc.JSONRPCClient, 0)
	chainID, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return nil, err
	}
	max := len(uris)
	if !checkAllChains {
		max = 1
	}
	for _, uri := range uris[:max] {
		rcli := rpc.NewJSONRPCClient(uris[0])
		networkID, _, _, err := rcli.Network(context.TODO())
		if err != nil {
			return nil, err
		}
		clients = append(clients, brpc.NewJSONRPCClient(uri, networkID, chainID))
	}
	return clients, nil
}

func (*Handler) GetBalance(
	ctx context.Context,
	cli *brpc.JSONRPCClient,
	addr codec.Address,
) (uint64, error) {
	saddr, err := codec.AddressBech32(consts.HRP, addr)
	if err != nil {
		return 0, err
	}
	balance, err := cli.Balance(ctx, saddr)
	if err != nil {
		return 0, err
	}
	if balance == 0 {
		utils.Outf("{{red}}balance:{{/}} 0 %s\n", consts.Symbol)
		utils.Outf("{{red}}please send funds to %s{{/}}\n", saddr)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return 0, nil
	}
	utils.Outf(
		"{{yellow}}balance:{{/}} %s %s\n",
		utils.FormatBalance(balance, consts.Decimals),
		consts.Symbol,
	)
	return balance, nil
}

func (*Handler) GetEmissionInfo(
	ctx context.Context,
	cli *brpc.JSONRPCClient,
) (uint64, uint64, uint64, error) {
	totalSupply, maxSupply, rewardsPerBlock, err := cli.EmissionInfo(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	utils.Outf(
		"{{yellow}}emission info: {{/}}\nTotalSupply=%d MaxSupply=%d RewardsPerBlock=%d\n",
		totalSupply,
		maxSupply,
		rewardsPerBlock,
	)
	return totalSupply, maxSupply, rewardsPerBlock, err
}

func (*Handler) GetAllValidators(
	ctx context.Context,
	cli *brpc.JSONRPCClient,
) ([]*emission.Validator, error) {
	validators, err := cli.Validators(ctx)
	if err != nil {
		return nil, err
	}
	for index, validator := range validators {
		utils.Outf(
			"{{yellow}}validator %d:{{/}} NodeID=%s NodePublicKey=%s UserStake=%v StakedAmount=%d StakedReward=%d\n",
			index,
			validator.NodeID,
			validator.NodePublicKey,
			validator.UserStake,
			validator.StakedAmount,
			validator.StakedReward,
		)
	}
	return validators, nil
}

func (*Handler) GetUserStake(ctx context.Context,
	cli *brpc.JSONRPCClient, nodeID ids.NodeID, owner codec.Address,
) (*emission.UserStake, error) {
	saddr, err := codec.AddressBech32(consts.HRP, owner)
	if err != nil {
		return nil, err
	}
	userStake, err := cli.UserStakeInfo(ctx, nodeID, saddr)
	if err != nil {
		return nil, err
	}

	if userStake.Owner == "" {
		utils.Outf("{{yellow}}user stake: {{/}} Not staked yet\n")
	} else {
		utils.Outf(
			"{{yellow}}user stake: {{/}} Owner=%s StakedAmount=%d\n",
			userStake.Owner,
			userStake.StakedAmount,
		)
	}

	index := 1
	for txID, stakeInfo := range userStake.StakeInfo {
		utils.Outf(
			"{{yellow}}stake #%d:{{/}} TxID=%s Amount=%d StartLockUp=%d\n",
			index,
			txID,
			stakeInfo.Amount,
			stakeInfo.StartLockUp,
		)
		index++
	}
	return userStake, err
}

type Controller struct {
	databasePath string
}

var _ cli.Controller = (*Controller)(nil)

func NewController(databasePath string) *Controller {
	return &Controller{databasePath}
}

func (c *Controller) DatabasePath() string {
	return c.databasePath
}

func (*Controller) Symbol() string {
	return consts.Symbol
}

func (*Controller) Decimals() uint8 {
	return consts.Decimals
}

func (*Controller) Address(addr codec.Address) string {
	return codec.MustAddressBech32(consts.HRP, addr)
}

func (*Controller) ParseAddress(addr string) (codec.Address, error) {
	return codec.ParseAddressBech32(consts.HRP, addr)
}
