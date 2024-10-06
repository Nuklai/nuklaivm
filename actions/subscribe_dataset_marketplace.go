// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/marketplace"
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
	ErrMarketplaceAssetIDInvalid                      = errors.New("marketplace asset ID is invalid")
	ErrBaseAssetNotSupported                          = errors.New("base asset is not supported")
	ErrOutputNumBlocksToSubscribeInvalid              = errors.New("num blocks to subscribe is invalid")
	_                                    chain.Action = (*SubscribeDatasetMarketplace)(nil)
)

type SubscribeDatasetMarketplace struct {
	// DatasetID ID
	DatasetID ids.ID `serialize:"true" json:"dataset_id"`

	// Marketplace ID(This is also the asset ID in the marketplace that represents the dataset)
	MarketplaceAssetID ids.ID `serialize:"true" json:"marketplace_asset_id"`

	// Asset to use for the subscription
	AssetForPayment ids.ID `serialize:"true" json:"asset_for_payment"`

	// Total amount of blocks to subscribe to
	NumBlocksToSubscribe uint64 `serialize:"true" json:"num_blocks_to_subscribe"`
}

func (*SubscribeDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.SubscribeDatasetMarketplaceID
}

func (d *SubscribeDatasetMarketplace) StateKeys(actor codec.Address) state.Keys {
	nftID := utils.GenerateIDWithAddress(d.MarketplaceAssetID, actor)
	return state.Keys{
		string(storage.DatasetInfoKey(d.DatasetID)):             state.Read,
		string(storage.AssetInfoKey(d.MarketplaceAssetID)):      state.Read | state.Write,
		string(storage.AssetNFTKey(nftID)):                      state.All,
		string(storage.BalanceKey(actor, d.AssetForPayment)):    state.Read | state.Write,
		string(storage.BalanceKey(actor, d.MarketplaceAssetID)): state.Allocate | state.Write,
		string(storage.BalanceKey(actor, nftID)):                state.Allocate | state.Write,
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
	nftID := utils.GenerateIDWithAddress(d.MarketplaceAssetID, actor)
	exists, _, _, _, _, _, _ := storage.GetAssetNFT(ctx, mu, nftID)
	if exists {
		return nil, ErrNFTAlreadyExists
	}

	// Check if the dataset exists
	exists, _, _, _, _, _, _, _, _, saleID, baseAsset, basePrice, _, _, _, _, _, err := storage.GetDatasetInfoNoController(ctx, mu, d.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}

	// Check if the dataset is on sale
	if saleID == ids.Empty {
		return nil, ErrDatasetNotOnSale
	}
	// Check if the marketplace ID is correct
	if saleID != d.MarketplaceAssetID {
		return nil, ErrMarketplaceAssetIDInvalid
	}

	// Ensure assetForPayment is supported
	if d.AssetForPayment != baseAsset {
		return nil, ErrBaseAssetNotSupported
	}

	// Ensure numBlocksToSubscribe is valid
	dataConfig := marketplace.GetDatasetConfig()
	if d.NumBlocksToSubscribe < dataConfig.MinBlocksToSubscribe {
		return nil, ErrOutputNumBlocksToSubscribeInvalid
	}

	// Check for the asset
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAssetInfoNoController(ctx, mu, d.MarketplaceAssetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssetMissing
	}
	if assetType != nconsts.AssetMarketplaceTokenID {
		return nil, ErrOutputWrongAssetType
	}

	// Mint the subscription non-fungible token to represent the user is subscribed
	// to the dataset
	amountOfToken := uint64(1)
	newSupply, err := smath.Add(totalSupply, amountOfToken)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}
	totalSupply = newSupply

	// Calculate the total cost of the subscription
	totalCost := d.NumBlocksToSubscribe * basePrice

	// Check if the actor has enough balance to subscribe
	if totalCost > 0 {
		if _, err := storage.SubBalance(ctx, mu, actor, d.AssetForPayment, totalCost); err != nil {
			return nil, err
		}
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()
	currentBlock := emissionInstance.GetLastAcceptedBlockHeight()

	// Mint the NFT for the subscription
	metadataNFTMap := make(map[string]string, 0)
	metadataNFTMap["datasetID"] = d.DatasetID.String()
	metadataNFTMap["marketplaceAssetID"] = d.MarketplaceAssetID.String()
	metadataNFTMap["datasetPricePerBlock"] = fmt.Sprint(basePrice)
	metadataNFTMap["assetForPayment"] = d.AssetForPayment.String()
	metadataNFTMap["totalCost"] = fmt.Sprint(totalCost)
	metadataNFTMap["issuanceBlock"] = fmt.Sprint(currentBlock)
	metadataNFTMap["numBlocksToSubscribe"] = fmt.Sprint(d.NumBlocksToSubscribe)
	metadataNFTMap["expirationBlock"] = fmt.Sprint(currentBlock + d.NumBlocksToSubscribe)
	// Convert the map to a JSON string
	metadataNFT, err := utils.MapToBytes(metadataNFTMap)
	if err != nil {
		return nil, err
	}
	// Set the NFT metadata
	if err := storage.SetAssetNFT(ctx, mu, d.MarketplaceAssetID, totalSupply, nftID, []byte(d.DatasetID.String()), metadataNFT, actor); err != nil {
		return nil, err
	}

	// Unmarshal the metadata JSON into a map
	metadataMap, err := utils.BytesToMap(metadata)
	if err != nil {
		return nil, err
	}
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
	// Update the asset with the new total supply and updated metadata
	if err := storage.SetAssetInfo(ctx, mu, d.MarketplaceAssetID, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	if _, err := storage.AddBalance(ctx, mu, actor, d.MarketplaceAssetID, amountOfToken, true); err != nil {
		return nil, err
	}
	// Add the balance to individual NFT
	if _, err := storage.AddBalance(ctx, mu, actor, nftID, amountOfToken, true); err != nil {
		return nil, err
	}

	return &SubscribeDatasetMarketplaceResult{
		MarketplaceAssetID:               d.MarketplaceAssetID,
		MarketplaceAssetNumSubscriptions: prevSubscriptions + 1,
		SubscriptionNftID:                nftID,
		AssetForPayment:                  d.AssetForPayment,
		DatasetPricePerBlock:             basePrice,
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

var _ chain.Marshaler = (*SubscribeDatasetMarketplace)(nil)

func (*SubscribeDatasetMarketplace) Size() int {
	return ids.IDLen*3 + consts.Uint64Len
}

func (d *SubscribeDatasetMarketplace) Marshal(p *codec.Packer) {
	p.PackID(d.DatasetID)
	p.PackID(d.MarketplaceAssetID)
	p.PackID(d.AssetForPayment)
	p.PackUint64(d.NumBlocksToSubscribe)
}

func UnmarshalSubscribeDatasetMarketplace(p *codec.Packer) (chain.Action, error) {
	var subscribe SubscribeDatasetMarketplace
	p.UnpackID(true, &subscribe.DatasetID)
	p.UnpackID(true, &subscribe.MarketplaceAssetID)
	p.UnpackID(false, &subscribe.AssetForPayment)
	subscribe.NumBlocksToSubscribe = p.UnpackUint64(true)
	return &subscribe, p.Err()
}

var (
	_ codec.Typed     = (*SubscribeDatasetMarketplaceResult)(nil)
	_ chain.Marshaler = (*SubscribeDatasetMarketplaceResult)(nil)
)

type SubscribeDatasetMarketplaceResult struct {
	MarketplaceAssetID               ids.ID `serialize:"true" json:"marketplace_asset_id"`
	MarketplaceAssetNumSubscriptions uint64 `serialize:"true" json:"marketplace_asset_num_subscriptions"`
	SubscriptionNftID                ids.ID `serialize:"true" json:"subscription_nft_id"`
	AssetForPayment                  ids.ID `serialize:"true" json:"asset_for_payment"`
	DatasetPricePerBlock             uint64 `serialize:"true" json:"dataset_price_per_block"`
	TotalCost                        uint64 `serialize:"true" json:"total_cost"`
	NumBlocksToSubscribe             uint64 `serialize:"true" json:"num_blocks_to_subscribe"`
	IssuanceBlock                    uint64 `serialize:"true" json:"issuance_block"`
	ExpirationBlock                  uint64 `serialize:"true" json:"expiration_block"`
}

func (*SubscribeDatasetMarketplaceResult) GetTypeID() uint8 {
	return nconsts.SubscribeDatasetMarketplaceID
}

func (*SubscribeDatasetMarketplaceResult) Size() int {
	return ids.IDLen*3 + consts.Uint64Len*6
}

func (r *SubscribeDatasetMarketplaceResult) Marshal(p *codec.Packer) {
	p.PackID(r.MarketplaceAssetID)
	p.PackLong(r.MarketplaceAssetNumSubscriptions)
	p.PackID(r.SubscriptionNftID)
	p.PackID(r.AssetForPayment)
	p.PackUint64(r.DatasetPricePerBlock)
	p.PackUint64(r.TotalCost)
	p.PackUint64(r.NumBlocksToSubscribe)
	p.PackUint64(r.IssuanceBlock)
	p.PackUint64(r.ExpirationBlock)
}

func UnmarshalSubscribeDatasetMarketplaceResult(p *codec.Packer) (codec.Typed, error) {
	var result SubscribeDatasetMarketplaceResult
	p.UnpackID(true, &result.MarketplaceAssetID)
	result.MarketplaceAssetNumSubscriptions = p.UnpackUint64(false)
	p.UnpackID(true, &result.SubscriptionNftID)
	p.UnpackID(false, &result.AssetForPayment)
	result.DatasetPricePerBlock = p.UnpackUint64(false)
	result.TotalCost = p.UnpackUint64(false)
	result.NumBlocksToSubscribe = p.UnpackUint64(true)
	result.IssuanceBlock = p.UnpackUint64(true)
	result.ExpirationBlock = p.UnpackUint64(true)
	return &result, p.Err()
}
