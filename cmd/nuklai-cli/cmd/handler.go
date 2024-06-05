// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"encoding/base64"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
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

func (h *Handler) ImportCLI(cliPath string) error {
	oldChains, err := h.h.DeleteChains()
	if err != nil {
		return err
	}
	if len(oldChains) > 0 {
		hutils.Outf("{{yellow}}deleted old chains:{{/}} %+v\n", oldChains)
	}

	// Load yaml file
	chainID, nodes, err := ReadCLIFile(cliPath)
	if err != nil {
		return err
	}
	for name, uri := range nodes {
		if err := h.h.StoreChain(chainID, uri); err != nil {
			return err
		}
		hutils.Outf(
			"{{yellow}}[%s] stored chainID:{{/}} %s {{yellow}}uri:{{/}} %s\n",
			name,
			chainID,
			uri,
		)
	}
	return h.h.StoreDefaultChain(chainID)
}

func (*Handler) GetAssetInfo(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
	addr codec.Address,
	assetID ids.ID,
	checkBalance bool,
) ([]byte, uint8, uint64, ids.ID, error) {
	var sourceChainID ids.ID
	exists, symbol, decimals, metadata, supply, _, err := cli.Asset(ctx, assetID, false)
	if err != nil {
		return nil, 0, 0, ids.Empty, err
	}
	if assetID != ids.Empty {
		if !exists {
			hutils.Outf("{{red}}%s does not exist{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil, 0, 0, ids.Empty, nil
		}
		hutils.Outf(
			"{{yellow}}symbol:{{/}} %s {{yellow}}decimals:{{/}} %d {{yellow}}metadata:{{/}} %s {{yellow}}supply:{{/}} %d\n",
			symbol,
			decimals,
			metadata,
			supply,
		)
	}
	if !checkBalance {
		return symbol, decimals, 0, sourceChainID, nil
	}
	saddr, err := codec.AddressBech32(nconsts.HRP, addr)
	if err != nil {
		return nil, 0, 0, ids.Empty, err
	}
	balance, err := cli.Balance(ctx, saddr, assetID)
	if err != nil {
		return nil, 0, 0, ids.Empty, err
	}
	if balance == 0 {
		hutils.Outf("{{red}}balance:{{/}} 0 %s\n", assetID)
		hutils.Outf("{{red}}please send funds to %s{{/}}\n", saddr)
		hutils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		hutils.Outf(
			"{{yellow}}balance:{{/}} %s %s\n",
			hutils.FormatBalance(balance, decimals),
			symbol,
		)
	}
	return symbol, decimals, balance, sourceChainID, nil
}

func (h *Handler) DefaultActor() (
	ids.ID, *cli.PrivateKey, chain.AuthFactory,
	*rpc.JSONRPCClient, *rpc.WebSocketClient, *nrpc.JSONRPCClient, error,
) {
	addr, priv, err := h.h.GetDefaultKey(true)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	chainID, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	// For [defaultActor], we always send requests to the first returned URI.
	jcli := rpc.NewJSONRPCClient(uris[0])
	networkID, _, _, err := jcli.Network(context.TODO())
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	scli, err := rpc.NewWebSocketClient(
		uris[0],
		rpc.DefaultHandshakeTimeout,
		pubsub.MaxPendingMessages,
		pubsub.MaxReadMessageSize,
	)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	return chainID, &cli.PrivateKey{Address: addr, Bytes: priv}, auth.NewED25519Factory(ed25519.PrivateKey(priv)), jcli, scli,
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
) (uint64, uint64, uint64, uint64, uint64, uint64, string, uint64, error) {
	currentBlockHeight, totalSupply, maxSupply, totalStaked, rewardsPerEpoch, emissionAccount, epochTracker, err := cli.EmissionInfo(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, "", 0, err
	}

	emissionAddress, err := codec.AddressBech32(nconsts.HRP, emissionAccount.Address)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, "", 0, err
	}

	hutils.Outf(
		"{{yellow}}emission info: {{/}}\nCurrentBlockHeight=%d TotalSupply=%d MaxSupply=%d TotalStaked=%d RewardsPerEpoch=%d NumBlocksInEpoch=%d EmissionAddress=%s EmissionAccumulatedReward=%d\n",
		currentBlockHeight,
		totalSupply,
		maxSupply,
		totalStaked,
		rewardsPerEpoch,
		epochTracker.EpochLength,
		emissionAddress,
		emissionAccount.AccumulatedReward,
	)
	return currentBlockHeight, totalSupply, maxSupply, totalStaked, rewardsPerEpoch, epochTracker.EpochLength, emissionAddress, emissionAccount.AccumulatedReward, err
}

