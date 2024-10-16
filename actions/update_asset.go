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
	UpdateAssetComputeUnits = 15
)

var (
	ErrWrongOwner                                             = errors.New("wrong owner")
	ErrAssetNotFound                                          = errors.New("asset not found")
	ErrOutputMustUpdateAtLeastOneField                        = errors.New("must update at least one field")
	ErrOutputMaxSupplyInvalid                                 = errors.New("max supply must be greater than or equal to total supply")
	ErrOutputOwnerInvalid                                     = errors.New("owner is invalid")
	ErrOutputMintAdminInvalid                                 = errors.New("mint admin is invalid")
	ErrOutputPauseUnpauseAdminInvalid                         = errors.New("pause/unpause admin is invalid")
	ErrOutputFreezeUnfreezeAdminInvalid                       = errors.New("freeze/unfreeze admin is invalid")
	ErrOutputEnableDisableKYCAccountAdminInvalid              = errors.New("enable/disable KYC account admin is invalid")
	_                                            chain.Action = (*UpdateAsset)(nil)
)

type UpdateAsset struct {
	// AssetAddress to update
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

	// The name of the asset
	Name string `serialize:"true" json:"name"`

	// The symbol of the asset
	Symbol string `serialize:"true" json:"symbol"`

	// The metadata of the asset
	Metadata string `serialize:"true" json:"metadata"`

	// URI for the asset
	URI string `serialize:"true" json:"uri"`

	// The max supply of the asset
	MaxSupply uint64 `serialize:"true" json:"max_supply"`

	// Owner of the asset
	Owner string `serialize:"true" json:"owner"`

	// The wallet address that can mint/burn assets
	MintAdmin string `serialize:"true" json:"mint_admin"`

	// The wallet address that can pause/unpause assets
	PauseUnpauseAdmin string `serialize:"true" json:"pause_unpause_admin"`

	// The wallet address that can freeze/unfreeze assets
	FreezeUnfreezeAdmin string `serialize:"true" json:"freeze_unfreeze_admin"`

	// The wallet address that can enable/disable KYC account flag
	EnableDisableKYCAccountAdmin string `serialize:"true" json:"enable_disable_kyc_account_admin"`
}

func (*UpdateAsset) GetTypeID() uint8 {
	return nconsts.UpdateAssetID
}

func (u *UpdateAsset) StateKeys(_ codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(u.AssetAddress)): state.Read | state.Write,
	}
}

func (u *UpdateAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the asset exists
	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, mu, u.AssetAddress)
	if err != nil {
		return nil, err
	}
	// Check if the actor is the owner of the asset
	if owner != actor {
		return nil, ErrWrongOwner
	}

	// Note that maxSupply can never be set to 0 on an update.
	// It can only be increased or decreased.
	// If you want to set the max supply to 0, you should set this value
	// to be the max value of a uint64.
	if u.MaxSupply == 0 {
		u.MaxSupply = maxSupply
	} else if u.MaxSupply < totalSupply {
		// Ensure that the max supply is greater than or equal to the total supply
		return nil, ErrOutputMaxSupplyInvalid
	}

	// Ensure that at least one field is being updated
	if (len(u.Name) == 0 || bytes.Equal([]byte(u.Name), name)) && (len(u.Symbol) == 0 || bytes.Equal([]byte(u.Symbol), symbol)) && (len(u.Metadata) == 0 || bytes.Equal([]byte(u.Metadata), metadata)) && (len(u.URI) == 0 || bytes.Equal([]byte(u.URI), uri)) && (u.MaxSupply == maxSupply) && (len(u.Owner) == 0 || u.Owner != owner.String()) && (len(u.MintAdmin) == 0 || u.MintAdmin != mintAdmin.String()) && (len(u.PauseUnpauseAdmin) == 0 || u.PauseUnpauseAdmin != pauseUnpauseAdmin.String()) && (len(u.FreezeUnfreezeAdmin) == 0 || u.FreezeUnfreezeAdmin != freezeUnfreezeAdmin.String()) && (len(u.EnableDisableKYCAccountAdmin) == 0 || u.EnableDisableKYCAccountAdmin != enableDisableKYCAccountAdmin.String()) {
		return nil, ErrOutputMustUpdateAtLeastOneField
	}

	var updateAssetResult UpdateAssetResult

	// if u.Name is passed, update the dataset name
	// otherwise, keep the existing name
	if len(u.Name) > 0 {
		if len(u.Name) < 3 || len(u.Name) > storage.MaxNameSize {
			return nil, ErrNameInvalid
		}
		name = []byte(u.Name)
		updateAssetResult.Name = u.Name
	}

	if len(u.Symbol) > 0 {
		if len(u.Symbol) < 3 || len(u.Symbol) > storage.MaxSymbolSize {
			return nil, ErrSymbolInvalid
		}
		symbol = []byte(u.Symbol)
		updateAssetResult.Symbol = u.Symbol
	}

	if len(u.Metadata) > 0 {
		if len(u.Metadata) < 3 || len(u.Metadata) > storage.MaxDatasetMetadataSize {
			return nil, ErrMetadataInvalid
		}
		metadata = []byte(u.Metadata)
		updateAssetResult.Metadata = u.Metadata
	}

	if len(u.URI) > 0 {
		if len(u.URI) < 3 || len(u.URI) > storage.MaxTextSize {
			return nil, ErrURIInvalid
		}
		uri = []byte(u.URI)
		updateAssetResult.URI = u.URI
	}

	if u.MaxSupply > 0 {
		maxSupply = u.MaxSupply
		updateAssetResult.MaxSupply = u.MaxSupply
	}

	if len(u.Owner) > 0 {
		if owner, err = codec.StringToAddress(u.Owner); err != nil {
			return nil, ErrOutputOwnerInvalid
		}
		updateAssetResult.Owner = u.Owner
	}
	if len(u.MintAdmin) > 0 {
		if mintAdmin, err = codec.StringToAddress(u.MintAdmin); err != nil {
			return nil, ErrOutputMintAdminInvalid
		}
		updateAssetResult.MintAdmin = u.MintAdmin
	}
	if len(u.PauseUnpauseAdmin) > 0 {
		if pauseUnpauseAdmin, err = codec.StringToAddress(u.PauseUnpauseAdmin); err != nil {
			return nil, ErrOutputPauseUnpauseAdminInvalid
		}
		updateAssetResult.PauseUnpauseAdmin = u.PauseUnpauseAdmin
	}
	if len(u.FreezeUnfreezeAdmin) > 0 {
		if freezeUnfreezeAdmin, err = codec.StringToAddress(u.FreezeUnfreezeAdmin); err != nil {
			return nil, ErrOutputFreezeUnfreezeAdminInvalid
		}
		updateAssetResult.FreezeUnfreezeAdmin = u.FreezeUnfreezeAdmin
	}
	if len(u.EnableDisableKYCAccountAdmin) > 0 {
		if enableDisableKYCAccountAdmin, err = codec.StringToAddress(u.EnableDisableKYCAccountAdmin); err != nil {
			return nil, ErrOutputEnableDisableKYCAccountAdminInvalid
		}
		updateAssetResult.EnableDisableKYCAccountAdmin = u.EnableDisableKYCAccountAdmin
	}

	// Update the asset
	if err := storage.SetAssetInfo(ctx, mu, u.AssetAddress, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	return &updateAssetResult, nil
}

