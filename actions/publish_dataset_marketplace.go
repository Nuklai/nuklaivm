// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*PublishDatasetMarketplace)(nil)

type PublishDatasetMarketplace struct {
	// Dataset ID
	Dataset ids.ID `json:"dataset"`

	// This is the asset ID to use for calculating the price for one block
	BaseAsset ids.ID `json:"baseAsset"`

	// THis is the base price in the `baseAsset` amount for one block
	// For example, if the base price is 1 and the base asset is NAI,
	// then the price for one block is 1 NAI
	BasePrice uint64 `json:"basePrice"`
}

func (*PublishDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.PublishDatasetMarketplaceID
}

func (d *PublishDatasetMarketplace) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(d.Dataset)):        state.Read,
		string(storage.BalanceKey(actor, ids.Empty)): state.Read | state.Write,
	}
}

func (*PublishDatasetMarketplace) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks, storage.BalanceChunks}
}

func (d *PublishDatasetMarketplace) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if the dataset exists
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, _, _, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDataset(ctx, mu, d.Dataset)
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

	// Check if the dataset is already on sale
	if onSale {
		return nil, ErrDatasetAlreadyOnSale
	}

	// Check that BaseAsset is supported
	if d.BaseAsset != ids.Empty {
		return nil, ErrBaseAssetNotSupported
	}
	// Check that BasePrice is valid
	if d.BasePrice == 0 {
		return nil, ErrBasePriceInvalid
	}

	// Update the dataset
	if err := storage.SetDataset(ctx, mu, d.Dataset, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, true, d.BaseAsset, d.BasePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
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
	p.UnpackID(true, &publish.BaseAsset)
	publish.BasePrice = p.UnpackUint64(true)
	return &publish, p.Err()
}

func (*PublishDatasetMarketplace) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
