// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	PublishDatasetMarketplaceComputeUnits = 5
)

var _ chain.Action = (*PublishDatasetMarketplace)(nil)

type PublishDatasetMarketplace struct {
	// DatasetAddress
	DatasetAddress codec.Address `serialize:"true" json:"dataset_address"`

	// This is the asset to use for calculating the price for one block
	PaymentAssetAddress codec.Address `serialize:"true" json:"payment_asset_address"`

	// This is the base price in the `baseAsset` amount for one block
	// For example, if the base price is 1 and the base asset is NAI,
	// then the price for one block is 1 NAI
	// If 0 is passed, the dataset is free to use
	DatasetPricePerBlock uint64 `serialize:"true" json:"dataset_price_per_block"`
}

func (*PublishDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.PublishDatasetMarketplaceID
}

func (d *PublishDatasetMarketplace) StateKeys(_ codec.Address) state.Keys {
	marketplaceAssetAddress := storage.AssetAddressFractional(d.DatasetAddress)
	return state.Keys{
		string(storage.AssetInfoKey(marketplaceAssetAddress)): state.All,
		string(storage.DatasetInfoKey(d.DatasetAddress)):      state.Read | state.Write,
	}
}

func (d *PublishDatasetMarketplace) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	marketplaceAssetAddress := storage.AssetAddressFractional(d.DatasetAddress)

	// Check if the marketplace asset already exists
	if storage.AssetExists(ctx, mu, marketplaceAssetAddress) {
		return nil, ErrAssetExists
	}

	// Check if the dataset exists
	name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, _, _, _, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDatasetInfoNoController(ctx, mu, d.DatasetAddress)
	if err != nil {
		return nil, err
	}
	// Check if the actor is the owner of the dataset
	if owner != actor {
		return nil, ErrWrongOwner
	}

	// Update the dataset
	if err := storage.SetDatasetInfo(ctx, mu, d.DatasetAddress, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, marketplaceAssetAddress, d.PaymentAssetAddress, d.DatasetPricePerBlock, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
		return nil, err
	}

	// Create an asset that represents that this dataset is published to the marketplace
	// This is a special type of token that cannot be manually created/minted
	metadataMap := make(map[string]string, 0)
	metadataMap["datasetAddress"] = d.DatasetAddress.String()
	metadataMap["marketplaceAssetAddress"] = marketplaceAssetAddress.String()
	metadataMap["datasetPricePerBlock"] = fmt.Sprint(d.DatasetPricePerBlock)
	metadataMap["paymentAssetAddress"] = d.PaymentAssetAddress.String()
	metadataMap["publisher"] = actor.String()
	metadataMap["lastClaimedBlock"] = "0"
	metadataMap["subscriptions"] = "0"
	metadataMap["paymentRemaining"] = "0"
	metadataMap["paymentClaimed"] = "0"
	// Convert the map to a JSON string
	metadata, err = utils.MapToBytes(metadataMap)
	if err != nil {
		return nil, err
	}
	// Create a new marketplace asset
	if err := storage.SetAssetInfo(ctx, mu, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte(storage.MarketplaceAssetName), []byte(storage.MarketplaceAssetSymbol), 0, metadata, []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}

	return &PublishDatasetMarketplaceResult{
		MarketplaceAssetAddress: marketplaceAssetAddress,
		PaymentAssetAddress:     d.PaymentAssetAddress,
		DatasetPricePerBlock:    d.DatasetPricePerBlock,
		Publisher:               actor,
	}, nil
}

func (*PublishDatasetMarketplace) ComputeUnits(chain.Rules) uint64 {
	return PublishDatasetMarketplaceComputeUnits
}

func (*PublishDatasetMarketplace) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalPublishDatasetMarketplace(p *codec.Packer) (chain.Action, error) {
	var publish PublishDatasetMarketplace
	p.UnpackAddress(&publish.DatasetAddress)
	p.UnpackAddress(&publish.PaymentAssetAddress)
	publish.DatasetPricePerBlock = p.UnpackUint64(false)
	return &publish, p.Err()
}

var (
	_ codec.Typed     = (*PublishDatasetMarketplaceResult)(nil)
	_ chain.Marshaler = (*PublishDatasetMarketplaceResult)(nil)
)

type PublishDatasetMarketplaceResult struct {
	MarketplaceAssetAddress codec.Address `serialize:"true" json:"marketplace_asset_address"`
	PaymentAssetAddress     codec.Address `serialize:"true" json:"payment_asset_address"`
	Publisher               codec.Address `serialize:"true" json:"publisher"`
	DatasetPricePerBlock    uint64        `serialize:"true" json:"dataset_price_per_block"`
}

func (*PublishDatasetMarketplaceResult) GetTypeID() uint8 {
	return nconsts.PublishDatasetMarketplaceID
}

func (*PublishDatasetMarketplaceResult) Size() int {
	return codec.AddressLen*3 + consts.Uint64Len
}

func (r *PublishDatasetMarketplaceResult) Marshal(p *codec.Packer) {
	p.PackAddress(r.MarketplaceAssetAddress)
	p.PackAddress(r.PaymentAssetAddress)
	p.PackAddress(r.Publisher)
	p.PackUint64(r.DatasetPricePerBlock)
}

func UnmarshalPublishDatasetMarketplaceResult(p *codec.Packer) (codec.Typed, error) {
	var result PublishDatasetMarketplaceResult
	p.UnpackAddress(&result.MarketplaceAssetAddress)
	p.UnpackAddress(&result.PaymentAssetAddress)
	p.UnpackAddress(&result.Publisher)
	result.DatasetPricePerBlock = p.UnpackUint64(false)
	return &result, p.Err()
}
