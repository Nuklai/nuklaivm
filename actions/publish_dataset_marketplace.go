// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	PublishDatasetMarketplaceComputeUnits = 5
)

var _ chain.Action = (*PublishDatasetMarketplace)(nil)

type PublishDatasetMarketplace struct {
	// DatasetID ID
	DatasetID ids.ID `serialize:"true" json:"dataset_id"`

	// This is the asset ID to use for calculating the price for one block
	BaseAssetID ids.ID `serialize:"true" json:"base_asset_id"`

	// This is the base price in the `baseAsset` amount for one block
	// For example, if the base price is 1 and the base asset is NAI,
	// then the price for one block is 1 NAI
	// If 0 is passed, the dataset is free to use
	BasePrice uint64 `serialize:"true" json:"base_price"`
}

func (*PublishDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.PublishDatasetMarketplaceID
}

func (d *PublishDatasetMarketplace) StateKeys(_ codec.Address, actionID ids.ID) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(d.DatasetID)): state.Read | state.Write,
		string(storage.AssetKey(d.DatasetID)):   state.Read,
		string(storage.AssetKey(actionID)):      state.Allocate | state.Write,
	}
}

func (*PublishDatasetMarketplace) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks, storage.AssetChunks, storage.AssetChunks}
}

func (d *PublishDatasetMarketplace) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	actionID ids.ID,
) (codec.Typed, error) {
	// Check if the dataset exists
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, _, _, _, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDataset(ctx, mu, d.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}

	// Check if the actor is the owner of the dataset
	if owner != actor {
		return nil, ErrOutputWrongOwner
	}

	// Update the dataset
	if err := storage.SetDataset(ctx, mu, d.DatasetID, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, actionID, d.BaseAssetID, d.BasePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
		return nil, err
	}

	// Retrieve the asset info
	exists, _, name, symbol, _, _, _, _, _, _, _, _, _, _, err := storage.GetAsset(ctx, mu, d.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetNotFound
	}
	name = nchain.CombineWithPrefix([]byte("Dataset-Marketplace-"), name, MaxMetadataSize)
	symbol = nchain.CombineWithPrefix([]byte("DM-"), symbol, MaxTextSize)

	// Create an asset that represents that this dataset is published to the marketplace
	// This is a special type of token that cannot be manually created/minted
	metadataMap := make(map[string]string, 0)
	metadataMap["datasetID"] = d.DatasetID.String()
	metadataMap["marketplaceAssetID"] = actionID.String()
	metadataMap["datasetPricePerBlock"] = fmt.Sprint(d.BasePrice)
	metadataMap["assetForPayment"] = d.BaseAssetID.String()
	metadataMap["publisher"] = actor.String()
	metadataMap["lastClaimedBlock"] = "0"
	metadataMap["subscriptions"] = "0"
	metadataMap["paymentRemaining"] = "0"
	metadataMap["paymentClaimed"] = "0"
	// Convert the map to a JSON string
	metadata, err = nchain.MapToBytes(metadataMap)
	if err != nil {
		return nil, err
	}
	if err := storage.SetAsset(ctx, mu, actionID, nconsts.AssetMarketplaceTokenID, name, symbol, 0, metadata, []byte(d.DatasetID.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}

	return &PublishDatasetMarketplaceResult{
		MarketplaceAssetID:   actionID,
		AssetForPayment:      d.BaseAssetID,
		DatasetPricePerBlock: d.BasePrice,
		Publisher:            actor,
	}, nil
}

func (*PublishDatasetMarketplace) ComputeUnits(chain.Rules) uint64 {
	return PublishDatasetMarketplaceComputeUnits
}

func (*PublishDatasetMarketplace) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*PublishDatasetMarketplace)(nil)

func (*PublishDatasetMarketplace) Size() int {
	return ids.IDLen*2 + consts.Uint64Len
}

func (d *PublishDatasetMarketplace) Marshal(p *codec.Packer) {
	p.PackID(d.DatasetID)
	p.PackID(d.BaseAssetID)
	p.PackUint64(d.BasePrice)
}

func UnmarshalPublishDatasetMarketplace(p *codec.Packer) (chain.Action, error) {
	var publish PublishDatasetMarketplace
	p.UnpackID(true, &publish.DatasetID)
	p.UnpackID(false, &publish.BaseAssetID)
	publish.BasePrice = p.UnpackUint64(false)
	return &publish, p.Err()
}

var (
	_ codec.Typed     = (*PublishDatasetMarketplaceResult)(nil)
	_ chain.Marshaler = (*PublishDatasetMarketplaceResult)(nil)
)

type PublishDatasetMarketplaceResult struct {
	MarketplaceAssetID   ids.ID        `serialize:"true" json:"marketplace_asset_id"`
	AssetForPayment      ids.ID        `serialize:"true" json:"asset_for_payment"`
	DatasetPricePerBlock uint64        `serialize:"true" json:"dataset_price_per_block"`
	Publisher            codec.Address `serialize:"true" json:"publisher"`
}

func (*PublishDatasetMarketplaceResult) GetTypeID() uint8 {
	return nconsts.PublishDatasetMarketplaceID
}

func (*PublishDatasetMarketplaceResult) Size() int {
	return ids.IDLen*2 + consts.Uint64Len + codec.AddressLen
}

func (r *PublishDatasetMarketplaceResult) Marshal(p *codec.Packer) {
	p.PackID(r.MarketplaceAssetID)
	p.PackID(r.AssetForPayment)
	p.PackUint64(r.DatasetPricePerBlock)
	p.PackAddress(r.Publisher)
}

func UnmarshalPublishDatasetMarketplaceResult(p *codec.Packer) (codec.Typed, error) {
	var result PublishDatasetMarketplaceResult
	p.UnpackID(true, &result.MarketplaceAssetID)
	p.UnpackID(false, &result.AssetForPayment)
	result.DatasetPricePerBlock = p.UnpackUint64(false)
	p.UnpackAddress(&result.Publisher)
	return &result, p.Err()
}
