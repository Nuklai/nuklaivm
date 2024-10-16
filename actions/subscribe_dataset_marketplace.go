// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/dataset"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	SubscribeDatasetMarketplaceComputeUnits = 5
)

var (
	ErrDatasetNotOnSale                               = errors.New("dataset is not on sale")
	ErrMarketplaceAssetAddressInvalid                 = errors.New("marketplace asset ID is invalid")
	ErrPaymentAssetNotSupported                       = errors.New("base asset is not supported")
	ErrOutputNumBlocksToSubscribeInvalid              = errors.New("num blocks to subscribe is invalid")
	ErrUserAlreadySubscribed                          = errors.New("user is already subscribed")
	_                                    chain.Action = (*SubscribeDatasetMarketplace)(nil)
)

type SubscribeDatasetMarketplace struct {
	// Marketplace asset address that represents the dataset subscription in the
	// marketplace
	MarketplaceAssetAddress codec.Address `serialize:"true" json:"marketplace_asset_address"`

	// Asset to use for the subscription
	PaymentAssetAddress codec.Address `serialize:"true" json:"payment_asset_address"`

	// Total amount of blocks to subscribe to
	NumBlocksToSubscribe uint64 `serialize:"true" json:"num_blocks_to_subscribe"`
}

func (*SubscribeDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.SubscribeDatasetMarketplaceID
}

func (d *SubscribeDatasetMarketplace) StateKeys(actor codec.Address) state.Keys {
	nftAddress := storage.AssetAddressNFT(d.MarketplaceAssetAddress, nil, actor)
	return state.Keys{
		string(storage.AssetInfoKey(d.MarketplaceAssetAddress)):                  state.Read | state.Write,
		string(storage.AssetInfoKey(nftAddress)):                                 state.All,
		string(storage.AssetAccountBalanceKey(d.MarketplaceAssetAddress, actor)): state.Allocate | state.Write,
		string(storage.AssetAccountBalanceKey(nftAddress, actor)):                state.All,
		string(storage.AssetInfoKey(d.PaymentAssetAddress)):                      state.Read | state.Write,
	}
}

func (d *SubscribeDatasetMarketplace) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the nftID already exists(This means the user is already subscribed)
	nftAddress := storage.AssetAddressNFT(d.MarketplaceAssetAddress, nil, actor)
	if storage.AssetExists(ctx, mu, nftAddress) {
		return nil, ErrUserAlreadySubscribed
	}

	// Ensure numBlocksToSubscribe is valid
	dataConfig := dataset.GetDatasetConfig()
	if d.NumBlocksToSubscribe < dataConfig.MinBlocksToSubscribe {
		return nil, ErrOutputNumBlocksToSubscribeInvalid
	}

	// Check for the asset
	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, mu, d.MarketplaceAssetAddress)
	if err != nil {
		return nil, err
	}
	// Ensure the asset is a marketplace token
	if assetType != nconsts.AssetMarketplaceTokenID {
		return nil, ErrAssetTypeInvalid
	}

	// Convert the metdata to a map
	metadataMap, err := utils.BytesToMap(metadata)
	if err != nil {
		return nil, err
	}
	// Ensure paymentAssetAddress is supported
	if metadataMap["paymentAssetAddress"] != d.PaymentAssetAddress.String() {
		return nil, ErrPaymentAssetNotSupported
	}

	// Calculate the total cost of the subscription
	datasetPricePerBlock, err := strconv.ParseUint(metadataMap["datasetPricePerBlock"], 10, 64)
	if err != nil {
		return nil, err
	}
	totalCost := d.NumBlocksToSubscribe * datasetPricePerBlock

	// Check if the actor has enough balance to subscribe
	if totalCost > 0 {
		balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, d.PaymentAssetAddress, actor)
		if err != nil {
			return nil, err
		}
		if balance < totalCost {
			return nil, storage.ErrInsufficientAssetBalance
		}
		newBalance, err := smath.Sub(balance, totalCost)
		if err != nil {
			return nil, err
		}
		if err = storage.SetAssetAccountBalance(ctx, mu, d.PaymentAssetAddress, actor, newBalance); err != nil {
			return nil, err
		}
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()
	currentBlock := emissionInstance.GetLastAcceptedBlockHeight()

	// Update the paymentRemaining, subscriptions and lastClaimedBlock fields
	prevPaymentRemaining, err := strconv.ParseUint(metadataMap["paymentRemaining"], 10, 64)
	if err != nil {
		return nil, err
	}
	metadataMap["paymentRemaining"] = fmt.Sprint(prevPaymentRemaining + totalCost)
	prevSubscriptions, err := strconv.ParseUint(metadataMap["subscriptions"], 10, 64)
	if err != nil {
		return nil, err
	}
	metadataMap["subscriptions"] = fmt.Sprint(prevSubscriptions + 1)
	prevStartSubscriptionBlock, err := strconv.ParseUint(metadataMap["lastClaimedBlock"], 10, 64)
	if err != nil {
		return nil, err
	}
	if prevStartSubscriptionBlock == 0 {
		metadataMap["lastClaimedBlock"] = fmt.Sprint(currentBlock)
	}
	// Marshal the map back to a JSON byte slice
	metadata, err = utils.MapToBytes(metadataMap)
	if err != nil {
		return nil, err
	}
	// Update the asset with the updated metadata
	if err := storage.SetAssetInfo(ctx, mu, d.MarketplaceAssetAddress, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	// Mint the NFT for the subscription
	metadataNFTMap := make(map[string]string, 0)
	metadataNFTMap["datasetAddress"] = metadataMap["datasetAddress"]
	metadataNFTMap["marketplaceAssetAddress"] = d.MarketplaceAssetAddress.String()
	metadataNFTMap["datasetPricePerBlock"] = metadataMap["datasetPricePerBlock"]
	metadataNFTMap["paymentAssetAddress"] = d.PaymentAssetAddress.String()
	metadataNFTMap["totalCost"] = fmt.Sprint(totalCost)
	metadataNFTMap["issuanceBlock"] = fmt.Sprint(currentBlock)
	metadataNFTMap["numBlocksToSubscribe"] = fmt.Sprint(d.NumBlocksToSubscribe)
	metadataNFTMap["expirationBlock"] = fmt.Sprint(currentBlock + d.NumBlocksToSubscribe)
	// Convert the map to a JSON string
	metadataNFT, err := utils.MapToBytes(metadataNFTMap)
	if err != nil {
		return nil, err
	}

	// Minting logic for non-fungible tokens
	if _, err := storage.MintAsset(ctx, mu, d.MarketplaceAssetAddress, actor, 1); err != nil {
		return nil, err
	}
	symbol = utils.CombineWithSuffix(symbol, totalSupply, storage.MaxSymbolSize)
	if err := storage.SetAssetInfo(ctx, mu, nftAddress, nconsts.AssetNonFungibleTokenID, name, symbol, 0, metadataNFT, []byte(d.MarketplaceAssetAddress.String()), 0, 1, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}
	if _, err := storage.MintAsset(ctx, mu, nftAddress, actor, 1); err != nil {
		return nil, err
	}

	return &SubscribeDatasetMarketplaceResult{
		MarketplaceAssetAddress:          d.MarketplaceAssetAddress.String(),
		MarketplaceAssetNumSubscriptions: prevSubscriptions + 1,
		SubscriptionNftAddress:           nftAddress.String(),
		PaymentAssetAddress:              d.PaymentAssetAddress.String(),
		DatasetPricePerBlock:             datasetPricePerBlock,
		TotalCost:                        totalCost,
		NumBlocksToSubscribe:             d.NumBlocksToSubscribe,
		IssuanceBlock:                    currentBlock,
		ExpirationBlock:                  currentBlock + d.NumBlocksToSubscribe,
	}, nil
}

