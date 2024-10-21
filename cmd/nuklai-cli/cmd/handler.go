// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
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

func (h *Handler) SetKey() error {
	keys, err := h.h.GetKeys()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		utils.Outf("{{red}}no stored keys{{/}}\n")
		return nil
	}
	_, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return err
	}
	if len(uris) == 0 {
		utils.Outf("{{red}}no available chains{{/}}\n")
		return nil
	}
	utils.Outf("{{cyan}}stored keys:{{/}} %d\n", len(keys))
	for i := 0; i < len(keys); i++ {
		addrStr := keys[i].Address
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		for _, ncli := range nclients {
			if _, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(context.TODO(), ncli, addrStr, storage.NAIAddress, true, false, i); err != nil {
				return err
			}
		}
	}

	// Select key
	keyIndex, err := prompt.Choice("set default key", len(keys))
	if err != nil {
		return err
	}
	key := keys[keyIndex]
	return h.h.StoreDefaultKey(key.Address)
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

func (h *Handler) BalanceAsset(checkAllChains bool, isNFT bool, printBalance func(string, codec.Address, codec.Address, bool) error) error {
	addr, _, err := h.h.GetDefaultKey(true)
	if err != nil {
		return err
	}
	_, uris, err := h.h.GetDefaultChain(true)
	if err != nil {
		return err
	}

	assetAddress, err := prompt.Address("assetAddress")
	if err != nil {
		return err
	}

	max := len(uris)
	if !checkAllChains {
		max = 1
	}
	for _, uri := range uris[:max] {
		utils.Outf("{{yellow}}uri:{{/}} %s\n", uri)
		if err := printBalance(uri, addr, assetAddress, isNFT); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) DefaultActor() (
	ids.ID, *auth.PrivateKey, chain.AuthFactory,
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
	return chainID, &auth.PrivateKey{
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
	actor codec.Address,
	assetAddress codec.Address,
	checkBalance bool,
	printOutput bool,
	index int,
) (uint64, string, string, string, uint8, string, uint64, uint64, string, string, string, string, string, error) {
	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := cli.Asset(ctx, assetAddress.String(), false)
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	if printOutput {
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
	balance, err := cli.Balance(ctx, actor.String(), assetAddress.String())
	if err != nil {
		return 0, "", "", "", 0, "", 0, 0, "", "", "", "", "", err
	}
	output := ""
	if index >= 0 {
		output += fmt.Sprintf("%d) ", index)
	}
	output += fmt.Sprintf("{{cyan}}address:{{/}} %s {{cyan}}balance:{{/}} %s %s\n", actor, nutils.FormatBalance(balance, decimals), symbol)
	utils.Outf(output)

	return balance, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, nil
}

func (*Handler) GetAssetNFTInfo(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	actor codec.Address,
	nftAddress codec.Address,
	checkBalance bool,
) (uint64, string, string, string, string, string, string, error) {
	assetType, name, symbol, decimals, metadata, uri, _, _, owner, _, _, _, _, err := cli.Asset(ctx, nftAddress.String(), false)
	if err != nil {
		return 0, "", "", "", "", "", "", err
	}
	utils.Outf(
		"{{blue}}assetType: {{/}} %s {{blue}}name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}metadata:{{/}} %s {{blue}}collectionAssetAddress:{{/}} %s {{blue}}owner:{{/}} %s\n",
		assetType,
		name,
		symbol,
		metadata,
		uri,
		owner,
	)

	if !checkBalance {
		return 0, assetType, name, symbol, metadata, uri, owner, nil
	}
	balance, err := cli.Balance(ctx, actor.String(), uri)
	if err != nil {
		return 0, "", "", "", "", "", "", err
	}
	if owner != actor.String() {
		utils.Outf("{{red}}You do not own this NFT{{/}}\n")
		utils.Outf("{{red}}exiting...{{/}}\n")
	} else {
		utils.Outf(
			"{{blue}}collectionAssetAddress:{{/}} %s {{blue}}balance:{{/}} %s %s\n",
			uri,
			nutils.FormatBalance(balance, decimals),
			symbol,
		)
		utils.Outf("{{blue}}You own this NFT{{/}}\n")
	}
	return balance, assetType, name, symbol, metadata, uri, owner, nil
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
	datasetAddress codec.Address,
) (string, string, string, string, string, string, string, bool, string, string, uint64, uint8, uint8, uint8, uint8, string, error) {
	name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := cli.Dataset(ctx, datasetAddress.String(), false)
	if err != nil {
		return "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", err
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

func (*Handler) GetDataContributionInfo(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	contributionID ids.ID,
) (string, string, string, string, bool, error) {
	datasetAddress, dataLocation, dataIdentifier, contributor, contributionAcceptedByDatasetOwner, err := cli.DatasetContribution(ctx, contributionID.String())
	if err != nil {
		return "", "", "", "", false, err
	}
	utils.Outf(
		"{{blue}}contribution info: {{/}}\nDatasetAddress=%s DataLocation=%s DataIdentifier=%s Contributor=%s ContributionAcceptedByDatasetOwner=%t\n",
		datasetAddress,
		dataLocation,
		dataIdentifier,
		contributor,
		contributionAcceptedByDatasetOwner,
	)
	return datasetAddress, dataLocation, dataIdentifier, contributor, contributionAcceptedByDatasetOwner, nil
}

func (*Handler) GetDatasetInfoFromMarketplace(
	ctx context.Context,
	cli *vm.JSONRPCClient,
	datasetAddress codec.Address,
) (string, string, bool, string, string, uint64, string, string, string, string, string, uint64, uint64, string, map[string]string, error) {
	datasetName, description, _, _, _, _, _, isCommunityDataset, marketplaceAssetAddress, paymentAssetAddress, datasetPricePerBlock, _, _, _, _, owner, err := cli.Dataset(ctx, datasetAddress.String(), false)
	if marketplaceAssetAddress == codec.EmptyAddress.String() {
		utils.Outf("{{red}}Dataset '%s' is not on sale{{/}}\n", datasetAddress)
		utils.Outf("{{red}}exiting...{{/}}\n")
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}

	assetType, assetName, symbol, _, metadata, uri, totalSupply, maxSupply, admin, _, _, _, _, err := cli.Asset(ctx, marketplaceAssetAddress, false)
	if err != nil {
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}

	metadataMap, err := nutils.BytesToMap([]byte(metadata))
	if err != nil {
		return "", "", false, "", "", 0, "", "", "", "", "", 0, 0, "", nil, err
	}
	utils.Outf(
		"{{blue}}marketplace dataset info: {{/}}\nDatasetName=%s DatasetDescription=%s IsCommunityDataset=%t MarketplaceAssetAddress=%s PaymentAssetAddress=%s DatasetPricePerBlock=%d DatasetOwner=%s\n{{blue}}\nmarketplace asset info: {{/}}\nAssetType=%s AssetName=%s AssetSymbol=%s AssetURI=%s TotalSupply=%d MaxSupply=%d Owner=%s\nAssetMetadata=%#v\n",
		datasetName,
		description,
		isCommunityDataset,
		marketplaceAssetAddress,
		paymentAssetAddress,
		datasetPricePerBlock,
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
		marketplaceAssetAddress,
		paymentAssetAddress,
		datasetPricePerBlock,
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
