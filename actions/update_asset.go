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
	ErrOutputAssetNotFound                                    = errors.New("asset not found")
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
	// Asset ID to update
	AssetID ids.ID `serialize:"true" json:"asset_id"`

	// The name of the asset
	Name []byte `serialize:"true" json:"name"`

	// The symbol of the asset
	Symbol []byte `serialize:"true" json:"symbol"`

	// The metadata of the asset
	Metadata []byte `serialize:"true" json:"metadata"`

	// URI for the asset
	URI []byte `serialize:"true" json:"uri"`

	// The max supply of the asset
	MaxSupply uint64 `serialize:"true" json:"max_supply"`

	// Owner of the asset
	Owner []byte `serialize:"true" json:"owner"`

	// The wallet address that can mint/burn assets
	MintAdmin []byte `serialize:"true" json:"mint_admin"`

	// The wallet address that can pause/unpause assets
	PauseUnpauseAdmin []byte `serialize:"true" json:"pause_unpause_admin"`

	// The wallet address that can freeze/unfreeze assets
	FreezeUnfreezeAdmin []byte `serialize:"true" json:"freeze_unfreeze_admin"`

	// The wallet address that can enable/disable KYC account flag
	EnableDisableKYCAccountAdmin []byte `serialize:"true" json:"enable_disable_kyc_account_admin"`
}

func (*UpdateAsset) GetTypeID() uint8 {
	return nconsts.UpdateAssetID
}

func (u *UpdateAsset) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(u.AssetID)): state.Read | state.Write,
	}
}

