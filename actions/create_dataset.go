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

var _ chain.Action = (*CreateDataset)(nil)

type CreateDataset struct {
	// The title of the dataset
	Name []byte `json:"name"`

	// The description of the dataset
	Description []byte `json:"description"`

	// The categories of the dataset
	Categories []byte `json:"categories"`

	// License of the dataset
	LicenseName   []byte `json:"licenseName"`
	LicenseSymbol []byte `json:"licenseSymbol"`
	LicenseURL    []byte `json:"licenseURL"`

	// Metadata of the dataset
	Metadata []byte `json:"metadata"`

	// False for sole contributor and true for open contribution
	IsCommunityDataset bool `json:"isCommunityDataset"`
}

func (*CreateDataset) GetTypeID() uint8 {
	return nconsts.CreateDatasetID
}

func (*CreateDataset) StateKeys(_ codec.Address, actionID ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(actionID)):   state.Allocate | state.Write,
		string(storage.DatasetKey(actionID)): state.Allocate | state.Write,
	}
}

func (*CreateDataset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.DatasetChunks}
}

func (c *CreateDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	actionID ids.ID,
) ([][]byte, error) {
	if len(c.Name) < 3 || len(c.Name) > MaxTextSize {
		return nil, ErrOutputNameInvalid
	}
	if len(c.Description) < 3 || len(c.Description) > MaxMetadataSize {
		return nil, ErrOutputDescriptionInvalid
	}
	if len(c.Categories) < 3 || len(c.Categories) > MaxMetadataSize {
		return nil, ErrOutputCategoriesInvalid
	}
	if len(c.LicenseName) < 3 || len(c.LicenseName) > MaxTextSize {
		return nil, ErrOutputLicenseNameInvalid
	}
	if len(c.LicenseSymbol) < 3 || len(c.LicenseSymbol) > MaxTextSize {
		return nil, ErrOutputLicenseSymbolInvalid
	}
	if len(c.LicenseURL) < 3 || len(c.LicenseURL) > MaxMetadataSize {
		return nil, ErrOutputLicenseURLInvalid
	}
	if len(c.Metadata) < 3 || len(c.Metadata) > MaxDatasetMetadataSize {
		return nil, ErrOutputMetadataInvalid
	}

	// Create a new dataset with the following parameters:
	// onSale = false
	// baseAsset = ids.Empty
	// basePrice = 0
	// revenueModelDataShare = 100
	// revenueModelMetadataShare = 0
	// revenueModeldataOwnerCut = 10
	// revenueModelMetadataOwnerCut = 0
	if err := storage.SetDataset(ctx, mu, actionID, c.Name, c.Description, c.Categories, c.LicenseName, c.LicenseSymbol, c.LicenseURL, c.Metadata, c.IsCommunityDataset, false, ids.Empty, 0, 100, 0, 10, 0, actor); err != nil {
		return nil, err
	}

	// Create an asset for the dataset
	if err := storage.SetAsset(ctx, mu, actionID, c.Name, c.Name, 0, c.Description, 0, 0, actor, actor, actor, actor, actor, actor); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*CreateDataset) ComputeUnits(chain.Rules) uint64 {
	return CreateDatasetComputeUnits
}

func (c *CreateDataset) Size() int {
	return codec.BytesLen(c.Name) + codec.BytesLen(c.Description) + codec.BytesLen(c.Categories) + codec.BytesLen(c.LicenseName) + codec.BytesLen(c.LicenseSymbol) + codec.BytesLen(c.LicenseURL) + codec.BytesLen(c.Metadata) + consts.BoolLen
}

func (c *CreateDataset) Marshal(p *codec.Packer) {
	p.PackBytes(c.Name)
	p.PackBytes(c.Description)
	p.PackBytes(c.Categories)
	p.PackBytes(c.LicenseName)
	p.PackBytes(c.LicenseSymbol)
	p.PackBytes(c.LicenseURL)
	p.PackBytes(c.Metadata)
	p.PackBool(c.IsCommunityDataset)
}

func UnmarshalCreateDataset(p *codec.Packer) (chain.Action, error) {
	var create CreateDataset
	p.UnpackBytes(MaxTextSize, true, &create.Name)
	p.UnpackBytes(MaxMetadataSize, true, &create.Description)
	p.UnpackBytes(MaxMetadataSize, true, &create.Categories)
	p.UnpackBytes(MaxTextSize, true, &create.LicenseName)
	p.UnpackBytes(MaxTextSize, true, &create.LicenseSymbol)
	p.UnpackBytes(MaxMetadataSize, true, &create.LicenseURL)
	p.UnpackBytes(MaxMetadataSize, true, &create.Metadata)
	create.IsCommunityDataset = p.UnpackBool()
	return &create, p.Err()
}

func (*CreateDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