func (*SubscribeDatasetMarketplace) ComputeUnits(chain.Rules) uint64 {
	return SubscribeDatasetMarketplaceComputeUnits
}

func (*SubscribeDatasetMarketplace) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalSubscribeDatasetMarketplace(p *codec.Packer) (chain.Action, error) {
	var subscribe SubscribeDatasetMarketplace
	p.UnpackAddress(&subscribe.MarketplaceAssetAddress)
	p.UnpackAddress(&subscribe.PaymentAssetAddress)
	subscribe.NumBlocksToSubscribe = p.UnpackUint64(true)
	return &subscribe, p.Err()
}

var (
	_ codec.Typed     = (*SubscribeDatasetMarketplaceResult)(nil)
	_ chain.Marshaler = (*SubscribeDatasetMarketplaceResult)(nil)
)

type SubscribeDatasetMarketplaceResult struct {
	MarketplaceAssetAddress          string `serialize:"true" json:"marketplace_asset_address"`
	MarketplaceAssetNumSubscriptions uint64 `serialize:"true" json:"marketplace_asset_num_subscriptions"`
	SubscriptionNftAddress           string `serialize:"true" json:"subscription_nft_address"`
	PaymentAssetAddress              string `serialize:"true" json:"payment_asset_address"`
	DatasetPricePerBlock             uint64 `serialize:"true" json:"dataset_price_per_block"`
	TotalCost                        uint64 `serialize:"true" json:"total_cost"`
	NumBlocksToSubscribe             uint64 `serialize:"true" json:"num_blocks_to_subscribe"`
	IssuanceBlock                    uint64 `serialize:"true" json:"issuance_block"`
	ExpirationBlock                  uint64 `serialize:"true" json:"expiration_block"`
}

func (*SubscribeDatasetMarketplaceResult) GetTypeID() uint8 {
	return nconsts.SubscribeDatasetMarketplaceID
}

func (r *SubscribeDatasetMarketplaceResult) Size() int {
	return codec.StringLen(r.MarketplaceAssetAddress) + consts.Uint64Len*6 + codec.StringLen(r.SubscriptionNftAddress) + codec.StringLen(r.PaymentAssetAddress)
}

func (r *SubscribeDatasetMarketplaceResult) Marshal(p *codec.Packer) {
	p.PackString(r.MarketplaceAssetAddress)
	p.PackLong(r.MarketplaceAssetNumSubscriptions)
	p.PackString(r.SubscriptionNftAddress)
	p.PackString(r.PaymentAssetAddress)
	p.PackUint64(r.DatasetPricePerBlock)
	p.PackUint64(r.TotalCost)
	p.PackUint64(r.NumBlocksToSubscribe)
	p.PackUint64(r.IssuanceBlock)
	p.PackUint64(r.ExpirationBlock)
}

func UnmarshalSubscribeDatasetMarketplaceResult(p *codec.Packer) (codec.Typed, error) {
	var result SubscribeDatasetMarketplaceResult
	result.MarketplaceAssetAddress = p.UnpackString(true)
	result.MarketplaceAssetNumSubscriptions = p.UnpackUint64(false)
	result.SubscriptionNftAddress = p.UnpackString(true)
	result.PaymentAssetAddress = p.UnpackString(true)
	result.DatasetPricePerBlock = p.UnpackUint64(false)
	result.TotalCost = p.UnpackUint64(false)
	result.NumBlocksToSubscribe = p.UnpackUint64(true)
	result.IssuanceBlock = p.UnpackUint64(true)
	result.ExpirationBlock = p.UnpackUint64(true)
	return &result, p.Err()
}
