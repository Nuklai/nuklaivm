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

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*CreateAsset)(nil)

type CreateAsset struct {
	// Asset type
	AssetType uint8 `json:"assetType"`

	// The name of the asset
	Name []byte `json:"name"`

	// The symbol of the asset
	Symbol []byte `json:"symbol"`

	// The number of decimal places in the asset
	Decimals uint8 `json:"decimals"`

	// The metadata of the asset
	Metadata []byte `json:"metadata"`

	// URI for the asset
	URI []byte `json:"uri"`

	// The max supply of the asset
	MaxSupply uint64 `json:"maxSupply"`

	// The wallet address that can mint assets
	MintActor codec.Address `json:"mintActor"`

	// The wallet address that can pause/unpause assets
	PauseUnpauseActor codec.Address `json:"pauseUnpauseActor"`

	// The wallet address that can freeze/unfreeze assets
	FreezeUnfreezeActor codec.Address `json:"freezeUnfreezeActor"`

	// The wallet address that can enable/disable KYC account flag
	EnableDisableKYCAccountActor codec.Address `json:"enableDisableKYCAccountActor"`
}

func (*CreateAsset) GetTypeID() uint8 {
	return nconsts.CreateAssetID
}

func (*CreateAsset) StateKeys(actor codec.Address, actionID ids.ID) state.Keys {
	nftID := nchain.GenerateIDWithIndex(actionID, 0)
	return state.Keys{
		string(storage.AssetKey(actionID)): state.Allocate | state.Write,
		string(storage.AssetNFTKey(nftID)): state.Allocate | state.Write,
		/* string(storage.AssetCollectionPageKey(actionID, 0)):   state.Allocate | state.Write,
		string(storage.AssetCollectionPageCountKey(actionID)): state.Allocate | state.Write, */
		string(storage.BalanceKey(actor, actionID)): state.Allocate | state.Write,
		string(storage.BalanceKey(actor, nftID)):    state.Allocate | state.Write,
	}
}

func (*CreateAsset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.AssetNFTChunks, storage.BalanceChunks, storage.BalanceChunks}

	// return []uint16{storage.AssetChunks, storage.AssetNFTChunks, storage.AssetCollectionPageChunks, storage.AssetCollectionPageCountChunks, storage.BalanceChunks}
}

func (c *CreateAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	actionID ids.ID,
) ([][]byte, error) {
	if c.AssetType != nconsts.AssetFungibleTokenID && c.AssetType != nconsts.AssetNonFungibleTokenID && c.AssetType != nconsts.AssetDatasetTokenID {
		return nil, ErrOutputAssetTypeInvalid
	}
	if len(c.Name) < 3 || len(c.Name) > MaxMetadataSize {
		return nil, ErrOutputNameInvalid
	}
	if len(c.Symbol) < 3 || len(c.Symbol) > MaxTextSize {
		return nil, ErrOutputSymbolInvalid
	}
	if c.AssetType == nconsts.AssetFungibleTokenID && c.Decimals > MaxDecimals {
		return nil, ErrOutputDecimalsInvalid
	}
	if c.AssetType != nconsts.AssetFungibleTokenID && c.Decimals != 0 {
		return nil, ErrOutputDecimalsInvalid
	}
	if len(c.Metadata) < 3 || len(c.Metadata) > MaxMetadataSize {
		return nil, ErrOutputMetadataInvalid
	}
	if len(c.URI) < 3 || len(c.URI) > MaxMetadataSize {
		return nil, ErrOutputURIInvalid
	}
	mintActor := codec.EmptyAddress
	if _, err := codec.AddressBech32(nconsts.HRP, c.MintActor); err == nil {
		mintActor = c.MintActor
	}
	pauseUnpauseActor := codec.EmptyAddress
	if _, err := codec.AddressBech32(nconsts.HRP, c.PauseUnpauseActor); err == nil {
		pauseUnpauseActor = c.PauseUnpauseActor
	}
	freezeUnfreezeActor := codec.EmptyAddress
	if _, err := codec.AddressBech32(nconsts.HRP, c.FreezeUnfreezeActor); err == nil {
		freezeUnfreezeActor = c.FreezeUnfreezeActor
	}
	enableDisableKYCAccountActor := codec.EmptyAddress
	if _, err := codec.AddressBech32(nconsts.HRP, c.EnableDisableKYCAccountActor); err == nil {
		enableDisableKYCAccountActor = c.EnableDisableKYCAccountActor
	}

	totalSupply := uint64(0)
	if c.AssetType == nconsts.AssetDatasetTokenID {
		// Mint the parent NFT for the dataset(fractionalized asset)
		nftID := nchain.GenerateIDWithIndex(actionID, 0)
		if err := storage.SetAssetNFT(ctx, mu, actionID, 0, nftID, c.URI, c.Metadata, actor); err != nil {
			return nil, err
		}
		totalSupply += 1
		// Add the balance to NFT collection
		if err := storage.AddBalance(ctx, mu, actor, actionID, 1, true); err != nil {
			return nil, err
		}

		// Add the balance to individual NFT
		if err := storage.AddBalance(ctx, mu, actor, nftID, 1, true); err != nil {
			return nil, err
		}
		/* 		// Add NFT to Collection
		   		if err := storage.AddAssetNFT(ctx, mu, actionID, nftID); err != nil {
		   			return nil, err
		   		} */
	}

	if err := storage.SetAsset(ctx, mu, actionID, c.AssetType, c.Name, c.Symbol, c.Decimals, c.Metadata, c.URI, totalSupply, c.MaxSupply, actor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*CreateAsset) ComputeUnits(chain.Rules) uint64 {
	return CreateAssetComputeUnits
}

func (c *CreateAsset) Size() int {
	// TODO: add small bytes (smaller int prefix)
	return consts.Uint8Len + codec.BytesLen(c.Name) + codec.BytesLen(c.Symbol) + consts.Uint8Len + codec.BytesLen(c.Metadata) + codec.BytesLen(c.URI) + consts.Uint64Len + codec.AddressLen*4
}

func (c *CreateAsset) Marshal(p *codec.Packer) {
	p.PackByte(c.AssetType)
	p.PackBytes(c.Name)
	p.PackBytes(c.Symbol)
	p.PackByte(c.Decimals)
	p.PackBytes(c.Metadata)
	p.PackBytes(c.URI)
	p.PackUint64(c.MaxSupply)
	p.PackAddress(c.MintActor)
	p.PackAddress(c.PauseUnpauseActor)
	p.PackAddress(c.FreezeUnfreezeActor)
	p.PackAddress(c.EnableDisableKYCAccountActor)
}

func UnmarshalCreateAsset(p *codec.Packer) (chain.Action, error) {
	var create CreateAsset
	create.AssetType = p.UnpackByte()
	p.UnpackBytes(MaxMetadataSize, true, &create.Name)
	p.UnpackBytes(MaxTextSize, true, &create.Symbol)
	create.Decimals = p.UnpackByte()
	p.UnpackBytes(MaxMetadataSize, true, &create.Metadata)
	p.UnpackBytes(MaxMetadataSize, true, &create.URI)
	create.MaxSupply = p.UnpackUint64(false)
	p.UnpackAddress(&create.MintActor)
	p.UnpackAddress(&create.PauseUnpauseActor)
	p.UnpackAddress(&create.FreezeUnfreezeActor)
	p.UnpackAddress(&create.EnableDisableKYCAccountActor)
	return &create, p.Err()
}

func (*CreateAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
