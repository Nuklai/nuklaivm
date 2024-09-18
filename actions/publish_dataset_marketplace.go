// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*PublishDatasetMarketplace)(nil)

type PublishDatasetMarketplace struct {
	// Dataset ID
	Dataset ids.ID `json:"dataset"`

	// This is the asset ID to use for calculating the price for one block
	BaseAsset ids.ID `json:"baseAsset"`

	// This is the base price in the `baseAsset` amount for one block
	// For example, if the base price is 1 and the base asset is NAI,
	// then the price for one block is 1 NAI
	// If 0 is passed, the dataset is free to use
	BasePrice uint64 `json:"basePrice"`
}

func (*PublishDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.PublishDatasetMarketplaceID
}

func (d *PublishDatasetMarketplace) StateKeys(actor codec.Address, actionID ids.ID) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(d.Dataset)): state.Read | state.Write,
		string(storage.AssetKey(d.Dataset)):   state.Read,
		string(storage.AssetKey(actionID)):    state.Allocate | state.Write,
	}
}

func (*PublishDatasetMarketplace) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks, storage.AssetChunks}
}

func (d *PublishDatasetMarketplace) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	actionID ids.ID,
) ([][]byte, error) {
	// Check if the dataset exists
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, _, _, _, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDataset(ctx, mu, d.Dataset)
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
	if err := storage.SetDataset(ctx, mu, d.Dataset, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, actionID, d.BaseAsset, d.BasePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
		return nil, err
	}

	// Retrieve the asset info
	exists, _, name, symbol, _, _, _, _, _, _, _, _, _, _, err := storage.GetAsset(ctx, mu, d.Dataset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssetNotFound
	}
	name = nchain.CombineWithPrefix([]byte("Dataset-Marketplace-"), name, MaxMetadataSize)
	symbol = nchain.CombineWithPrefix([]byte("DM-"), symbol, MaxTextSize)

	// Create an asset that represents that this dataset is published to the marketplace
	// This is a special type of token that cannot be manually created/minted
	metadata = []byte("{\"dataset\":\"" + d.Dataset.String() + "\",\"datasetPricePerBlock\":\"" + fmt.Sprint(d.BasePrice) + "\",\"assetForPayment\":\"" + d.BaseAsset.String() + "\",\"publisher\":\"" + codec.MustAddressBech32(nconsts.HRP, actor) + "\"}")
	if err := storage.SetAsset(ctx, mu, actionID, nconsts.AssetMarketplaceTokenID, name, symbol, 0, metadata, []byte(d.Dataset.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*PublishDatasetMarketplace) ComputeUnits(chain.Rules) uint64 {
	return PublishDatasetMarketplaceComputeUnits
}

func (*PublishDatasetMarketplace) Size() int {
	return ids.IDLen*2 + consts.Uint64Len
}

func (d *PublishDatasetMarketplace) Marshal(p *codec.Packer) {
	p.PackID(d.Dataset)
	p.PackID(d.BaseAsset)
	p.PackUint64(d.BasePrice)
}

func UnmarshalPublishDatasetMarketplace(p *codec.Packer) (chain.Action, error) {
	var publish PublishDatasetMarketplace
	p.UnpackID(true, &publish.Dataset)
	p.UnpackID(false, &publish.BaseAsset)
	publish.BasePrice = p.UnpackUint64(false)
	return &publish, p.Err()
}

func (*PublishDatasetMarketplace) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
