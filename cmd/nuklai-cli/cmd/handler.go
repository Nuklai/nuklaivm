// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"encoding/base64"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/vm"

	"github.com/ava-labs/hypersdk/api/jsonrpc"
	"github.com/ava-labs/hypersdk/api/ws"
	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/utils"

	nutils "github.com/nuklai/nuklaivm/utils"
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

func (h *Handler) ImportChain(uri string) error {
	client := jsonrpc.NewJSONRPCClient(uri)
	_, _, chainID, err := client.Network(context.TODO())
	if err != nil {
		return err
	}
	if err := h.h.StoreChain(chainID, uri); err != nil {
		return err
	}
	if err := h.h.StoreDefaultChain(chainID); err != nil {
		return err
	}
	return nil
}

func (h *Handler) ImportCLI(cliPath string) error {
	oldChains, err := h.h.DeleteChains()
	if err != nil {
		return err
	}
	if len(oldChains) > 0 {
		utils.Outf("{{blue}}deleted old chains:{{/}} %+v\n", oldChains)
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
		utils.Outf(
			"{{blue}}[%s] stored chainID:{{/}} %s {{blue}}uri:{{/}} %s\n",
			name,
			chainID,
			uri,
		)
	}
	return h.h.StoreDefaultChain(chainID)
}

