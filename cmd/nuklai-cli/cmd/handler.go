// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/pubsub"
	hrpc "github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

var _ cli.Controller = (*Controller)(nil)

type Handler struct {
	h *cli.Handler
}

func NewHandler(h *cli.Handler) *Handler {
	return &Handler{h}
}

func (h *Handler) Root() *cli.Handler {
	return h.h
}

func (*Handler) GetAssetInfo(
	ctx context.Context,
	ncli *nrpc.JSONRPCClient,
	addr codec.Address,
	assetID ids.ID,
	checkBalance bool,
) ([]byte, uint8, uint64, ids.ID, error) {
	var sourceChainID ids.ID
	exists, symbol, decimals, metadata, supply, _, warp, err := ncli.Asset(ctx, assetID, false)
	if err != nil {
		return nil, 0, 0, ids.Empty, err
	}
	if assetID != ids.Empty {
		if !exists {
			hutils.Outf("{{red}}%s does not exist{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil, 0, 0, ids.Empty, nil
		}
		if warp {
			sourceChainID = ids.ID(metadata[hconsts.IDLen:])
			sourceAssetID := ids.ID(metadata[:hconsts.IDLen])
			hutils.Outf(
				"{{yellow}}sourceChainID:{{/}} %s {{yellow}}sourceAssetID:{{/}} %s {{yellow}}supply:{{/}} %d\n",
				sourceChainID,
				sourceAssetID,
				supply,
			)
		} else {
			hutils.Outf(
				"{{yellow}}symbol:{{/}} %s {{yellow}}decimals:{{/}} %d {{yellow}}metadata:{{/}} %s {{yellow}}supply:{{/}} %d {{yellow}}warp:{{/}} %t\n",
				symbol,
				decimals,
				metadata,
				supply,
				warp,
			)
		}
	}
	if !checkBalance {
		return symbol, decimals, 0, sourceChainID, nil
	}
	saddr, err := codec.AddressBech32(nconsts.HRP, addr)
	if err != nil {
		return nil, 0, 0, ids.Empty, err
	}
	balance, err := ncli.Balance(ctx, saddr, assetID)
	if err != nil {
		return nil, 0, 0, ids.Empty, err
	}
	balanceToShow := fmt.Sprintf("%s %s", hutils.FormatBalance(balance, decimals), symbol)
	if assetID != ids.Empty {
		balanceToShow = fmt.Sprintf("%s %s", hutils.FormatBalance(balance, decimals), assetID)
	}
	if balance == 0 {
		hutils.Outf("{{red}}balance:{{/}} %s\n", balanceToShow)
		hutils.Outf("{{red}}please send funds to %s{{/}}\n", saddr)
		hutils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		hutils.Outf(
			"{{yellow}}balance:{{/}} %s\n",
			balanceToShow,
		)
	}
	return symbol, decimals, balance, sourceChainID, nil
}

func (h *Handler) DefaultActor() (
	ids.ID, *cli.PrivateKey, chain.AuthFactory,
	*hrpc.JSONRPCClient, *hrpc.WebSocketClient, *nrpc.JSONRPCClient, error,
) {
	addr, priv, err := h.h.GetDefaultKey(true)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}

	var factory chain.AuthFactory
	switch addr[0] {
	case nconsts.ED25519ID:
		factory = auth.NewED25519Factory(ed25519.PrivateKey(priv))
	case nconsts.SECP256R1ID:
		factory = auth.NewSECP256R1Factory(secp256r1.PrivateKey(priv))
	case nconsts.BLSID:
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
	// For [defaultActor], we always send requests to the first returned URI.
	hcli := hrpc.NewJSONRPCClient(uris[0])
	networkID, _, _, err := hcli.Network(context.TODO())
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	hws, err := hrpc.NewWebSocketClient(
		uris[0],
		hrpc.DefaultHandshakeTimeout,
		pubsub.MaxPendingMessages,
		pubsub.MaxReadMessageSize,
	)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	return chainID, &cli.PrivateKey{Address: addr, Bytes: priv}, factory, hcli, hws,
		nrpc.NewJSONRPCClient(
			uris[0],
			networkID,
			chainID,
		), nil
}

func (h *Handler) DefaultNuklaiVMJSONRPCClient(checkAllChains bool) ([]*nrpc.JSONRPCClient, error) {
	clients := make([]*nrpc.JSONRPCClient, 0)
	chainID, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return nil, err
	}
	max := len(uris)
	if !checkAllChains {
		max = 1
	}
	for _, uri := range uris[:max] {
		hcli := hrpc.NewJSONRPCClient(uris[0])
		networkID, _, _, err := hcli.Network(context.TODO())
		if err != nil {
			return nil, err
		}
		clients = append(clients, nrpc.NewJSONRPCClient(uri, networkID, chainID))
	}
	return clients, nil
}

