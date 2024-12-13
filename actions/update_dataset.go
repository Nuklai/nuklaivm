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
	// DatasetAddress to update
	DatasetAddress codec.Address `serialize:"true" json:"dataset_address"`

	// The title of the dataset
	Name string `serialize:"true" json:"name"`

	// The description of the dataset
	Description string `serialize:"true" json:"description"`

	// The categories of the dataset
	Categories string `serialize:"true" json:"categories"`

	// License of the dataset
	LicenseName   string `serialize:"true" json:"license_name"`
	LicenseSymbol string `serialize:"true" json:"license_symbol"`
	LicenseURL    string `serialize:"true" json:"license_url"`

	// False for sole contributor and true for open contribution
	IsCommunityDataset bool `serialize:"true" json:"is_community_dataset"`
}

func (*UpdateDataset) GetTypeID() uint8 {
	return nconsts.UpdateDatasetID
}

func (u *UpdateDataset) StateKeys(codec.Address) state.Keys {
	return state.Keys{
		string(storage.DatasetInfoKey(u.DatasetAddress)): state.Read | state.Write,
	}
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
	name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, marketplaceAssetAddress, baseAssetAddress, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := storage.GetDatasetInfoNoController(ctx, mu, u.DatasetAddress)
	if err != nil {
		return nil, err
	}
	// Check if the actor is the owner of the dataset
	if owner != actor {
		return nil, ErrWrongOwner
	}

	// Ensure that at least one field is being updated
	if (len(u.Name) == 0 || bytes.Equal([]byte(u.Name), name)) && (len(u.Description) == 0 || bytes.Equal([]byte(u.Description), description)) && (len(u.Categories) == 0 || bytes.Equal([]byte(u.Categories), categories)) && (len(u.LicenseName) == 0 || bytes.Equal([]byte(u.LicenseName), licenseName)) && (len(u.LicenseSymbol) == 0 || bytes.Equal([]byte(u.LicenseSymbol), licenseSymbol)) && (len(u.LicenseURL) == 0 || bytes.Equal([]byte(u.LicenseURL), licenseURL)) && u.IsCommunityDataset == isCommunityDataset {
		return nil, ErrOutputMustUpdateAtLeastOneField
	}

	var updateDatasetResult UpdateDatasetResult
	updateDatasetResult.Actor = actor.String()
	updateDatasetResult.Receiver = ""

	// if u.Name is passed, update the dataset name
	// otherwise, keep the existing name
	if len(u.Name) > 0 {
		if len(u.Name) < 3 || len(u.Name) > storage.MaxNameSize {
			return nil, ErrNameInvalid
		}
		name = []byte(u.Name)
		updateDatasetResult.Name = u.Name
	}

	if len(u.Description) > 0 {
		if len(u.Description) < 3 || len(u.Description) > storage.MaxTextSize {
			return nil, ErrDescriptionInvalid
		}
		description = []byte(u.Description)
		updateDatasetResult.Description = u.Description
	}

	if len(u.Categories) > 0 {
		if len(u.Categories) < 3 || len(u.Categories) > storage.MaxTextSize {
			return nil, ErrCategoriesInvalid
		}
		categories = []byte(u.Categories)
		updateDatasetResult.Categories = u.Categories
	}

	if len(u.LicenseName) > 0 {
		if len(u.LicenseName) < 3 || len(u.LicenseName) > storage.MaxNameSize {
			return nil, ErrLicenseNameInvalid
		}
		licenseName = []byte(u.LicenseName)
		updateDatasetResult.LicenseName = u.LicenseName
	}

	if len(u.LicenseSymbol) > 0 {
		if len(u.LicenseSymbol) < 3 || len(u.LicenseSymbol) > storage.MaxSymbolSize {
			return nil, ErrLicenseSymbolInvalid
		}
		licenseSymbol = []byte(u.LicenseSymbol)
		updateDatasetResult.LicenseSymbol = u.LicenseSymbol
	}

	if len(u.LicenseURL) > 0 {
		if len(u.LicenseURL) < 3 || len(u.LicenseURL) > storage.MaxTextSize {
			return nil, ErrLicenseURLInvalid
		}
		licenseURL = []byte(u.LicenseURL)
		updateDatasetResult.LicenseURL = u.LicenseURL
	}

	if u.IsCommunityDataset {
		revenueModelDataOwnerCut = 10
		updateDatasetResult.IsCommunityDataset = true
	}

	// Update the dataset
	if err := storage.SetDatasetInfo(ctx, mu, u.DatasetAddress, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, u.IsCommunityDataset, marketplaceAssetAddress, baseAssetAddress, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner); err != nil {
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

func UnmarshalUpdateDataset(p *codec.Packer) (chain.Action, error) {
	var update UpdateDataset
	p.UnpackAddress(&update.DatasetAddress)
	update.Name = p.UnpackString(false)
	update.Description = p.UnpackString(false)
	update.Categories = p.UnpackString(false)
	update.LicenseName = p.UnpackString(false)
	update.LicenseSymbol = p.UnpackString(false)
	update.LicenseURL = p.UnpackString(false)
	update.IsCommunityDataset = p.UnpackBool()
	return &update, p.Err()
}

var (
	_ codec.Typed = (*UpdateDatasetResult)(nil)
)

type UpdateDatasetResult struct {
	Actor              string `serialize:"true" json:"actor"`
	Receiver           string `serialize:"true" json:"receiver"`
	Name               string `serialize:"true" json:"name"`
	Description        string `serialize:"true" json:"description"`
	Categories         string `serialize:"true" json:"categories"`
	LicenseName        string `serialize:"true" json:"license_name"`
	LicenseSymbol      string `serialize:"true" json:"license_symbol"`
	LicenseURL         string `serialize:"true" json:"license_url"`
	IsCommunityDataset bool   `serialize:"true" json:"is_community_dataset"`
}

func (*UpdateDatasetResult) GetTypeID() uint8 {
	return nconsts.UpdateDatasetID
}

func UnmarshalUpdateDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result UpdateDatasetResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.Name = p.UnpackString(false)
	result.Description = p.UnpackString(false)
	result.Categories = p.UnpackString(false)
	result.LicenseName = p.UnpackString(false)
	result.LicenseSymbol = p.UnpackString(false)
	result.LicenseURL = p.UnpackString(false)
	result.IsCommunityDataset = p.UnpackBool()
	return &result, p.Err()
}
