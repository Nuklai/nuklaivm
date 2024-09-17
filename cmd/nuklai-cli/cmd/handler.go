// Copyright (C) 2024, Nuklai. All rights reserved.
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
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
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
		hutils.Outf("{{blue}}deleted old chains:{{/}} %+v\n", oldChains)
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
			"{{blue}}[%s] stored chainID:{{/}} %s {{blue}}uri:{{/}} %s\n",
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
) (uint64, string, string, string, uint8, string, uint64, uint64, string, string, string, string, string, error) {
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := cli.Asset(ctx, assetID.String(), false)
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	if assetID != ids.Empty {
		if !exists {
			hutils.Outf("{{red}}%s does not exist{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", nil
		}
		hutils.Outf(
			"{{blue}}assetType: {{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			admin,
			mintActor,
			pauseUnpauseActor,
			freezeUnfreezeActor,
			enableDisableKYCAccountActor,
		)
	}
	if !checkBalance {
		return 0, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, nil
	}
	saddr, err := codec.AddressBech32(nconsts.HRP, addr)
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	balance, err := cli.Balance(ctx, saddr, assetID.String())
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	if balance == 0 {
		hutils.Outf("{{red}}assetID:{{/}} %s\n", assetID)
		hutils.Outf("{{red}}name:{{/}} %s\n", name)
		hutils.Outf("{{red}}symbol:{{/}} %s\n", symbol)
		hutils.Outf("{{red}}balance:{{/}} 0\n")
		hutils.Outf("{{red}}please send funds to %s{{/}}\n", saddr)
		hutils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		hutils.Outf(
			"{{blue}}balance:{{/}} %s %s\n",
			hutils.FormatBalance(balance, decimals),
			symbol,
		)
	}
	return balance, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, nil
}

func (*Handler) GetAssetNFTInfo(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
	addr codec.Address,
	nftID ids.ID,
	checkBalance bool,
) (bool, string, uint64, string, string, string, error) {
	exists, collectionID, uniqueID, uri, metadata, ownerAddress, err := cli.AssetNFT(ctx, nftID.String(), false)
	if err != nil {
		return false, "", 0, "", "", "", err
	}
	if !exists {
		hutils.Outf("{{red}}%s does not exist{{/}}\n", nftID)
		hutils.Outf("{{red}}exiting...{{/}}\n")
		return false, "", 0, "", "", "", nil
	}
	if nftID == ids.Empty {
		hutils.Outf("{{red}}%s is a native asset. Please pass in NFT ID{{/}}\n", nftID)
		hutils.Outf("{{red}}exiting...{{/}}\n")
		return false, "", 0, "", "", "", nil
	}

	if !checkBalance {
		return false, collectionID, uniqueID, uri, metadata, ownerAddress, nil
	}
	saddr, err := codec.AddressBech32(nconsts.HRP, addr)
	if err != nil {
		return false, "", 0, "", "", "", err
	}
	balance, err := cli.Balance(ctx, saddr, nftID.String())
	if err != nil {
		return false, "", 0, "", "", "", err
	}
	hutils.Outf("{{blue}}collectionID:{{/}} %s\n", collectionID)
	hutils.Outf("{{blue}}uniqueID:{{/}} %d\n", uniqueID)
	hutils.Outf("{{blue}}uri:{{/}} %s\n", uri)
	hutils.Outf("{{blue}}metadata:{{/}} %s\n", metadata)
	hutils.Outf("{{blue}}ownerAddress:{{/}} %s\n", ownerAddress)
	if address := codec.MustAddressBech32(nconsts.HRP, addr); ownerAddress != address || balance == 0 {
		hutils.Outf("{{red}}You do not own this NFT{{/}}\n")
		hutils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		hutils.Outf("{{blue}}You own this NFT{{/}}\n")
	}
	return true, collectionID, uniqueID, uri, metadata, ownerAddress, nil
}

func (h *Handler) DefaultActor() (
	ids.ID, *cli.PrivateKey, chain.AuthFactory,
	*rpc.JSONRPCClient, *rpc.WebSocketClient, *nrpc.JSONRPCClient, error,
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
	hcli := rpc.NewJSONRPCClient(uris[0])
	networkID, _, _, err := hcli.Network(context.TODO())
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	hws, err := rpc.NewWebSocketClient(
		uris[0],
		rpc.DefaultHandshakeTimeout,
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
		hcli := rpc.NewJSONRPCClient(uris[0])
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

	hutils.Outf(
		"{{blue}}emission info: {{/}}\nCurrentBlockHeight=%d TotalSupply=%d MaxSupply=%d TotalStaked=%d RewardsPerEpoch=%d NumBlocksInEpoch=%d EmissionAddress=%s EmissionAccumulatedReward=%d\n",
		currentBlockHeight,
		totalSupply,
		maxSupply,
		totalStaked,
		rewardsPerEpoch,
		epochTracker.EpochLength,
		emissionAccount.Address,
		emissionAccount.AccumulatedReward,
	)
	return currentBlockHeight, totalSupply, maxSupply, totalStaked, rewardsPerEpoch, epochTracker.EpochLength, emissionAccount.Address, emissionAccount.AccumulatedReward, err
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
			"{{blue}}validator %d:{{/}} NodeID=%s PublicKey=%s StakedAmount=%d AccumulatedStakedReward=%d DelegationFeeRate=%f DelegatedAmount=%d AccumulatedDelegatedReward=%d\n",
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
			"{{blue}}validator %d:{{/}} NodeID=%s PublicKey=%s Active=%t StakedAmount=%d AccumulatedStakedReward=%d DelegationFeeRate=%f DelegatedAmount=%d AccumulatedDelegatedReward=%d\n",
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

	hutils.Outf(
		"{{blue}}validator stake: {{/}}\nStakeStartBlock=%d StakeEndBlock=%d StakedAmount=%d DelegationFeeRate=%d RewardAddress=%s OwnerAddress=%s\n",
		stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		delegationFeeRate,
		rewardAddress,
		ownerAddress,
	)
	return stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		delegationFeeRate,
		rewardAddress,
		ownerAddress,
		err
}

func (*Handler) GetUserStake(ctx context.Context,
	cli *nrpc.JSONRPCClient, owner codec.Address, nodeID ids.NodeID,
) (uint64, uint64, uint64, string, string, error) {
	stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, err := cli.UserStake(ctx, codec.MustAddressBech32(nconsts.HRP, owner), nodeID.String())
	if err != nil {
		return 0, 0, 0, "", "", err
	}
	hutils.Outf(
		"{{blue}}user stake: {{/}}\nStakeStartBlock=%d StakeEndBlock=%d StakedAmount=%d RewardAddress=%s OwnerAddress=%s\n",
		stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		rewardAddress,
		ownerAddress,
	)
	return stakeStartBlock,
		stakeEndBlock,
		stakedAmount,
		rewardAddress,
		ownerAddress,
		err
}

func (*Handler) GetDatasetInfo(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
	datasetID ids.ID,
) (string, string, string, string, string, string, string, bool, string, string, uint64, uint8, uint8, uint8, uint8, string, error) {
	_, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := cli.Dataset(ctx, datasetID.String(), false)
	if err != nil {
		return "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", err
	}

	hutils.Outf(
		"{{blue}}dataset info: {{/}}\nName=%s Description=%s Categories=%s LicenseName=%s LicenseSymbol=%s LicenseURL=%s Metadata=%s IsCommunityDataset=%t OnSale=%t BaseAsset=%s BasePrice=%d RevenueModelDataShare=%d RevenueModelMetadataShare=%d RevenueModelDataOwnerCut=%d RevenueModelMetadataOwnerCut=%d Owner=%s\n",
		name,
		description,
		categories,
		licenseName,
		licenseSymbol,
		licenseURL,
		metadata,
		isCommunityDataset,
		saleID,
		baseAsset,
		basePrice,
		revenueModelDataShare,
		revenueModelMetadataShare,
		revenueModelDataOwnerCut,
		revenueModelMetadataOwnerCut,
		owner,
	)
	return name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err
}

func (*Handler) GetDataContributionPendingInfo(
	ctx context.Context,
	cli *nrpc.JSONRPCClient,
	datasetID ids.ID,
) ([]nrpc.DataContribution, error) {
	contributions, err := cli.DataContributionPending(ctx, datasetID.String())
	if err != nil {
		return nil, err
	}
	for index, contribution := range contributions {
		hutils.Outf(
			"{{blue}}Contribution %d:{{/}} Contributor=%s DataLocation=%s DataIdentifier=%s\n",
			index,
			contribution.Contributor,
			contribution.DataLocation,
			contribution.DataIdentifier,
		)
	}
	return contributions, nil
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