func (*Handler) GetAllValidators(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
) ([]*emission.Validator, error) {
	validators, err := cli.AllValidators(ctx)
	if err != nil {
		return nil, err
	}
	for index, validator := range validators {
		publicKey, err := bls.PublicKeyFromBytes(validator.PublicKey)
		if err != nil {
			return nil, err
		}
		hutils.Outf(
			"{{yellow}}validator %d:{{/}} NodeID=%s PublicKey=%s StakedAmount=%d AccumulatedStakedReward=%d DelegationFeeRate=%f DelegatedAmount=%d AccumulatedDelegatedReward=%d\n",
			index,
			validator.NodeID,
			base64.StdEncoding.EncodeToString(publicKey.Compress()),
			validator.StakedAmount,
			validator.AccumulatedStakedReward,
			validator.DelegationFeeRate,
			validator.DelegatedAmount,
			validator.AccumulatedDelegatedReward,
		)
	}
	return validators, nil
}

func (*Handler) GetStakedValidators(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
) ([]*emission.Validator, error) {
	validators, err := cli.StakedValidators(ctx)
	if err != nil {
		return nil, err
	}
	for index, validator := range validators {
		publicKey, err := bls.PublicKeyFromBytes(validator.PublicKey)
		if err != nil {
			return nil, err
		}
		hutils.Outf(
			"{{yellow}}validator %d:{{/}} NodeID=%s PublicKey=%s Active=%t StakedAmount=%d AccumulatedStakedReward=%d DelegationFeeRate=%f DelegatedAmount=%d AccumulatedDelegatedReward=%d\n",
			index,
			validator.NodeID,
			base64.StdEncoding.EncodeToString(publicKey.Compress()),
			validator.IsActive,
			validator.StakedAmount,
			validator.AccumulatedStakedReward,
			validator.DelegationFeeRate,
			validator.DelegatedAmount,
			validator.AccumulatedDelegatedReward,
		)
	}
	return validators, nil
}

func (*Handler) GetValidatorStake(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
	nodeID ids.NodeID,
) (uint64, uint64, uint64, uint64, string, string, error) {
	stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, err := cli.ValidatorStake(ctx, nodeID)
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
		"{{yellow}}validator stake: {{/}}\nStakeStartBlock=%d StakeEndBlock=%d StakedAmount=%d DelegationFeeRate=%d RewardAddress=%s OwnerAddress=%s\n",
		stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		delegationFeeRate,
		rewardAddressString,
		ownerAddressString,
	)
	return stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		delegationFeeRate,
		rewardAddressString,
		ownerAddressString, err
}

func (*Handler) GetUserStake(ctx context.Context,
	cli *nrpc.JSONRPCClient, owner codec.Address, nodeID ids.NodeID,
) (uint64, uint64, uint64, string, string, error) {
	stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, err := cli.UserStake(ctx, owner, nodeID)
	if err != nil {
		return 0, 0, 0, "", "", err
	}

	rewardAddressString, err := codec.AddressBech32(nconsts.HRP, rewardAddress)
	if err != nil {
		return 0, 0, 0, "", "", err
	}
	ownerAddressString, err := codec.AddressBech32(nconsts.HRP, ownerAddress)
	if err != nil {
		return 0, 0, 0, "", "", err
	}

	hutils.Outf(
		"{{yellow}}user stake: {{/}}\nStakeStartBlock=%d StakeEndBlock=%d StakedAmount=%d RewardAddress=%s OwnerAddress=%s\n",
		stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		rewardAddressString,
		ownerAddressString,
	)
	return stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		rewardAddressString,
		ownerAddressString, err
}

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
