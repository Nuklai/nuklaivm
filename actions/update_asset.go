// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"bytes"
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*UpdateAsset)(nil)

type UpdateAsset struct {
	// Asset ID to update
	Asset ids.ID `json:"asset"`

	// The name of the asset
	Name []byte `json:"name"`

	// The symbol of the asset
	Symbol []byte `json:"symbol"`

	// The metadata of the asset
	Metadata []byte `json:"metadata"`

	// URI for the asset
	URI []byte `json:"uri"`

	// The max supply of the asset
	MaxSupply uint64 `json:"maxSupply"`

	// Admin of the asset
	Admin []byte `json:"admin"`

	// The wallet address that can mint/burn assets
	MintActor []byte `json:"mintActor"`

	// The wallet address that can pause/unpause assets
	PauseUnpauseActor []byte `json:"pauseUnpauseActor"`

	// The wallet address that can freeze/unfreeze assets
	FreezeUnfreezeActor []byte `json:"freezeUnfreezeActor"`

	// The wallet address that can enable/disable KYC account flag
	EnableDisableKYCAccountActor []byte `json:"enableDisableKYCAccountActor"`
}

func (*UpdateAsset) GetTypeID() uint8 {
	return nconsts.UpdateAssetID
}

func (u *UpdateAsset) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(u.Asset)): state.Allocate | state.Write,
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
) ([][]byte, error) {
	// Check if the asset exists
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, u.Asset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssetNotFound
	}

	// Check if the actor is the owner of the asset
	if admin != actor {
		return nil, ErrOutputWrongOwner
	}

	// Ensure that at least one field is being updated
	if (len(u.Name) == 0 || bytes.Equal(u.Name, name)) && (len(u.Symbol) == 0 || bytes.Equal(u.Symbol, symbol)) && (len(u.Metadata) == 0 || bytes.Equal(u.Metadata, metadata)) && (len(u.URI) == 0 || bytes.Equal(u.URI, uri)) && u.MaxSupply == maxSupply && (len(u.Admin) == 0 || bytes.Equal(u.Admin, admin[:])) && (len(u.MintActor) == 0 || bytes.Equal(u.MintActor, mintActor[:])) && (len(u.PauseUnpauseActor) == 0 || bytes.Equal(u.PauseUnpauseActor, pauseUnpauseActor[:])) && (len(u.FreezeUnfreezeActor) == 0 || bytes.Equal(u.FreezeUnfreezeActor, freezeUnfreezeActor[:])) && (len(u.EnableDisableKYCAccountActor) == 0 || bytes.Equal(u.EnableDisableKYCAccountActor, enableDisableKYCAccountActor[:])) {
		return nil, ErrOutputMustUpdateAtLeastOneField
	}

	// if u.Name is passed, update the dataset name
	// otherwise, keep the existing name
	if len(u.Name) > 0 {
		if len(u.Name) < 3 || len(u.Name) > MaxMetadataSize {
			return nil, ErrOutputNameInvalid
		}
		name = u.Name
	}

	if len(u.Symbol) > 0 {
		if len(u.Symbol) < 3 || len(u.Symbol) > MaxTextSize {
			return nil, ErrOutputSymbolInvalid
		}
		symbol = u.Symbol
	}

	if len(u.Metadata) > 0 {
		if len(u.Metadata) < 3 || len(u.Metadata) > MaxMetadataSize {
			return nil, ErrOutputMetadataInvalid
		}
		metadata = u.Metadata
	}

	if len(u.URI) > 0 {
		if len(u.URI) < 3 || len(u.URI) > MaxMetadataSize {
			return nil, ErrOutputURIInvalid
		}
		uri = u.URI
	}

	maxSupply = u.MaxSupply

	if len(u.Admin) > 0 {
		adminAddr, err := codec.ParseAddressBech32(nconsts.HRP, string(u.Admin))
		if err != nil {
			return nil, err
		}
		admin = adminAddr
	}

	if len(u.MintActor) > 0 {
		mintAddr, err := codec.ParseAddressBech32(nconsts.HRP, string(u.MintActor))
		if err != nil {
			return nil, err
		}
		mintActor = mintAddr
	}

	if len(u.PauseUnpauseActor) > 0 {
		pauseUnpauseAddr, err := codec.ParseAddressBech32(nconsts.HRP, string(u.PauseUnpauseActor))
		if err != nil {
			return nil, err
		}
		pauseUnpauseActor = pauseUnpauseAddr
	}

	if len(u.FreezeUnfreezeActor) > 0 {
		freezeUnfreezeAddr, err := codec.ParseAddressBech32(nconsts.HRP, string(u.FreezeUnfreezeActor))
		if err != nil {
			return nil, err
		}
		freezeUnfreezeActor = freezeUnfreezeAddr
	}

	if len(u.EnableDisableKYCAccountActor) > 0 {
		enableDisableKYCAccountAddr, err := codec.ParseAddressBech32(nconsts.HRP, string(u.EnableDisableKYCAccountActor))
		if err != nil {
			return nil, err
		}
		enableDisableKYCAccountActor = enableDisableKYCAccountAddr
	}

	if err := storage.SetAsset(ctx, mu, u.Asset, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*UpdateAsset) ComputeUnits(chain.Rules) uint64 {
	return UpdateAssetComputeUnits
}

func (u *UpdateAsset) Size() int {
	return ids.IDLen + codec.BytesLen(u.Name) + codec.BytesLen(u.Symbol) + codec.BytesLen(u.Metadata) + codec.BytesLen(u.URI) + consts.Uint64Len + codec.AddressLen*5
}

func (u *UpdateAsset) Marshal(p *codec.Packer) {
	p.PackID(u.Asset)
	p.PackBytes(u.Name)
	p.PackBytes(u.Symbol)
	p.PackBytes(u.Metadata)
	p.PackBytes(u.URI)
	p.PackUint64(u.MaxSupply)
	p.PackBytes(u.Admin)
	p.PackBytes(u.MintActor)
	p.PackBytes(u.PauseUnpauseActor)
	p.PackBytes(u.FreezeUnfreezeActor)
	p.PackBytes(u.EnableDisableKYCAccountActor)
}

func UnmarshalUpdateAsset(p *codec.Packer) (chain.Action, error) {
	var create UpdateAsset
	p.UnpackID(true, &create.Asset)
	p.UnpackBytes(MaxMetadataSize, false, &create.Name)
	p.UnpackBytes(MaxTextSize, false, &create.Symbol)
	p.UnpackBytes(MaxMetadataSize, false, &create.Metadata)
	p.UnpackBytes(MaxMetadataSize, false, &create.URI)
	create.MaxSupply = p.UnpackUint64(false)
	p.UnpackBytes(codec.AddressLen, false, &create.Admin)
	p.UnpackBytes(codec.AddressLen, false, &create.MintActor)
	p.UnpackBytes(codec.AddressLen, false, &create.PauseUnpauseActor)
	p.UnpackBytes(codec.AddressLen, false, &create.FreezeUnfreezeActor)
	p.UnpackBytes(codec.AddressLen, false, &create.EnableDisableKYCAccountActor)
	return &create, p.Err()
}

func (*UpdateAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
