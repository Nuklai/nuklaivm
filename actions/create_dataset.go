// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
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
	CreateDatasetComputeUnits = 100
)

var (
	ErrDatasetAlreadyExists              = errors.New("dataset already exists")
	ErrDescriptionInvalid                = errors.New("description is invalid")
	ErrCategoriesInvalid                 = errors.New("categories is invalid")
	ErrLicenseNameInvalid                = errors.New("license name is invalid")
	ErrLicenseSymbolInvalid              = errors.New("license symbol is invalid")
	ErrLicenseURLInvalid                 = errors.New("license url is invalid")
	_                       chain.Action = (*CreateDataset)(nil)
)

type CreateDataset struct {
	// Asset id if it was already created
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

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

	// Metadata of the dataset
	Metadata string `serialize:"true" json:"metadata"`

	// False for sole contributor and true for open contribution
	IsCommunityDataset bool `serialize:"true" json:"is_community_dataset"`
}

func (*CreateDataset) GetTypeID() uint8 {
	return nconsts.CreateDatasetID
}

func (c *CreateDataset) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(c.AssetAddress)):                  state.Read | state.Write,
		string(storage.DatasetInfoKey(c.AssetAddress)):                state.All,
		string(storage.AssetAccountBalanceKey(c.AssetAddress, actor)): state.All,
	}
}

func (c *CreateDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	if len(c.Name) < 3 || len(c.Name) > storage.MaxNameSize {
		return nil, ErrNameInvalid
	}
	if len(c.Description) > storage.MaxTextSize {
		return nil, ErrDescriptionInvalid
	}
	if len(c.Categories) > storage.MaxTextSize {
		return nil, ErrCategoriesInvalid
	}
	if len(c.LicenseName) > storage.MaxNameSize {
		return nil, ErrLicenseNameInvalid
	}
	if len(c.LicenseSymbol) > storage.MaxSymbolSize {
		return nil, ErrLicenseSymbolInvalid
	}
	if len(c.LicenseURL) > storage.MaxTextSize {
		return nil, ErrLicenseURLInvalid
	}
	if len(c.Metadata) > storage.MaxDatasetMetadataSize {
		return nil, ErrMetadataInvalid
	}

	// Check if the asset exists
	assetType, _, _, _, metadata, _, _, _, owner, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, c.AssetAddress)
	if err != nil {
		return nil, err
	}
	if assetType != nconsts.AssetFractionalTokenID {
		return nil, ErrAssetTypeInvalid
	}
	if owner != actor {
		return nil, ErrWrongOwner
	}

	revenueModelDataShare, revenueModelDataOwnerCut := 100, 100
	if c.IsCommunityDataset {
		revenueModelDataOwnerCut = 10
	}

	// Continue only if dataset doesn't exist
	if storage.DatasetExists(ctx, mu, c.AssetAddress) {
		return nil, ErrDatasetAlreadyExists
	}

	// Create a new dataset with the following parameters:
	// saleID = ids.Empty
	// baseAsset = ids.Empty
	// basePrice = 0
	// revenueModelDataShare = 100
	// revenueModelMetadataShare = 0
	// revenueModelDataOwnerCut = 10 for community datasets, 100 for sole contributor datasets
	// revenueModelMetadataOwnerCut = 0
	if err := storage.SetDatasetInfo(ctx, mu, c.AssetAddress, []byte(c.Name), []byte(c.Description), []byte(c.Categories), []byte(c.LicenseName), []byte(c.LicenseSymbol), []byte(c.LicenseURL), []byte(c.Metadata), c.IsCommunityDataset, codec.EmptyAddress, codec.EmptyAddress, 0, uint8(revenueModelDataShare), 0, uint8(revenueModelDataOwnerCut), 0, actor); err != nil {
		return nil, err
	}

	return &CreateDatasetResult{
		Actor:                   actor.String(),
		Receiver:                actor.String(),
		DatasetAddress:          c.AssetAddress.String(),
		DatasetParentNftAddress: storage.AssetAddressNFT(c.AssetAddress, metadata, owner).String(),
	}, nil
}

func (*CreateDataset) ComputeUnits(chain.Rules) uint64 {
	return CreateDatasetComputeUnits
}

func (*CreateDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalCreateDataset(p *codec.Packer) (chain.Action, error) {
	var create CreateDataset
	p.UnpackAddress(&create.AssetAddress)
	create.Name = p.UnpackString(true)
	create.Description = p.UnpackString(false)
	create.Categories = p.UnpackString(false)
	create.LicenseName = p.UnpackString(false)
	create.LicenseSymbol = p.UnpackString(false)
	create.LicenseURL = p.UnpackString(false)
	create.Metadata = p.UnpackString(false)
	create.IsCommunityDataset = p.UnpackBool()
	return &create, p.Err()
}

var _ codec.Typed = (*CreateDatasetResult)(nil)

type CreateDatasetResult struct {
	Actor                   string `serialize:"true" json:"actor"`
	Receiver                string `serialize:"true" json:"receiver"`
	DatasetAddress          string `serialize:"true" json:"dataset_address"`
	DatasetParentNftAddress string `serialize:"true" json:"dataset_parent_nft_address"`
}

func (*CreateDatasetResult) GetTypeID() uint8 {
	return nconsts.CreateDatasetID
}

func UnmarshalCreateDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result CreateDatasetResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.DatasetAddress = p.UnpackString(true)
	result.DatasetParentNftAddress = p.UnpackString(true)
	return &result, p.Err()
}
