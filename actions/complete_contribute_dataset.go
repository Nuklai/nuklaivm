// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/marketplace"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	CompleteContributeDatasetComputeUnits = 5
)

var _ chain.Action = (*CompleteContributeDataset)(nil)

type CompleteContributeDataset struct {
	// DatasetID ID
	DatasetID ids.ID `serialize:"true" json:"dataset_id"`

	// Contributor
	Contributor codec.Address `serialize:"true" json:"contributor"`

	// Unique NFT ID to assign to the NFT
	UniqueNFTIDForContributor uint64 `serialize:"true" json:"unique_nft_id_for_contributor"`
}

func (*CompleteContributeDataset) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (d *CompleteContributeDataset) StateKeys(_ codec.Address) state.Keys {
	nftID := utils.GenerateIDWithIndex(d.DatasetID, d.UniqueNFTIDForContributor)
	return state.Keys{
		string(storage.AssetInfoKey(d.DatasetID)):              state.Read | state.Write,
		string(storage.AssetNFTKey(nftID)):                     state.Allocate | state.Write,
		string(storage.DatasetInfoKey(d.DatasetID)):            state.Read,
		string(storage.BalanceKey(d.Contributor, ids.Empty)):   state.Read | state.Write,
		string(storage.BalanceKey(d.Contributor, d.DatasetID)): state.Allocate | state.Write,
		string(storage.BalanceKey(d.Contributor, nftID)):       state.Allocate | state.Write,
	}
}

func (d *CompleteContributeDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the dataset exists
	exists, _, description, _, _, _, _, _, _, saleID, _, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, mu, d.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}
	if actor != owner {
		return nil, ErrWrongOwner
	}

	// Check if the dataset is already on sale
	if saleID != ids.Empty {
		return nil, ErrDatasetAlreadyOnSale
	}

	// Check if the nftID already exists
	nftID := utils.GenerateIDWithIndex(d.DatasetID, d.UniqueNFTIDForContributor)
	exists, _, _, _, _, _, _ = storage.GetAssetNFT(ctx, mu, nftID)
	if exists {
		return nil, ErrNFTAlreadyExists
	}

	// Retrieve the asset info
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAssetInfoNoController(ctx, mu, d.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssetNotFound
	}
	if actor != mintActor {
		return nil, ErrWrongMintAdmin
	}

	// Ensure that total supply is less than max supply
	amountOfToken := uint64(1)
	newSupply, err := smath.Add(totalSupply, amountOfToken)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}
	totalSupply = newSupply

	// Get the marketplace instance
	marketplaceInstance := marketplace.GetMarketplace()
	dataContribution, err := marketplaceInstance.CompleteContributeDataset(d.DatasetID, d.Contributor)
	if err != nil {
		return nil, err
	}

	// Mint the child NFT for the dataset(fractionalized asset)
	metadataNFTMap := make(map[string]string, 0)
	metadataNFTMap["dataLocation"] = string(dataContribution.DataLocation)
	metadataNFTMap["dataIdentifier"] = string(dataContribution.DataIdentifier)
	metadataNFT, err := utils.MapToBytes(metadataNFTMap)
	if err != nil {
		return nil, err
	}
	if err := storage.SetAssetNFT(ctx, mu, d.DatasetID, d.UniqueNFTIDForContributor, nftID, description, metadataNFT, d.Contributor); err != nil {
		return nil, err
	}

	// Update asset with new total supply
	if err := storage.SetAssetInfo(ctx, mu, d.DatasetID, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	if _, err := storage.AddBalance(ctx, mu, d.Contributor, d.DatasetID, 1, true); err != nil {
		return nil, err
	}
	// Add the balance to individual NFT
	if _, err := storage.AddBalance(ctx, mu, d.Contributor, nftID, 1, true); err != nil {
		return nil, err
	}

	// Refund the collateral back to the contributor
	dataConfig := marketplace.GetDatasetConfig()
	if _, err := storage.AddBalance(ctx, mu, d.Contributor, dataConfig.CollateralAssetIDForDataContribution, dataConfig.CollateralAmountForDataContribution, true); err != nil {
		return nil, err
	}

	return &CompleteContributeDatasetResult{
		CollateralAssetID:        dataConfig.CollateralAssetIDForDataContribution,
		CollateralAmountRefunded: dataConfig.CollateralAmountForDataContribution,
		DatasetID:                d.DatasetID,
		DatasetChildNftID:        nftID,
		To:                       d.Contributor,
		DataLocation:             dataContribution.DataLocation,
		DataIdentifier:           dataContribution.DataIdentifier,
	}, nil
}

func (*CompleteContributeDataset) ComputeUnits(chain.Rules) uint64 {
	return CompleteContributeDatasetComputeUnits
}

func (*CompleteContributeDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*CompleteContributeDataset)(nil)

func (*CompleteContributeDataset) Size() int {
	return ids.IDLen + codec.AddressLen + consts.Uint64Len
}

func (d *CompleteContributeDataset) Marshal(p *codec.Packer) {
	p.PackID(d.DatasetID)
	p.PackAddress(d.Contributor)
	p.PackUint64(d.UniqueNFTIDForContributor)
}

func UnmarshalCompleteContributeDataset(p *codec.Packer) (chain.Action, error) {
	var complete CompleteContributeDataset
	p.UnpackID(true, &complete.DatasetID)
	p.UnpackAddress(&complete.Contributor)
	complete.UniqueNFTIDForContributor = p.UnpackUint64(true)
	return &complete, p.Err()
}

var (
	_ codec.Typed     = (*CompleteContributeDatasetResult)(nil)
	_ chain.Marshaler = (*CompleteContributeDatasetResult)(nil)
)

type CompleteContributeDatasetResult struct {
	CollateralAssetID        ids.ID        `serialize:"true" json:"collateral_asset_id"`
	CollateralAmountRefunded uint64        `serialize:"true" json:"collateral_amount_refunded"`
	DatasetID                ids.ID        `serialize:"true" json:"dataset_id"`
	DatasetChildNftID        ids.ID        `serialize:"true" json:"dataset_child_nft_id"`
	To                       codec.Address `serialize:"true" json:"to"`
	DataLocation             []byte        `serialize:"true" json:"data_location"`
	DataIdentifier           []byte        `serialize:"true" json:"data_identifier"`
}

func (*CompleteContributeDatasetResult) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (r *CompleteContributeDatasetResult) Size() int {
	return ids.IDLen*3 + consts.Uint64Len + codec.AddressLen + codec.BytesLen(r.DataLocation) + codec.BytesLen(r.DataIdentifier)
}

func (r *CompleteContributeDatasetResult) Marshal(p *codec.Packer) {
	p.PackID(r.CollateralAssetID)
	p.PackUint64(r.CollateralAmountRefunded)
	p.PackID(r.DatasetID)
	p.PackID(r.DatasetChildNftID)
	p.PackAddress(r.To)
	p.PackBytes(r.DataLocation)
	p.PackBytes(r.DataIdentifier)
}

func UnmarshalCompleteContributeDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result CompleteContributeDatasetResult
	p.UnpackID(false, &result.CollateralAssetID)
	result.CollateralAmountRefunded = p.UnpackUint64(true)
	p.UnpackID(true, &result.DatasetID)
	p.UnpackID(true, &result.DatasetChildNftID)
	p.UnpackAddress(&result.To)
	p.UnpackBytes(MaxTextSize, true, &result.DataLocation)
	p.UnpackBytes(MaxMetadataSize-MaxTextSize, true, &result.DataIdentifier)
	return &result, p.Err()
}
