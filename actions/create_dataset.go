// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

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
	CreateDatasetComputeUnits = 100
)

var (
	ErrDescriptionInvalid                = errors.New("description is invalid")
	ErrCategoriesInvalid                 = errors.New("categories is invalid")
	ErrLicenseNameInvalid                = errors.New("license name is invalid")
	ErrLicenseSymbolInvalid              = errors.New("license symbol is invalid")
	ErrLicenseURLInvalid                 = errors.New("license url is invalid")
	_                       chain.Action = (*CreateDataset)(nil)
)

type CreateDataset struct {
	// Asset id if it was already created
	AssetID ids.ID `serialize:"true" json:"asset_id"`

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

	// Metadata of the dataset
	Metadata []byte `serialize:"true" json:"metadata"`

	// False for sole contributor and true for open contribution
	IsCommunityDataset bool `serialize:"true" json:"is_community_dataset"`
}

func (*CreateDataset) GetTypeID() uint8 {
	return nconsts.CreateDatasetID
}

func (c *CreateDataset) StateKeys(actor codec.Address) state.Keys {
	/* 	assetID := actionID
	   	if c.AssetID != ids.Empty {
	   		assetID = c.AssetID
	   	}
	   	nftID := utils.GenerateIDWithIndex(actionID, 0) */
	// TODO: Remove after hypersdk adds pseudorandom actionID generation
	assetID := c.AssetID
	return state.Keys{
		// string(storage.AssetKey(assetID)):          state.Allocate | state.Write,
		string(storage.AssetInfoKey(assetID)):   state.All,
		string(storage.DatasetInfoKey(assetID)): state.Allocate | state.Write,
		// string(storage.AssetNFTKey(nftID)):         state.Allocate | state.Write,
		string(storage.BalanceKey(actor, assetID)): state.Allocate | state.Write,
		// string(storage.BalanceKey(actor, nftID)):   state.Allocate | state.Write,
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
	if len(c.Name) < 3 || len(c.Name) > MaxMetadataSize {
		return nil, ErrNameInvalid
	}
	if len(c.Description) < 3 || len(c.Description) > MaxMetadataSize {
		return nil, ErrDescriptionInvalid
	}
	if len(c.Categories) < 3 || len(c.Categories) > MaxMetadataSize {
		return nil, ErrCategoriesInvalid
	}
	if len(c.LicenseName) < 3 || len(c.LicenseName) > MaxMetadataSize {
		return nil, ErrLicenseNameInvalid
	}
	if len(c.LicenseSymbol) < 3 || len(c.LicenseSymbol) > MaxTextSize {
		return nil, ErrLicenseSymbolInvalid
	}
	if len(c.LicenseURL) < 3 || len(c.LicenseURL) > MaxMetadataSize {
		return nil, ErrLicenseURLInvalid
	}
	if len(c.Metadata) < 3 || len(c.Metadata) > MaxDatasetMetadataSize {
		return nil, ErrMetadataInvalid
	}

	assetID := c.AssetID
	// TODO: Remove after hypersdk adds pseudorandom actionID generation
	// Check if the asset exists
	exists, assetType, _, _, _, _, _, _, _, _, mintActor, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, assetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetDatasetTokenID {
		return nil, ErrOutputWrongAssetType
	}
	if mintActor != actor {
		return nil, ErrOutputWrongMintAdmin
	}
	/*
		var assetID ids.ID
		if c.AssetID != ids.Empty {
			assetID = c.AssetID
			// Check if the asset exists
			exists, assetType, _, _, _, _, _, _, _, _, mintActor, _, _, _, err := storage.GetAsset(ctx, mu, assetID)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, ErrOutputAssetMissing
			}
			if assetType != nconsts.AssetDatasetTokenID {
				return nil, ErrOutputWrongAssetType
			}
			if mintActor != actor {
				return nil, ErrOutputWrongMintAdmin
			}
		} 	else {
			assetID = actionID

			// Mint the parent NFT for the dataset(fractionalized asset)
			nftID := utils.GenerateIDWithIndex(assetID, 0)
			if err := storage.SetAssetNFT(ctx, mu, assetID, 0, nftID, c.Description, c.Description, actor); err != nil {
				return nil, err
			}

			// Create a new asset for the dataset
			symbol := utils.CombineWithPrefix([]byte(""), c.Name, MaxTextSize)
			if err := storage.SetAsset(ctx, mu, assetID, nconsts.AssetDatasetTokenID, c.Name, symbol, 0, c.Description, c.Description, 1, 0, actor, actor, actor, actor, actor); err != nil {
				return nil, err
			}

			// Add the balance to NFT collection
			if _, err := storage.AddBalance(ctx, mu, actor, assetID, 1, true); err != nil {
				return nil, err
			}

			// Add the balance to individual NFT
			if _, err := storage.AddBalance(ctx, mu, actor, nftID, 1, true); err != nil {
				return nil, err
			}
		} */

	revenueModelDataShare, revenueModelDataOwnerCut := 100, 100
	if c.IsCommunityDataset {
		revenueModelDataOwnerCut = 10
	}
	// Create a new dataset with the following parameters:
	// saleID = ids.Empty
	// baseAsset = ids.Empty
	// basePrice = 0
	// revenueModelDataShare = 100
	// revenueModelMetadataShare = 0
	// revenueModelDataOwnerCut = 10 for community datasets, 100 for sole contributor datasets
	// revenueModelMetadataOwnerCut = 0
	if err := storage.SetDatasetInfo(ctx, mu, assetID, c.Name, c.Description, c.Categories, c.LicenseName, c.LicenseSymbol, c.LicenseURL, c.Metadata, c.IsCommunityDataset, ids.Empty, ids.Empty, 0, uint8(revenueModelDataShare), 0, uint8(revenueModelDataOwnerCut), 0, actor); err != nil {
		return nil, err
	}

	return &CreateDatasetResult{
		DatasetID:          assetID,
		DatasetParentNftID: utils.GenerateIDWithIndex(assetID, 0),
	}, nil
}

func (*CreateDataset) ComputeUnits(chain.Rules) uint64 {
	return CreateDatasetComputeUnits
}

func (*CreateDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*CreateDataset)(nil)

func (c *CreateDataset) Size() int {
	return ids.IDLen + codec.BytesLen(c.Name) + codec.BytesLen(c.Description) + codec.BytesLen(c.Categories) + codec.BytesLen(c.LicenseName) + codec.BytesLen(c.LicenseSymbol) + codec.BytesLen(c.LicenseURL) + codec.BytesLen(c.Metadata) + consts.BoolLen
}

func (c *CreateDataset) Marshal(p *codec.Packer) {
	p.PackID(c.AssetID)
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
	p.UnpackID(false, &create.AssetID)
	p.UnpackBytes(MaxMetadataSize, true, &create.Name)
	p.UnpackBytes(MaxMetadataSize, true, &create.Description)
	p.UnpackBytes(MaxMetadataSize, true, &create.Categories)
	p.UnpackBytes(MaxMetadataSize, true, &create.LicenseName)
	p.UnpackBytes(MaxTextSize, true, &create.LicenseSymbol)
	p.UnpackBytes(MaxMetadataSize, true, &create.LicenseURL)
	p.UnpackBytes(MaxDatasetMetadataSize, true, &create.Metadata)
	create.IsCommunityDataset = p.UnpackBool()
	return &create, p.Err()
}

var (
	_ codec.Typed     = (*CreateDatasetResult)(nil)
	_ chain.Marshaler = (*CreateDatasetResult)(nil)
)

type CreateDatasetResult struct {
	DatasetID          ids.ID `serialize:"true" json:"dataset_id"`
	DatasetParentNftID ids.ID `serialize:"true" json:"parent_nft_id"`
}

func (*CreateDatasetResult) GetTypeID() uint8 {
	return nconsts.CreateDatasetID
}

func (*CreateDatasetResult) Size() int {
	return ids.IDLen * 2
}

func (r *CreateDatasetResult) Marshal(p *codec.Packer) {
	p.PackID(r.DatasetID)
	p.PackID(r.DatasetParentNftID)
}

func UnmarshalCreateDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result CreateDatasetResult
	p.UnpackID(true, &result.DatasetID)
	p.UnpackID(true, &result.DatasetParentNftID)
	return &result, p.Err()
}