func (*UpdateAsset) ComputeUnits(chain.Rules) uint64 {
	return UpdateAssetComputeUnits
}

func (*UpdateAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalUpdateAsset(p *codec.Packer) (chain.Action, error) {
	var create UpdateAsset
	p.UnpackAddress(&create.AssetAddress)
	create.Name = p.UnpackString(false)
	create.Symbol = p.UnpackString(false)
	create.Metadata = p.UnpackString(false)
	create.URI = p.UnpackString(false)
	create.MaxSupply = p.UnpackUint64(false)
	create.Owner = p.UnpackString(false)
	create.MintAdmin = p.UnpackString(false)
	create.PauseUnpauseAdmin = p.UnpackString(false)
	create.FreezeUnfreezeAdmin = p.UnpackString(false)
	create.EnableDisableKYCAccountAdmin = p.UnpackString(false)
	return &create, p.Err()
}

var (
	_ codec.Typed     = (*UpdateAssetResult)(nil)
	_ chain.Marshaler = (*UpdateAssetResult)(nil)
)

type UpdateAssetResult struct {
	Name                         string `serialize:"true" json:"name"`
	Symbol                       string `serialize:"true" json:"symbol"`
	Metadata                     string `serialize:"true" json:"metadata"`
	URI                          string `serialize:"true" json:"uri"`
	MaxSupply                    uint64 `serialize:"true" json:"max_supply"`
	Owner                        string `serialize:"true" json:"owner"`
	MintAdmin                    string `serialize:"true" json:"mint_admin"`
	PauseUnpauseAdmin            string `serialize:"true" json:"pause_unpause_admin"`
	FreezeUnfreezeAdmin          string `serialize:"true" json:"freeze_unfreeze_admin"`
	EnableDisableKYCAccountAdmin string `serialize:"true" json:"enable_disable_kyc_account_admin"`
}

func (*UpdateAssetResult) GetTypeID() uint8 {
	return nconsts.UpdateAssetID
}

func (r *UpdateAssetResult) Size() int {
	return codec.StringLen(r.Name) + codec.StringLen(r.Symbol) + codec.StringLen(r.Metadata) + codec.StringLen(r.URI) + consts.Uint64Len + codec.StringLen(r.Owner) + codec.StringLen(r.MintAdmin) + codec.StringLen(r.PauseUnpauseAdmin) + codec.StringLen(r.FreezeUnfreezeAdmin) + codec.StringLen(r.EnableDisableKYCAccountAdmin)
}

func (r *UpdateAssetResult) Marshal(p *codec.Packer) {
	p.PackString(r.Name)
	p.PackString(r.Symbol)
	p.PackString(r.Metadata)
	p.PackString(r.URI)
	p.PackUint64(r.MaxSupply)
	p.PackString(r.Owner)
	p.PackString(r.MintAdmin)
	p.PackString(r.PauseUnpauseAdmin)
	p.PackString(r.FreezeUnfreezeAdmin)
	p.PackString(r.EnableDisableKYCAccountAdmin)
}

func UnmarshalUpdateAssetResult(p *codec.Packer) (codec.Typed, error) {
	var result UpdateAssetResult
	result.Name = p.UnpackString(false)
	result.Symbol = p.UnpackString(false)
	result.Metadata = p.UnpackString(false)
	result.URI = p.UnpackString(false)
	result.MaxSupply = p.UnpackUint64(false)
	p.UnpackString(false)
	p.UnpackString(false)
	p.UnpackString(false)
	p.UnpackString(false)
	p.UnpackString(false)
	return &result, p.Err()
}
