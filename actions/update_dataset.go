// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"bytes"
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*UpdateDataset)(nil)

type UpdateDataset struct {
	// Dataset ID to update
	Dataset ids.ID `json:"dataset"`

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

	// False for sole contributor and true for open contribution
	IsCommunityDataset bool `json:"isCommunityDataset"`
}

func (*UpdateDataset) GetTypeID() uint8 {
	return nconsts.UpdateDatasetID
}

func (u *UpdateDataset) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetDatasetKey(u.Dataset)): state.Allocate | state.Write,
	}
}

func (*UpdateDataset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks}
}

func (u *UpdateDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if the dataset exists
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetAssetDataset(ctx, mu, u.Dataset)
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

	// Ensure that at least one field is being updated
	if (len(u.Name) == 0 || bytes.Equal(u.Name, name)) && (len(u.Description) == 0 || bytes.Equal(u.Description, description)) && (len(u.Categories) == 0 || bytes.Equal(u.Categories, categories)) && (len(u.LicenseName) == 0 || bytes.Equal(u.LicenseName, licenseName)) && (len(u.LicenseSymbol) == 0 || bytes.Equal(u.LicenseSymbol, licenseSymbol)) && (len(u.LicenseURL) == 0 || bytes.Equal(u.LicenseURL, licenseURL)) && u.IsCommunityDataset == isCommunityDataset {
		return nil, ErrOutputMustUpdateAtLeastOneField
	}

	// if u.Name is passed, update the dataset name
	// otherwise, keep the existing name
	if len(u.Name) > 0 {
		if len(u.Name) < 3 || len(u.Name) > MaxMetadataSize {
			return nil, ErrOutputNameInvalid
		}
		name = u.Name
	}

	if len(u.Description) > 0 {
		if len(u.Description) < 3 || len(u.Description) > MaxMetadataSize {
			return nil, ErrOutputDescriptionInvalid
		}
		description = u.Description
	}

	if len(u.Categories) > 0 {
		if len(u.Categories) < 3 || len(u.Categories) > MaxMetadataSize {
			return nil, ErrOutputCategoriesInvalid
		}
		categories = u.Categories
	}

	if len(u.LicenseName) > 0 {
		if len(u.LicenseName) < 3 || len(u.LicenseName) > MaxMetadataSize {
			return nil, ErrOutputLicenseNameInvalid
		}
		licenseName = u.LicenseName
	}

	if len(u.LicenseSymbol) > 0 {
		if len(u.LicenseSymbol) < 3 || len(u.LicenseSymbol) > MaxTextSize {
			return nil, ErrOutputLicenseSymbolInvalid
		}
		licenseSymbol = u.LicenseSymbol
	}

	if len(u.LicenseURL) > 0 {
		if len(u.LicenseURL) < 3 || len(u.LicenseURL) > MaxMetadataSize {
			return nil, ErrOutputLicenseURLInvalid
		}
		licenseURL = u.LicenseURL
	}

	if u.IsCommunityDataset {
		revenueModelDataOwnerCut = 10
	}

	// Update the dataset
	if err := storage.SetAssetDataset(ctx, mu, u.Dataset, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, u.IsCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*UpdateDataset) ComputeUnits(chain.Rules) uint64 {
	return UpdateDatasetComputeUnits
}

func (u *UpdateDataset) Size() int {
	return ids.IDLen + codec.BytesLen(u.Name) + codec.BytesLen(u.Description) + codec.BytesLen(u.Categories) + codec.BytesLen(u.LicenseName) + codec.BytesLen(u.LicenseSymbol) + codec.BytesLen(u.LicenseURL) + consts.BoolLen
}

func (c *UpdateDataset) Marshal(p *codec.Packer) {
	p.PackID(c.Dataset)
	p.PackBytes(c.Name)
	p.PackBytes(c.Description)
	p.PackBytes(c.Categories)
	p.PackBytes(c.LicenseName)
	p.PackBytes(c.LicenseSymbol)
	p.PackBytes(c.LicenseURL)
	p.PackBool(c.IsCommunityDataset)
}

func UnmarshalUpdateDataset(p *codec.Packer) (chain.Action, error) {
	var update UpdateDataset
	p.UnpackID(true, &update.Dataset)
	p.UnpackBytes(MaxMetadataSize, false, &update.Name)
	p.UnpackBytes(MaxMetadataSize, false, &update.Description)
	p.UnpackBytes(MaxMetadataSize, false, &update.Categories)
	p.UnpackBytes(MaxMetadataSize, false, &update.LicenseName)
	p.UnpackBytes(MaxTextSize, false, &update.LicenseSymbol)
	p.UnpackBytes(MaxMetadataSize, false, &update.LicenseURL)
	update.IsCommunityDataset = p.UnpackBool()
	return &update, p.Err()
}

func (*UpdateDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
