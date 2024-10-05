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
	CreateAssetComputeUnits = 15
)

var (
	ErrOutputAssetTypeInvalid              = errors.New("asset type is invalid")
	ErrNameInvalid                         = errors.New("name is invalid")
	ErrOutputSymbolInvalid                 = errors.New("symbol is invalid")
	ErrOutputDecimalsInvalid               = errors.New("decimals is invalid")
	ErrMetadataInvalid                     = errors.New("metadata is invalid")
	ErrOutputURIInvalid                    = errors.New("uri is invalid")
	ErrAssetAlreadyExists                  = errors.New("asset already exists")
	_                         chain.Action = (*CreateAsset)(nil)
)

type CreateAsset struct {
	// Asset type
	AssetType uint8 `serialize:"true" json:"asset_type"`

	// The name of the asset
	Name string `serialize:"true" json:"name"`

	// The symbol of the asset
	Symbol string `serialize:"true" json:"symbol"`

	// The number of decimal places in the asset
	Decimals uint8 `serialize:"true" json:"decimals"`

	// The metadata of the asset
	Metadata string `serialize:"true" json:"metadata"`

	// URI for the asset
	URI string `serialize:"true" json:"uri"`

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

func (c *CreateAsset) StateKeys(actor codec.Address) state.Keys {
	assetAddress := storage.AssetAddress(c.AssetType, []byte(c.Name), []byte(c.Symbol), c.Decimals, []byte(c.Metadata), []byte(c.URI), actor)
	stateKeys := state.Keys{
		string(storage.AssetInfoKey(assetAddress)):                  state.All,
		string(storage.AssetAccountBalanceKey(assetAddress, actor)): state.Allocate | state.Write,
	}

	// Check if c.AssetType is a dataset type so we
	// can create the NFT ID
	if c.AssetType == nconsts.AssetDatasetTokenID {
		nftAddress := storage.AssetNFTAddress(assetAddress, 0)
		stateKeys[string(storage.AssetAccountBalanceKey(nftAddress, actor))] = state.Allocate | state.Write
	}
	return stateKeys
}

func (c *CreateAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	assetAddress := storage.AssetAddress(c.AssetType, []byte(c.Name), []byte(c.Symbol), c.Decimals, []byte(c.Metadata), []byte(c.URI), actor)

	if c.AssetType != nconsts.AssetFungibleTokenID && c.AssetType != nconsts.AssetNonFungibleTokenID && c.AssetType != nconsts.AssetDatasetTokenID {
		return nil, ErrOutputAssetTypeInvalid
	}
	if len(c.Name) < 3 || len(c.Name) > MaxMetadataSize {
		return nil, ErrNameInvalid
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
	if len(c.Metadata) > MaxMetadataSize {
		return nil, ErrMetadataInvalid
	}
	if len(c.URI) > MaxMetadataSize {
		return nil, ErrOutputURIInvalid
	}
	mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin := codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress
	if c.MintAdmin != mintAdmin {
		mintAdmin = c.MintAdmin
	}
	if c.PauseUnpauseAdmin != pauseUnpauseAdmin {
		pauseUnpauseAdmin = c.PauseUnpauseAdmin
	}
	if c.FreezeUnfreezeAdmin != freezeUnfreezeAdmin {
		freezeUnfreezeAdmin = c.FreezeUnfreezeAdmin
	}
	if c.EnableDisableKYCAccountAdmin != enableDisableKYCAccountAdmin {
		enableDisableKYCAccountAdmin = c.EnableDisableKYCAccountAdmin
	}

	exists, _, _, _, _, _, _, _, _, _, _, _, _, _, _ := storage.GetAssetInfoNoController(ctx, mu, assetAddress)
	if exists {
		return nil, ErrAssetAlreadyExists
	}

	output := CreateAssetResult{
		AssetID:      assetAddress,
		AssetBalance: uint64(0),
	}
	totalSupply := uint64(0)
	if c.AssetType == nconsts.AssetDatasetTokenID {
		// Mint the parent NFT for the dataset(fractionalized asset)
		nftID := utils.GenerateIDWithIndex(assetAddress, 0)
		output.DatasetParentNftID = nftID
		if err := storage.SetAssetNFT(ctx, mu, assetAddress, 0, nftID, []byte(c.URI), []byte(c.Metadata), actor); err != nil {
			return nil, err
		}
		amountOfToken := uint64(1)
		totalSupply += amountOfToken
		output.AssetBalance = amountOfToken
		// Add the balance to NFT collection
		if _, err := storage.AddBalance(ctx, mu, actor, assetAddress, amountOfToken, true); err != nil {
			return nil, err
		}

		// Add the balance to individual NFT
		if _, err := storage.AddBalance(ctx, mu, actor, nftID, amountOfToken, true); err != nil {
			return nil, err
		}
	}

	if err := storage.SetAssetInfo(ctx, mu, assetAddress, c.AssetType, []byte(c.Name), []byte(c.Symbol), c.Decimals, []byte(c.Metadata), []byte(c.URI), totalSupply, c.MaxSupply, actor, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	return &output, nil
}

func (*CreateAsset) ComputeUnits(chain.Rules) uint64 {
	return CreateAssetComputeUnits
}

func (*CreateAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalCreateAsset(p *codec.Packer) (chain.Action, error) {
	var create CreateAsset
	create.AssetID = p.UnpackString(true)
	create.AssetType = p.UnpackByte()
	create.Name = p.UnpackString(true)
	create.Symbol = p.UnpackString(true)
	create.Decimals = p.UnpackByte()
	create.Metadata = p.UnpackString(false)
	create.URI = p.UnpackString(false)
	create.MaxSupply = p.UnpackUint64(false)
	p.UnpackAddress(&create.MintAdmin)
	p.UnpackAddress(&create.PauseUnpauseAdmin)
	p.UnpackAddress(&create.FreezeUnfreezeAdmin)
	p.UnpackAddress(&create.EnableDisableKYCAccountAdmin)
	return &create, p.Err()
}

var (
	_ codec.Typed     = (*CreateAssetResult)(nil)
	_ chain.Marshaler = (*CreateAssetResult)(nil)
)

type CreateAssetResult struct {
	AssetID            ids.ID `serialize:"true" json:"asset_id"`
	AssetBalance       uint64 `serialize:"true" json:"asset_balance"`
	DatasetParentNftID ids.ID `serialize:"true" json:"nft_id"`
}

func (*CreateAssetResult) GetTypeID() uint8 {
	return nconsts.CreateAssetID
}

func (*CreateAssetResult) Size() int {
	return ids.IDLen + consts.Uint64Len + ids.IDLen
}

func (r *CreateAssetResult) Marshal(p *codec.Packer) {
	p.PackID(r.AssetID)
	p.PackUint64(r.AssetBalance)
	p.PackID(r.DatasetParentNftID)
}

func UnmarshalCreateAssetResult(p *codec.Packer) (codec.Typed, error) {
	var result CreateAssetResult
	p.UnpackID(true, &result.AssetID)
	result.AssetBalance = p.UnpackUint64(false)
	p.UnpackID(false, &result.DatasetParentNftID)
	return &result, p.Err()
}
