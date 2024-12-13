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
	ErrAssetExists                   = errors.New("asset already exists")
	ErrAssetTypeInvalid              = errors.New("asset type is invalid")
	ErrNameInvalid                   = errors.New("name is invalid")
	ErrSymbolInvalid                 = errors.New("symbol is invalid")
	ErrDecimalsInvalid               = errors.New("decimals is invalid")
	ErrMetadataInvalid               = errors.New("metadata is invalid")
	ErrURIInvalid                    = errors.New("uri is invalid")
	_                   chain.Action = (*CreateAsset)(nil)
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
	assetAddress := storage.AssetAddress(c.AssetType, []byte(c.Name), []byte(c.Symbol), c.Decimals, []byte(c.Metadata), actor)
	stateKeys := state.Keys{
		string(storage.AssetInfoKey(assetAddress)):                  state.All,
		string(storage.AssetAccountBalanceKey(assetAddress, actor)): state.Allocate | state.Write,
	}

	// Check if c.AssetType is a dataset type so we
	// can create the NFT ID
	if c.AssetType == nconsts.AssetFractionalTokenID {
		nftAddress := storage.AssetAddressNFT(assetAddress, []byte(c.Metadata), actor)
		stateKeys[string(storage.AssetInfoKey(nftAddress))] = state.All
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
	assetAddress := storage.AssetAddress(c.AssetType, []byte(c.Name), []byte(c.Symbol), c.Decimals, []byte(c.Metadata), actor)

	if c.AssetType != nconsts.AssetFungibleTokenID && c.AssetType != nconsts.AssetNonFungibleTokenID && c.AssetType != nconsts.AssetFractionalTokenID {
		return nil, ErrAssetTypeInvalid
	}
	if len(c.Name) < 3 || len(c.Name) > storage.MaxNameSize {
		return nil, ErrNameInvalid
	}
	if len(c.Symbol) < 3 || len(c.Symbol) > storage.MaxSymbolSize {
		return nil, ErrSymbolInvalid
	}
	if c.AssetType == nconsts.AssetFungibleTokenID && c.Decimals > storage.MaxAssetDecimals {
		return nil, ErrDecimalsInvalid
	}
	if c.AssetType != nconsts.AssetFungibleTokenID && c.Decimals != 0 {
		return nil, ErrDecimalsInvalid
	}
	if len(c.Metadata) > storage.MaxAssetMetadataSize {
		return nil, ErrMetadataInvalid
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

	// Continue only if asset doesn't exist
	if storage.AssetExists(ctx, mu, assetAddress) {
		return nil, ErrAssetExists
	}

	output := CreateAssetResult{
		CommonResult: FillCommonResult(actor.String(), ""),
		AssetAddress: assetAddress.String(),
		AssetBalance: uint64(0),
	}

	// Create the asset
	if err := storage.SetAssetInfo(ctx, mu, assetAddress, c.AssetType, []byte(c.Name), []byte(c.Symbol), c.Decimals, []byte(c.Metadata), []byte(assetAddress.String()), 0, c.MaxSupply, actor, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	if c.AssetType == nconsts.AssetFractionalTokenID {
		amountOfToken := uint64(1)
		// Add to NFT collection too
		if _, err := storage.MintAsset(ctx, mu, assetAddress, actor, amountOfToken); err != nil {
			return nil, err
		}
		// Mint the parent NFT for the dataset(fractionalized asset)
		nftAddress := storage.AssetAddressNFT(assetAddress, []byte(c.Metadata), actor)
		output.DatasetParentNftAddress = nftAddress.String()
		output.CommonResult.Receiver = actor.String()
		symbol := utils.CombineWithSuffix([]byte(c.Symbol), 0, storage.MaxSymbolSize)
		if err := storage.SetAssetInfo(ctx, mu, nftAddress, nconsts.AssetNonFungibleTokenID, []byte(c.Name), symbol, 0, []byte(c.Metadata), []byte(assetAddress.String()), 0, 1, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
			return nil, err
		}
		newBalance, err := storage.MintAsset(ctx, mu, nftAddress, actor, amountOfToken)
		if err != nil {
			return nil, err
		}
		output.AssetBalance = newBalance
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
	create.AssetType = p.UnpackByte()
	create.Name = p.UnpackString(true)
	create.Symbol = p.UnpackString(true)
	create.Decimals = p.UnpackByte()
	create.Metadata = p.UnpackString(false)
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
	CommonResult
	AssetAddress            string `serialize:"true" json:"asset_address"`
	AssetBalance            uint64 `serialize:"true" json:"asset_balance"`
	DatasetParentNftAddress string `serialize:"true" json:"dataset_parent_nft_address"`
}

func (*CreateAssetResult) GetTypeID() uint8 {
	return nconsts.CreateAssetID
}

func (r *CreateAssetResult) Size() int {
	return codec.StringLen(r.AssetAddress) + consts.Uint64Len + codec.StringLen(r.DatasetParentNftAddress)
}

func (r *CreateAssetResult) Marshal(p *codec.Packer) {
	p.PackString(r.AssetAddress)
	p.PackUint64(r.AssetBalance)
	p.PackString(r.DatasetParentNftAddress)
}

func UnmarshalCreateAssetResult(p *codec.Packer) (codec.Typed, error) {
	var result CreateAssetResult
	result.AssetAddress = p.UnpackString(true)
	result.AssetBalance = p.UnpackUint64(false)
	result.DatasetParentNftAddress = p.UnpackString(false)
	return &result, p.Err()
}
