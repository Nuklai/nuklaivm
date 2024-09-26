// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"bytes"
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	UpdateDatasetComputeUnits = 5
)

var (
	ErrDatasetNotFound              = errors.New("dataset not found")
	_                  chain.Action = (*UpdateDataset)(nil)
)

type UpdateDataset struct {
	// DatasetID ID to update
	DatasetID ids.ID `serialize:"true" json:"dataset_id"`

	// The title of the dataset
	Name []byte `serialize:"true" json:"name"`

	// The description of the dataset
	Description []byte `serialize:"true" json:"description"`

	// The categories of the dataset
	Categories []byte `serialize:"true" json:"categories"`

	// License of the dataset
	LicenseName   []byte `serialize:"true" json:"license_name"`
	LicenseSymbol []byte `serialize:"true" json:"license_symbol"`
	LicenseURL    []byte `serialize:"true" json:"license_url"`

	// False for sole contributor and true for open contribution
	IsCommunityDataset bool `serialize:"true" json:"is_community_dataset"`
}

func (*UpdateDataset) GetTypeID() uint8 {
	return nconsts.UpdateDatasetID
}

func (u *UpdateDataset) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(u.DatasetID)): state.Allocate | state.Write,
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
) (codec.Typed, error) {
	// Check if the dataset exists
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDataset(ctx, mu, u.DatasetID)
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

	updateDatasetResult := UpdateDatasetResult{
		OldDatasetInfo: DatasetInfo{
			Name:               string(name),
			Description:        string(description),
			Categories:         string(categories),
			LicenseName:        string(licenseName),
			LicenseSymbol:      string(licenseSymbol),
			LicenseURL:         string(licenseURL),
			IsCommunityDataset: isCommunityDataset,
		},
		NewDatasetInfo: DatasetInfo{
			Name:               string(name),
			Description:        string(description),
			Categories:         string(categories),
			LicenseName:        string(licenseName),
			LicenseSymbol:      string(licenseSymbol),
			LicenseURL:         string(licenseURL),
			IsCommunityDataset: isCommunityDataset,
		},
	}

	// if u.Name is passed, update the dataset name
	// otherwise, keep the existing name
	if len(u.Name) > 0 {
		if len(u.Name) < 3 || len(u.Name) > MaxMetadataSize {
			return nil, ErrOutputNameInvalid
		}
		name = u.Name
		updateDatasetResult.NewDatasetInfo.Name = string(name)
	}

	if len(u.Description) > 0 {
		if len(u.Description) < 3 || len(u.Description) > MaxMetadataSize {
			return nil, ErrOutputDescriptionInvalid
		}
		description = u.Description
		updateDatasetResult.NewDatasetInfo.Description = string(description)
	}

	if len(u.Categories) > 0 {
		if len(u.Categories) < 3 || len(u.Categories) > MaxMetadataSize {
			return nil, ErrOutputCategoriesInvalid
		}
		categories = u.Categories
		updateDatasetResult.NewDatasetInfo.Categories = string(categories)
	}

	if len(u.LicenseName) > 0 {
		if len(u.LicenseName) < 3 || len(u.LicenseName) > MaxMetadataSize {
			return nil, ErrOutputLicenseNameInvalid
		}
		licenseName = u.LicenseName
		updateDatasetResult.NewDatasetInfo.LicenseName = string(licenseName)
	}

	if len(u.LicenseSymbol) > 0 {
		if len(u.LicenseSymbol) < 3 || len(u.LicenseSymbol) > MaxTextSize {
			return nil, ErrOutputLicenseSymbolInvalid
		}
		licenseSymbol = u.LicenseSymbol
		updateDatasetResult.NewDatasetInfo.LicenseSymbol = string(licenseSymbol)
	}

	if len(u.LicenseURL) > 0 {
		if len(u.LicenseURL) < 3 || len(u.LicenseURL) > MaxMetadataSize {
			return nil, ErrOutputLicenseURLInvalid
		}
		licenseURL = u.LicenseURL
		updateDatasetResult.NewDatasetInfo.LicenseURL = string(licenseURL)
	}

	if u.IsCommunityDataset {
		revenueModelDataOwnerCut = 10
		updateDatasetResult.NewDatasetInfo.IsCommunityDataset = true
	}

	// Update the dataset
	if err := storage.SetDataset(ctx, mu, u.DatasetID, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, u.IsCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
		return nil, err
	}

	return &updateDatasetResult, nil
}

func (*UpdateDataset) ComputeUnits(chain.Rules) uint64 {
	return UpdateDatasetComputeUnits
}

func (*UpdateDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*UpdateDataset)(nil)

func (u *UpdateDataset) Size() int {
	return ids.IDLen + codec.BytesLen(u.Name) + codec.BytesLen(u.Description) + codec.BytesLen(u.Categories) + codec.BytesLen(u.LicenseName) + codec.BytesLen(u.LicenseSymbol) + codec.BytesLen(u.LicenseURL) + consts.BoolLen
}

func (u *UpdateDataset) Marshal(p *codec.Packer) {
	p.PackID(u.DatasetID)
	p.PackBytes(u.Name)
	p.PackBytes(u.Description)
	p.PackBytes(u.Categories)
	p.PackBytes(u.LicenseName)
	p.PackBytes(u.LicenseSymbol)
	p.PackBytes(u.LicenseURL)
	p.PackBool(u.IsCommunityDataset)
}

func UnmarshalUpdateDataset(p *codec.Packer) (chain.Action, error) {
	var update UpdateDataset
	p.UnpackID(true, &update.DatasetID)
	p.UnpackBytes(MaxMetadataSize, false, &update.Name)
	p.UnpackBytes(MaxMetadataSize, false, &update.Description)
	p.UnpackBytes(MaxMetadataSize, false, &update.Categories)
	p.UnpackBytes(MaxMetadataSize, false, &update.LicenseName)
	p.UnpackBytes(MaxTextSize, false, &update.LicenseSymbol)
	p.UnpackBytes(MaxMetadataSize, false, &update.LicenseURL)
	update.IsCommunityDataset = p.UnpackBool()
	return &update, p.Err()
}

var _ codec.Typed = (*UpdateDatasetResult)(nil)

type DatasetInfo struct {
	Name               string `serialize:"true" json:"name"`
	Description        string `serialize:"true" json:"description"`
	Categories         string `serialize:"true" json:"categories"`
	LicenseName        string `serialize:"true" json:"license_name"`
	LicenseSymbol      string `serialize:"true" json:"license_symbol"`
	LicenseURL         string `serialize:"true" json:"license_url"`
	IsCommunityDataset bool   `serialize:"true" json:"is_community_dataset"`
}

type UpdateDatasetResult struct {
	OldDatasetInfo DatasetInfo `serialize:"true" json:"old_dataset_info"`
	NewDatasetInfo DatasetInfo `serialize:"true" json:"new_dataset_info"`
}

func (*UpdateDatasetResult) GetTypeID() uint8 {
	return nconsts.UpdateDatasetID
}