func (*UpdateAsset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks}
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
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, mu, u.AssetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetNotFound
	}

	// Check if the actor is the owner of the asset
	if owner != actor {
		return nil, ErrOutputWrongOwner
	}

	// Note that maxSupply can never be set to 0 on an update.
	// It can only be increased or decreased.
	// If maxSupply is set to 0, it will be set to the current maxSupply.
	if u.MaxSupply == 0 {
		u.MaxSupply = maxSupply
	} else {
		// Ensure that the max supply is greater than or equal to the total supply
		if u.MaxSupply < totalSupply {
			return nil, ErrOutputMaxSupplyInvalid
		}
	}

	// Ensure that at least one field is being updated
	if (len(u.Name) == 0 || bytes.Equal(u.Name, name)) && (len(u.Symbol) == 0 || bytes.Equal(u.Symbol, symbol)) && (len(u.Metadata) == 0 || bytes.Equal(u.Metadata, metadata)) && (len(u.URI) == 0 || bytes.Equal(u.URI, uri)) && (u.MaxSupply == maxSupply) && (len(u.Owner) == 0 || string(u.Owner) != owner.String()) && (len(u.MintAdmin) == 0 || bytes.Equal(u.MintAdmin, mintAdmin[:])) && (len(u.PauseUnpauseAdmin) == 0 || bytes.Equal(u.PauseUnpauseAdmin, pauseUnpauseAdmin[:])) && (len(u.FreezeUnfreezeAdmin) == 0 || bytes.Equal(u.FreezeUnfreezeAdmin, freezeUnfreezeAdmin[:])) && (len(u.EnableDisableKYCAccountAdmin) == 0 || bytes.Equal(u.EnableDisableKYCAccountAdmin, enableDisableKYCAccountAdmin[:])) {
		return nil, ErrOutputMustUpdateAtLeastOneField
	}

	updateAssetResult := UpdateAssetResult{
		OldAssetInfo: AssetInfo{
			Name:                         string(name),
			Symbol:                       string(symbol),
			Metadata:                     string(metadata),
			URI:                          string(uri),
			MaxSupply:                    maxSupply,
			Owner:                        owner.String(),
			MintAdmin:                    mintAdmin.String(),
			PauseUnpauseAdmin:            pauseUnpauseAdmin.String(),
			FreezeUnfreezeAdmin:          freezeUnfreezeAdmin.String(),
			EnableDisableKYCAccountAdmin: enableDisableKYCAccountAdmin.String(),
		},
		NewAssetInfo: AssetInfo{
			Name:                         string(name),
			Symbol:                       string(symbol),
			Metadata:                     string(metadata),
			URI:                          string(uri),
			MaxSupply:                    maxSupply,
			Owner:                        owner.String(),
			MintAdmin:                    mintAdmin.String(),
			PauseUnpauseAdmin:            pauseUnpauseAdmin.String(),
			FreezeUnfreezeAdmin:          freezeUnfreezeAdmin.String(),
			EnableDisableKYCAccountAdmin: enableDisableKYCAccountAdmin.String(),
		},
	}

	// if u.Name is passed, update the dataset name
	// otherwise, keep the existing name
	if len(u.Name) > 0 {
		if len(u.Name) < 3 || len(u.Name) > MaxMetadataSize {
			return nil, ErrOutputNameInvalid
		}
		name = u.Name
		updateAssetResult.NewAssetInfo.Name = string(u.Name)
	}

	if len(u.Symbol) > 0 {
		if len(u.Symbol) < 3 || len(u.Symbol) > MaxTextSize {
			return nil, ErrOutputSymbolInvalid
		}
		symbol = u.Symbol
		updateAssetResult.NewAssetInfo.Symbol = string(u.Symbol)
	}

	if len(u.Metadata) > 0 {
		if len(u.Metadata) < 3 || len(u.Metadata) > MaxMetadataSize {
			return nil, ErrOutputMetadataInvalid
		}
		metadata = u.Metadata
		updateAssetResult.NewAssetInfo.Metadata = string(u.Metadata)
	}

	if len(u.URI) > 0 {
		if len(u.URI) < 3 || len(u.URI) > MaxMetadataSize {
			return nil, ErrOutputURIInvalid
		}
		uri = u.URI
		updateAssetResult.NewAssetInfo.URI = string(u.URI)
	}

	if u.MaxSupply > 0 {
		maxSupply = u.MaxSupply
		updateAssetResult.NewAssetInfo.MaxSupply = u.MaxSupply
	}

	if len(u.Owner) > 0 {
		if owner, err = codec.ToAddress(u.Owner); err != nil {
			return nil, ErrOutputOwnerInvalid
		}
		updateAssetResult.NewAssetInfo.Owner = string(u.Owner)
	}
	if len(u.MintAdmin) > 0 {
		if mintAdmin, err = codec.ToAddress(u.MintAdmin); err != nil {
			return nil, ErrOutputMintAdminInvalid
		}
		updateAssetResult.NewAssetInfo.MintAdmin = string(u.MintAdmin)
	}
	if len(u.PauseUnpauseAdmin) > 0 {
		if pauseUnpauseAdmin, err = codec.ToAddress(u.PauseUnpauseAdmin); err != nil {
			return nil, ErrOutputPauseUnpauseAdminInvalid
		}
		updateAssetResult.NewAssetInfo.PauseUnpauseAdmin = string(u.PauseUnpauseAdmin)
	}
	if len(u.FreezeUnfreezeAdmin) > 0 {
		if freezeUnfreezeAdmin, err = codec.ToAddress(u.FreezeUnfreezeAdmin); err != nil {
			return nil, ErrOutputFreezeUnfreezeAdminInvalid
		}
		updateAssetResult.NewAssetInfo.FreezeUnfreezeAdmin = string(u.FreezeUnfreezeAdmin)
	}
	if len(u.EnableDisableKYCAccountAdmin) > 0 {
		if enableDisableKYCAccountAdmin, err = codec.ToAddress(u.EnableDisableKYCAccountAdmin); err != nil {
			return nil, ErrOutputEnableDisableKYCAccountAdminInvalid
		}
		updateAssetResult.NewAssetInfo.EnableDisableKYCAccountAdmin = string(u.EnableDisableKYCAccountAdmin)
	}

	if err := storage.SetAsset(ctx, mu, u.AssetID, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
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

var _ chain.Marshaler = (*UpdateAsset)(nil)

func (u *UpdateAsset) Size() int {
	return ids.IDLen + codec.BytesLen(u.Name) + codec.BytesLen(u.Symbol) + codec.BytesLen(u.Metadata) + codec.BytesLen(u.URI) + consts.Uint64Len + codec.AddressLen*5
}

func (u *UpdateAsset) Marshal(p *codec.Packer) {
	p.PackID(u.AssetID)
	p.PackBytes(u.Name)
	p.PackBytes(u.Symbol)
	p.PackBytes(u.Metadata)
	p.PackBytes(u.URI)
	p.PackUint64(u.MaxSupply)
	p.PackBytes(u.Owner)
	p.PackBytes(u.MintAdmin)
	p.PackBytes(u.PauseUnpauseAdmin)
	p.PackBytes(u.FreezeUnfreezeAdmin)
	p.PackBytes(u.EnableDisableKYCAccountAdmin)
}

func UnmarshalUpdateAsset(p *codec.Packer) (chain.Action, error) {
	var create UpdateAsset
	p.UnpackID(true, &create.AssetID)
	p.UnpackBytes(MaxMetadataSize, false, &create.Name)
	p.UnpackBytes(MaxTextSize, false, &create.Symbol)
	p.UnpackBytes(MaxMetadataSize, false, &create.Metadata)
	p.UnpackBytes(MaxMetadataSize, false, &create.URI)
	create.MaxSupply = p.UnpackUint64(false)
	p.UnpackBytes(codec.AddressLen, false, &create.Owner)
	p.UnpackBytes(codec.AddressLen, false, &create.MintAdmin)
	p.UnpackBytes(codec.AddressLen, false, &create.PauseUnpauseAdmin)
	p.UnpackBytes(codec.AddressLen, false, &create.FreezeUnfreezeAdmin)
	p.UnpackBytes(codec.AddressLen, false, &create.EnableDisableKYCAccountAdmin)
	return &create, p.Err()
}

var _ codec.Typed = (*UpdateAssetResult)(nil)

type AssetInfo struct {
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

type UpdateAssetResult struct {
	OldAssetInfo AssetInfo `serialize:"true" json:"old_asset_info"`
	NewAssetInfo AssetInfo `serialize:"true" json:"new_asset_info"`
}

func (*UpdateAssetResult) GetTypeID() uint8 {
	return nconsts.UpdateAssetID
}
