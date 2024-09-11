// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*InitiateContributeDataset)(nil)

type InitiateContributeDataset struct {
	// Dataset ID
	Dataset ids.ID `json:"dataset"`

	// Data location(default, S3, Filecoin, etc.)
	DataLocation []byte `json:"dataLocation"`

	// Data Identifier(id/hash/URL)
	DataIdentifier []byte `json:"dataIdentifier"`
}

func (*InitiateContributeDataset) GetTypeID() uint8 {
	return nconsts.InitiateContributeDatasetID
}

func (d *InitiateContributeDataset) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(d.Dataset)): state.Read,
	}
}

func (*InitiateContributeDataset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks}
}

func (d *InitiateContributeDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	/* 	// Check if the dataset exists
	   	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDataset(ctx, mu, d.Dataset)
	   	if err != nil {
	   		return nil, err
	   	}
	   	if !exists {
	   		return nil, ErrDatasetNotFound
	   	} */

	// Check if the data location is valid
	dataLocation := []byte("default")
	if len(d.DataLocation) > 0 {
		dataLocation = d.DataLocation
	}
	if len(dataLocation) < 3 || len(dataLocation) > MaxTextSize {
		return nil, ErrOutputDataLocationInvalid
	}
	// Check if the data identifier is valid
	if len(d.DataIdentifier) == 0 || len(d.DataIdentifier) > MaxMetadataSize {
		return nil, ErrOutputURIInvalid
	}

	/* 	// Retrieve the asset info
	   	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, d.Dataset)
	   	if err != nil {
	   		return nil, err
	   	}
	   	if !exists {
	   		return nil, ErrAssetNotFound
	   	} */

	return nil, nil
}

func (*InitiateContributeDataset) ComputeUnits(chain.Rules) uint64 {
	return InitiateContributeDatasetComputeUnits
}

func (d *InitiateContributeDataset) Size() int {
	return ids.IDLen + codec.BytesLen(d.DataLocation) + codec.BytesLen(d.DataIdentifier)
}

func (d *InitiateContributeDataset) Marshal(p *codec.Packer) {
	p.PackID(d.Dataset)
	p.PackBytes(d.DataLocation)
	p.PackBytes(d.DataIdentifier)
}

func UnmarshalInitiateContributeDataset(p *codec.Packer) (chain.Action, error) {
	var initiate InitiateContributeDataset
	p.UnpackID(true, &initiate.Dataset)
	p.UnpackBytes(MaxTextSize, false, &initiate.DataLocation)
	p.UnpackBytes(MaxMetadataSize, true, &initiate.DataIdentifier)
	return &initiate, p.Err()
}

func (*InitiateContributeDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
