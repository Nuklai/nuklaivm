// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/marketplace"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*CompleteContributeDataset)(nil)

type CompleteContributeDataset struct {
	// Dataset ID
	Dataset ids.ID `json:"dataset"`

	// Contributor
	Contributor codec.Address `json:"contributor"`

	// Unique NFT ID to assign to the NFT
	UniqueNFTID uint64 `json:"uniqueNFTID"`
}

func (*CompleteContributeDataset) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (d *CompleteContributeDataset) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	nftID := nchain.GenerateID(d.Dataset, d.UniqueNFTID)
	return state.Keys{
		string(storage.AssetKey(d.Dataset)):                  state.Read | state.Write,
		string(storage.AssetNFTKey(nftID)):                   state.Allocate | state.Write,
		string(storage.DatasetKey(d.Dataset)):                state.Read,
		string(storage.BalanceKey(d.Contributor, ids.Empty)): state.Read | state.Write,
		string(storage.BalanceKey(d.Contributor, d.Dataset)): state.Allocate | state.Write,
		string(storage.BalanceKey(d.Contributor, nftID)):     state.Allocate | state.Write,
	}
}

func (*CompleteContributeDataset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.DatasetChunks, storage.BalanceChunks}
}

func (d *CompleteContributeDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if the dataset exists
	exists, _, description, _, _, _, _, _, _, _, _, _, _, _, _, _, owner, err := storage.GetDataset(ctx, mu, d.Dataset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}
	if actor != owner {
		return nil, ErrNotDatasetOwner
	}

	// Check if the nftID already exists
	nftID := nchain.GenerateID(d.Dataset, d.UniqueNFTID)
	exists, _, _, _, _, _, _ = storage.GetAssetNFT(ctx, mu, nftID)
	if exists {
		return nil, ErrOutputNFTAlreadyExists
	}

	// Retrieve the asset info
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, d.Dataset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssetNotFound
	}
	if actor != mintActor {
		return nil, ErrOutputWrongMintActor
	}

	// Ensure that total supply is less than max supply
	amountOfToken := uint64(1)
	newSupply, err := smath.Add64(totalSupply, amountOfToken)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}
	totalSupply = newSupply

	// Get the marketplace instance
	marketplaceInstance := marketplace.GetMarketplace()
	dataContribution, err := marketplaceInstance.CompleteContributeDataset(d.Dataset, d.Contributor)
	if err != nil {
		return nil, err
	}

	// Mint the child NFT for the dataset(fractionalized asset)
	metadataNFT := []byte("{\"dataLocation\":\"" + string(dataContribution.DataLocation) + "\",\"dataIdentifier\":\"" + string(dataContribution.DataIdentifier) + "\"}")
	if err := storage.SetAssetNFT(ctx, mu, d.Dataset, d.UniqueNFTID, nftID, description, metadataNFT, d.Contributor); err != nil {
		return nil, err
	}

	// Update asset with new total supply
	if err := storage.SetAsset(ctx, mu, d.Dataset, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	if err := storage.AddBalance(ctx, mu, d.Contributor, d.Dataset, 1, true); err != nil {
		return nil, err
	}
	// Add the balance to individual NFT
	if err := storage.AddBalance(ctx, mu, d.Contributor, nftID, 1, true); err != nil {
		return nil, err
	}

	// Refund the collateral back to the contributor
	dataConfig := marketplace.GetDatasetConfig()
	if err := storage.AddBalance(ctx, mu, d.Contributor, ids.Empty, dataConfig.CollateralForDataContribution, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*CompleteContributeDataset) ComputeUnits(chain.Rules) uint64 {
	return CompleteContributeDatasetComputeUnits
}

func (*CompleteContributeDataset) Size() int {
	return ids.IDLen + codec.AddressLen + consts.Uint64Len
}

func (d *CompleteContributeDataset) Marshal(p *codec.Packer) {
	p.PackID(d.Dataset)
	p.PackAddress(d.Contributor)
	p.PackUint64(d.UniqueNFTID)
}

func UnmarshalCompleteContributeDataset(p *codec.Packer) (chain.Action, error) {
	var complete CompleteContributeDataset
	p.UnpackID(true, &complete.Dataset)
	p.UnpackAddress(&complete.Contributor)
	complete.UniqueNFTID = p.UnpackUint64(true)
	return &complete, p.Err()
}

func (*CompleteContributeDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