func (h *Handler) BalanceAsset(checkAllChains bool, isNFT bool, printBalance func(string, codec.Address, ids.ID, bool) error) error {
	addr, _, err := h.h.GetDefaultKey(true)
	if err != nil {
		return err
	}
	_, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return err
	}

	assetID, err := prompt.ID("assetID")
	if err != nil {
		return err
	}

	max := len(uris)
	if !checkAllChains {
		max = 1
	}
	for _, uri := range uris[:max] {
		utils.Outf("{{yellow}}uri:{{/}} %s\n", uri)
		if err := printBalance(uri, addr, assetID, isNFT); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) DefaultActor() (
	ids.ID, *cli.PrivateKey, chain.AuthFactory,
	*jsonrpc.JSONRPCClient, *vm.JSONRPCClient, *ws.WebSocketClient, error,
) {
	addr, priv, err := h.h.GetDefaultKey(true)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	var factory chain.AuthFactory
	switch addr[0] {
	case auth.ED25519ID:
		factory = auth.NewED25519Factory(ed25519.PrivateKey(priv))
	case auth.SECP256R1ID:
		factory = auth.NewSECP256R1Factory(secp256r1.PrivateKey(priv))
	case auth.BLSID:
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
	jcli := jsonrpc.NewJSONRPCClient(uris[0])
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	ws, err := ws.NewWebSocketClient(uris[0], ws.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
	if err != nil {
		return ids.Empty, nil, nil, nil, nil, nil, err
	}
	// For [defaultActor], we always send requests to the first returned URI.
	return chainID, &cli.PrivateKey{
			Address: addr,
			Bytes:   priv,
		}, factory, jcli,
		vm.NewJSONRPCClient(
			uris[0],
		), ws, nil
}

func (h *Handler) DefaultNuklaiVMJSONRPCClient(checkAllChains bool) ([]*vm.JSONRPCClient, error) {
	clients := make([]*vm.JSONRPCClient, 0)
	_, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return nil, err
	}
	max := len(uris)
	if !checkAllChains {
		max = 1
	}
	for _, uri := range uris[:max] {
		clients = append(clients, vm.NewJSONRPCClient(uri))
	}
	return clients, nil
}

func (*Handler) GetBalance(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	addr codec.Address,
	asset ids.ID,
) (uint64, error) {
	balance, err := cli.Balance(ctx, addr.String(), asset.String())
	if err != nil {
		return 0, err
	}
	if balance == 0 {
		utils.Outf("{{red}}balance:{{/}} 0 %s\n", consts.Symbol)
		utils.Outf("{{red}}please send funds to %s{{/}}\n", addr)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return 0, nil
	}
	utils.Outf(
		"{{yellow}}balance:{{/}} %s %s\n",
		nutils.FormatBalance(balance, consts.Decimals),
		consts.Symbol,
	)
	return balance, nil
}

func (*Handler) GetAssetInfo(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	addr codec.Address,
	assetID ids.ID,
	checkBalance bool,
) (uint64, string, string, string, uint8, string, uint64, uint64, string, string, string, string, string, error) {
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := cli.Asset(ctx, assetID.String(), false)
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	if assetID != ids.Empty {
		if !exists {
			utils.Outf("{{red}}%s does not exist{{/}}\n", assetID)
			utils.Outf("{{red}}exiting...{{/}}\n")
			return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", nil
		}
		utils.Outf(
			"{{blue}}assetType: {{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}owner:{{/}} %s {{blue}}mintAdmin:{{/}} %s {{blue}}pauseUnpauseAdmin:{{/}} %s {{blue}}freezeUnfreezeAdmin:{{/}} %s {{blue}}enableDisableKYCAccountAdmin:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			owner,
			mintAdmin,
			pauseUnpauseAdmin,
			freezeUnfreezeAdmin,
			enableDisableKYCAccountAdmin,
		)
	}
	if !checkBalance {
		return 0, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, nil
	}
	balance, err := cli.Balance(ctx, addr.String(), assetID.String())
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	if balance == 0 {
		utils.Outf("{{red}}assetID:{{/}} %s\n", assetID)
		utils.Outf("{{red}}name:{{/}} %s\n", name)
		utils.Outf("{{red}}symbol:{{/}} %s\n", symbol)
		utils.Outf("{{red}}balance:{{/}} 0\n")
		utils.Outf("{{red}}please send funds to %s{{/}}\n", addr.String())
		utils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		utils.Outf(
			"{{blue}}balance:{{/}} %s %s\n",
			nutils.FormatBalance(balance, decimals),
			symbol,
		)
	}
	return balance, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, nil
}

func (*Handler) GetAssetNFTInfo(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	addr codec.Address,
	nftID ids.ID,
	checkBalance bool,
) (bool, string, uint64, string, string, string, error) {
	exists, collectionID, uniqueID, uri, metadata, ownerAddress, err := cli.AssetNFT(ctx, nftID.String(), false)
	if err != nil {
		return false, "", 0, "", "", "", err
	}
	if !exists {
		utils.Outf("{{red}}%s does not exist{{/}}\n", nftID)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return false, "", 0, "", "", "", nil
	}
	if nftID == ids.Empty {
		utils.Outf("{{red}}%s is a native asset. Please pass in NFT ID{{/}}\n", nftID)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return false, "", 0, "", "", "", nil
	}

	if !checkBalance {
		return false, collectionID, uniqueID, uri, metadata, ownerAddress, nil
	}
	balance, err := cli.Balance(ctx, addr.String(), nftID.String())
	if err != nil {
		return false, "", 0, "", "", "", err
	}
	utils.Outf("{{blue}}collectionID:{{/}} %s\n", collectionID)
	utils.Outf("{{blue}}uniqueID:{{/}} %d\n", uniqueID)
	utils.Outf("{{blue}}uri:{{/}} %s\n", uri)
	utils.Outf("{{blue}}metadata:{{/}} %s\n", metadata)
	utils.Outf("{{blue}}ownerAddress:{{/}} %s\n", ownerAddress)
	if ownerAddress != addr.String() || balance == 0 {
		utils.Outf("{{red}}You do not own this NFT{{/}}\n")
		utils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		utils.Outf("{{blue}}You own this NFT{{/}}\n")
	}
	return true, collectionID, uniqueID, uri, metadata, ownerAddress, nil
}

func (*Handler) GetEmissionInfo(
	ctx context.Context,
	cli *vm.JSONRPCClient,
) (uint64, uint64, uint64, uint64, uint64, uint64, string, uint64, error) {
	currentBlockHeight, totalSupply, maxSupply, totalStaked, rewardsPerEpoch, emissionAccount, epochTracker, err := cli.EmissionInfo(ctx)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, "", 0, err
	}

	utils.Outf(
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
	cli *vm.JSONRPCClient,
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
		utils.Outf(
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
	cli *vm.JSONRPCClient,
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
		utils.Outf(
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
	cli *vm.JSONRPCClient,
	nodeID ids.NodeID,
) (uint64, uint64, uint64, uint64, string, string, error) {
	stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, err := cli.ValidatorStake(ctx, nodeID)
	if err != nil {
		return 0, 0, 0, 0, "", "", err
	}

	utils.Outf(
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
	cli *vm.JSONRPCClient, owner codec.Address, nodeID ids.NodeID,
) (uint64, uint64, uint64, string, string, error) {
	stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, err := cli.UserStake(ctx, owner.String(), nodeID.String())
	if err != nil {
		return 0, 0, 0, "", "", err
	}
	utils.Outf(
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
	cli *vm.JSONRPCClient,
	datasetID ids.ID,
) (string, string, string, string, string, string, string, bool, string, string, uint64, uint8, uint8, uint8, uint8, string, error) {
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := cli.Dataset(ctx, datasetID.String(), false)
	if err != nil {
		return "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", err
	}
	if !exists {
		utils.Outf("{{red}}%s does not exist{{/}}\n", datasetID)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", nil
	}

	utils.Outf(
		"{{blue}}dataset info: {{/}}\nName=%s Description=%s Categories=%s LicenseName=%s LicenseSymbol=%s LicenseURL=%s Metadata=%s IsCommunityDataset=%t SaleID=%s BaseAsset=%s BasePrice=%d RevenueModelDataShare=%d RevenueModelMetadataShare=%d RevenueModelDataOwnerCut=%d RevenueModelMetadataOwnerCut=%d Owner=%s\n",
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
	cli *vm.JSONRPCClient,
	datasetID ids.ID,
) ([]vm.DataContribution, error) {
	contributions, err := cli.DataContributionPending(ctx, datasetID.String())
	if err != nil {
		return nil, err
	}
	for index, contribution := range contributions {
		utils.Outf(
			"{{blue}}Contribution %d:{{/}} Contributor=%s DataLocation=%s DataIdentifier=%s\n",
			index,
			contribution.Contributor,
			contribution.DataLocation,
			contribution.DataIdentifier,
		)
	}
	return contributions, nil
}

func (*Handler) GetDatasetInfoFromMarketplace(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	datasetID ids.ID,
) (string, string, bool, string, string, uint64, string, string, string, string, string, uint64, uint64, string, map[string]string, error) {
	exists, datasetName, description, _, _, _, _, _, isCommunityDataset, saleID, baseAsset, basePrice, _, _, _, _, owner, err := cli.Dataset(ctx, datasetID.String(), false)
	if !exists {
		utils.Outf("{{red}}Dataset '%s' does not exist{{/}}\n", datasetID)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}
	if saleID == ids.Empty.String() {
		utils.Outf("{{red}}Dataset '%s' is not on sale{{/}}\n", datasetID)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}

	_, assetType, assetName, symbol, _, metadata, uri, totalSupply, maxSupply, admin, _, _, _, _, err := cli.Asset(ctx, saleID, false)
	if err != nil {
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}

	metadataMap, err := nutils.BytesToMap([]byte(metadata))
	if err != nil {
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}
	utils.Outf(
		"{{blue}}dataset info from marketplace: {{/}}\nDatasetName=%s DatasetDescription=%s IsCommunityDataset=%t MarketplaceAssetID=%s AssetForPayment=%s PricePerBlock=%d DatasetOwner=%s\n{{blue}}marketplace asset info: {{/}}\nAssetType=%s AssetName=%s AssetSymbol=%s AssetURI=%s TotalSupply=%d MaxSupply=%d Owner=%s\nAssetMetadata=%#v\n",
		datasetName,
		description,
		isCommunityDataset,
		saleID,
		baseAsset,
		basePrice,
		owner,
		assetType,
		assetName,
		symbol,
		uri,
		totalSupply,
		maxSupply,
		admin,
		metadataMap,
	)
	return datasetName,
		description,
		isCommunityDataset,
		saleID,
		baseAsset,
		basePrice,
		owner,
		assetType,
		assetName,
		symbol,
		uri,
		totalSupply,
		maxSupply,
		admin,
		metadataMap,
		err
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
	return consts.Symbol
}

func (*Controller) Decimals() uint8 {
	return consts.Decimals
}

func (*Controller) GetParser(uri string) (chain.Parser, error) {
	cli := vm.NewJSONRPCClient(uri)
	return cli.Parser(context.TODO())
}

func (*Controller) HandleTx(tx *chain.Transaction, result *chain.Result) {
	handleTx(tx, result)
}

func (*Controller) LookupBalance(address codec.Address, uri string) (uint64, error) {
	cli := vm.NewJSONRPCClient(uri)
	balance, err := cli.Balance(context.TODO(), address.String(), ids.Empty.String())
	return balance, err
}
