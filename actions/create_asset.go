// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	CreateAssetComputeUnits = 15
)

var (
	ErrOutputAssetTypeInvalid              = errors.New("asset type is invalid")
	ErrOutputNameInvalid                   = errors.New("name is invalid")
	ErrOutputSymbolInvalid                 = errors.New("symbol is invalid")
	ErrOutputDecimalsInvalid               = errors.New("decimals is invalid")
	ErrOutputMetadataInvalid               = errors.New("metadata is invalid")
	ErrOutputURIInvalid                    = errors.New("uri is invalid")
	_                         chain.Action = (*CreateAsset)(nil)
)

type CreateAsset struct {
	// Asset type
	AssetType uint8 `serialize:"true" json:"asset_type"`

	// The name of the asset
	Name []byte `serialize:"true" json:"name"`

	// The symbol of the asset
	Symbol []byte `serialize:"true" json:"symbol"`

	// The number of decimal places in the asset
	Decimals uint8 `serialize:"true" json:"decimals"`

	// The metadata of the asset
	Metadata []byte `serialize:"true" json:"metadata"`

	// URI for the asset
	URI []byte `serialize:"true" json:"uri"`

	// The max supply of the asset
	MaxSupply uint64 `serialize:"true" json:"max_supply"`

	// The wallet address that can mint assets
	MintAdmin codec.Address `serialize:"true" json:"mint_admin"`

	// The wallet address that can pause/unpause assets
	PauseUnpauseAdmin codec.Address `serialize:"true" json:"pause_unpause_admin"`

	// The wallet address that can freeze/unfreeze assets
	FreezeUnfreezeAdmin codec.Address `serialize:"true" json:"freeze_unfreeze_admin"`

	// The wallet address that can enable/disable KYC account flag
	EnableDisableKYCAccountAdmin codec.Address `serialize:"true" json:"enable_disable_kyc_account_admin"`
}

func (*CreateAsset) GetTypeID() uint8 {
	return nconsts.CreateAssetID
}

func (c *CreateAsset) StateKeys(actor codec.Address, actionID ids.ID) state.Keys {
	// Initialize the base stateKeys map
	stateKeys := state.Keys{
		string(storage.BalanceKey(actor, actionID)): state.Allocate | state.Write,
		string(storage.AssetKey(actionID)):          state.Allocate | state.Write,
	}

	// Check if c.AssetType is a non-fungible type or dataset type so we
	// can create the NFT ID
	if c.AssetType == nconsts.AssetNonFungibleTokenID || c.AssetType == nconsts.AssetDatasetTokenID {
		nftID := nchain.GenerateIDWithIndex(actionID, 0)
		stateKeys[string(storage.BalanceKey(actor, nftID))] = state.Allocate | state.Write
		stateKeys[string(storage.AssetNFTKey(nftID))] = state.Allocate | state.Write
	}
	return stateKeys
}

func (c *CreateAsset) StateKeysMaxChunks() []uint16 {
	stateKeysChunks := make([]uint16, 0)
	stateKeysChunks = append(stateKeysChunks, storage.BalanceChunks)
	stateKeysChunks = append(stateKeysChunks, storage.AssetChunks)
	if c.AssetType == nconsts.AssetNonFungibleTokenID || c.AssetType == nconsts.AssetDatasetTokenID {
		stateKeysChunks = append(stateKeysChunks, storage.BalanceChunks)
		stateKeysChunks = append(stateKeysChunks, storage.AssetNFTChunks)
	}
	return stateKeysChunks
}

func (c *CreateAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	actionID ids.ID,
) (codec.Typed, error) {
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
	mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin := codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress
	if decoded, err := json.Marshal(c.MintAdmin); err == nil && len(decoded) == codec.AddressLen {
		mintAdmin = c.MintAdmin
	}
	if decoded, err := json.Marshal(c.PauseUnpauseAdmin); err == nil && len(decoded) == codec.AddressLen {
		pauseUnpauseAdmin = c.PauseUnpauseAdmin
	}
	if decoded, err := json.Marshal(c.FreezeUnfreezeAdmin); err == nil && len(decoded) == codec.AddressLen {
		freezeUnfreezeAdmin = c.FreezeUnfreezeAdmin
	}
	if decoded, err := json.Marshal(c.EnableDisableKYCAccountAdmin); err == nil && len(decoded) == codec.AddressLen {
		enableDisableKYCAccountAdmin = c.EnableDisableKYCAccountAdmin
	}

	totalSupply := uint64(0)
	assetBalanceResult := uint64(0)
	nftIDResult := ""
	if c.AssetType == nconsts.AssetDatasetTokenID {
		// Mint the parent NFT for the dataset(fractionalized asset)
		nftID := nchain.GenerateIDWithIndex(actionID, 0)
		nftIDResult = nftID.String()
		if err := storage.SetAssetNFT(ctx, mu, actionID, 0, nftID, c.URI, c.Metadata, actor); err != nil {
			return nil, err
		}
		amountOfToken := uint64(1)
		totalSupply += amountOfToken
		assetBalanceResult = amountOfToken
		// Add the balance to NFT collection
		if _, err := storage.AddBalance(ctx, mu, actor, actionID, amountOfToken, true); err != nil {
			return nil, err
		}

		// Add the balance to individual NFT
		if _, err := storage.AddBalance(ctx, mu, actor, nftID, amountOfToken, true); err != nil {
			return nil, err
		}
	}

	if err := storage.SetAsset(ctx, mu, actionID, c.AssetType, c.Name, c.Symbol, c.Decimals, c.Metadata, c.URI, totalSupply, c.MaxSupply, actor, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	return &CreateAssetResult{
		AssetID:      actionID.String(),
		AssetBalance: assetBalanceResult,
		NftID:        nftIDResult,
	}, nil
}

func (*CreateAsset) ComputeUnits(chain.Rules) uint64 {
	return CreateAssetComputeUnits
}

func (*CreateAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*CreateAsset)(nil)

func (c *CreateAsset) Size() int {
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
	p.PackAddress(c.MintAdmin)
	p.PackAddress(c.PauseUnpauseAdmin)
	p.PackAddress(c.FreezeUnfreezeAdmin)
	p.PackAddress(c.EnableDisableKYCAccountAdmin)
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
	p.UnpackAddress(&create.MintAdmin)
	p.UnpackAddress(&create.PauseUnpauseAdmin)
	p.UnpackAddress(&create.FreezeUnfreezeAdmin)
	p.UnpackAddress(&create.EnableDisableKYCAccountAdmin)
	return &create, p.Err()
}

var _ codec.Typed = (*CreateAssetResult)(nil)

type CreateAssetResult struct {
	AssetID      string `serialize:"true" json:"asset_id"`
	AssetBalance uint64 `serialize:"true" json:"asset_balance"`
	NftID        string `serialize:"true" json:"nft_id"`
}

func (*CreateAssetResult) GetTypeID() uint8 {
	return nconsts.CreateAssetID
}