func (*Handler) GetEmissionInfo(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
) (uint64, uint64, uint64, string, uint64, error) {
	totalSupply, maxSupply, rewardsPerBlock, emissionAccount, err := cli.EmissionInfo(ctx)
	if err != nil {
		return 0, 0, 0, "", 0, err
	}

	emissionAddress, err := codec.AddressBech32(nconsts.HRP, emissionAccount.Address)
	if err != nil {
		return 0, 0, 0, "", 0, err
	}

	hutils.Outf(
		"{{yellow}}emission info: {{/}}\nTotalSupply=%d MaxSupply=%d RewardsPerBlock=%d EmissionAddress=%s EmissionUnclaimedBalance=%d\n",
		totalSupply,
		maxSupply,
		rewardsPerBlock,
		emissionAddress,
		emissionAccount.UnclaimedBalance,
	)
	return totalSupply, maxSupply, rewardsPerBlock, emissionAddress, emissionAccount.UnclaimedBalance, err
}

func (*Handler) GetAllValidators(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
) ([]*emission.Validator, error) {
	validators, err := cli.Validators(ctx)
	if err != nil {
		return nil, err
	}
	for index, validator := range validators {
		hutils.Outf(
			"{{yellow}}validator %d:{{/}} NodeID=%s StakedAmount=%d DelegatedStake=%d DelegatedAmount=%d UnclaimedStakedReward=%d\n",
			index,
			validator.NodeID,
			validator.StakedAmount,
			validator.DelegatedStake,
			validator.DelegatedAmount,
			validator.UnclaimedStakedReward,
		)
	}
	return validators, nil
}

func (*Handler) GetValidatorStake(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
	nodeID ids.NodeID,
) (uint64, uint64, uint64, uint64, string, string, error) {
	stakeStartTime, stakeEndTime, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, err := cli.ValidatorStake(ctx, nodeID)
	if err != nil {
		return 0, 0, 0, 0, "", "", err
	}

	rewardAddressString, err := codec.AddressBech32(nconsts.HRP, rewardAddress)
	if err != nil {
		return 0, 0, 0, 0, "", "", err
	}
	ownerAddressString, err := codec.AddressBech32(nconsts.HRP, ownerAddress)
	if err != nil {
		return 0, 0, 0, 0, "", "", err
	}

	hutils.Outf(
		"{{yellow}}validator stake: {{/}}\nStakeStartTime=%d StakeEndTime=%d StakedAmount=%d DelegationFeeRate=%d RewardAddress=%s OwnerAddress=%s\n",
		stakeStartTime,
		stakeEndTime,
		stakedAmount,
		delegationFeeRate,
		rewardAddressString,
		ownerAddressString,
	)
	return stakeStartTime,
		stakeEndTime,
		stakedAmount,
		delegationFeeRate,
		rewardAddressString,
		ownerAddressString, err
}

func (*Handler) GetUserStake(ctx context.Context,
	cli *nrpc.JSONRPCClient, owner codec.Address, nodeID ids.NodeID,
) (uint64, uint64, string, string, error) {
	stakeStartTime, stakedAmount, rewardAddress, ownerAddress, err := cli.UserStake(ctx, owner, nodeID)
	if err != nil {
		return 0, 0, "", "", err
	}

	rewardAddressString, err := codec.AddressBech32(nconsts.HRP, rewardAddress)
	if err != nil {
		return 0, 0, "", "", err
	}
	ownerAddressString, err := codec.AddressBech32(nconsts.HRP, ownerAddress)
	if err != nil {
		return 0, 0, "", "", err
	}

	hutils.Outf(
		"{{yellow}}validator stake: {{/}}\nStakeStartTime=%d StakedAmount=%d RewardAddress=%s OwnerAddress=%s\n",
		stakeStartTime,
		stakedAmount,
		rewardAddressString,
		ownerAddressString,
	)
	return stakeStartTime,
		stakedAmount,
		rewardAddressString,
		ownerAddressString, err
}

var _ cli.Controller = (*Controller)(nil)

type Controller struct {
	databasePath string
}

func NewController(databasePath string) *Controller {
	return &Controller{databasePath}
}

func (c *Controller) DatabasePath() string {
	return c.databasePath
}

func (*Controller) Symbol() string {
	return nconsts.Symbol
}

func (*Controller) Decimals() uint8 {
	return nconsts.Decimals
}

func (*Controller) Address(addr codec.Address) string {
	return codec.MustAddressBech32(nconsts.HRP, addr)
}

func (*Controller) ParseAddress(address string) (codec.Address, error) {
	return codec.ParseAddressBech32(nconsts.HRP, address)
}
