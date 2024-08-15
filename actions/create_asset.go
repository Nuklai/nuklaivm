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

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*CreateAsset)(nil)

type CreateAsset struct {
	// The name of the asset
	Name []byte `json:"name"`

	// The symbol of the asset
	Symbol []byte `json:"symbol"`

	// The number of decimal places in the asset
	Decimals uint8 `json:"decimals"`

	// The metadata of the asset
	Metadata []byte `json:"metadata"`

	// The max supply of the asset
	MaxSupply uint64 `json:"maxSupply"`

	// The wallet address that can update this asset
	UpdateAssetActor codec.Address `json:"updateAssetActor"`

	// The wallet address that can mint/burn assets
	MintActor codec.Address `json:"mintBurnActor"`

	// The wallet address that can pause/unpause assets
	PauseUnpauseActor codec.Address `json:"pauseUnpauseActor"`

	// The wallet address that can freeze/unfreeze assets
	FreezeUnfreezeActor codec.Address `json:"freezeUnfreezeActor"`

	// The wallet address that can enable/disable KYC account flag
	EnableDisableKYCAccountActor codec.Address `json:"enableDisableKYCAccountActor"`

	// The wallet address that can delete assets
	DeleteActor codec.Address `json:"deleteActor"`
}

func (*CreateAsset) GetTypeID() uint8 {
	return nconsts.CreateAssetID
}

func (*CreateAsset) StateKeys(_ codec.Address, actionID ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(actionID)): state.Allocate | state.Write,
	}
}

func (*CreateAsset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks}
}

func (c *CreateAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	_ codec.Address,
	actionID ids.ID,
) ([][]byte, error) {
	if len(c.Name) == 0 || len(c.Name) > MaxTextSize {
		return nil, ErrOutputNameInvalid
	}
	if len(c.Symbol) == 0 || len(c.Symbol) > MaxTextSize {
		return nil, ErrOutputSymbolInvalid
	}
	if c.Decimals == 0 || c.Decimals > MaxDecimals {
		return nil, ErrOutputDecimalsInvalid
	}
	if len(c.Metadata) == 0 || len(c.Metadata) > MaxMetadataSize {
		return nil, ErrOutputMetadataInvalid
	}
	if _, err := codec.AddressBech32(nconsts.HRP, c.UpdateAssetActor); err != nil {
		return nil, err
	}
	if _, err := codec.AddressBech32(nconsts.HRP, c.MintActor); err != nil {
		return nil, err
	}
	if _, err := codec.AddressBech32(nconsts.HRP, c.PauseUnpauseActor); err != nil {
		return nil, err
	}
	if _, err := codec.AddressBech32(nconsts.HRP, c.FreezeUnfreezeActor); err != nil {
		return nil, err
	}
	if _, err := codec.AddressBech32(nconsts.HRP, c.EnableDisableKYCAccountActor); err != nil {
		return nil, err
	}
	if _, err := codec.AddressBech32(nconsts.HRP, c.DeleteActor); err != nil {
		return nil, err
	}

	if err := storage.SetAsset(ctx, mu, actionID, c.Name, c.Symbol, c.Decimals, c.Metadata, 0, c.MaxSupply, c.UpdateAssetActor, c.MintActor, c.PauseUnpauseActor, c.FreezeUnfreezeActor, c.EnableDisableKYCAccountActor, c.DeleteActor); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*CreateAsset) ComputeUnits(chain.Rules) uint64 {
	return CreateAssetComputeUnits
}

func (c *CreateAsset) Size() int {
	// TODO: add small bytes (smaller int prefix)
	return codec.BytesLen(c.Name) + codec.BytesLen(c.Symbol) + consts.Uint8Len + codec.BytesLen(c.Metadata) + consts.Uint64Len + codec.AddressLen*6
}

func (c *CreateAsset) Marshal(p *codec.Packer) {
	p.PackBytes(c.Name)
	p.PackBytes(c.Symbol)
	p.PackByte(c.Decimals)
	p.PackBytes(c.Metadata)
	p.PackUint64(c.MaxSupply)
	p.PackAddress(c.UpdateAssetActor)
	p.PackAddress(c.MintActor)
	p.PackAddress(c.PauseUnpauseActor)
	p.PackAddress(c.FreezeUnfreezeActor)
	p.PackAddress(c.EnableDisableKYCAccountActor)
	p.PackAddress(c.DeleteActor)
}

func UnmarshalCreateAsset(p *codec.Packer) (chain.Action, error) {
	var create CreateAsset
	p.UnpackBytes(MaxTextSize, true, &create.Name)
	p.UnpackBytes(MaxTextSize, true, &create.Symbol)
	create.Decimals = p.UnpackByte()
	p.UnpackBytes(MaxMetadataSize, true, &create.Metadata)
	create.MaxSupply = p.UnpackUint64(false)
	p.UnpackAddress(&create.UpdateAssetActor)
	p.UnpackAddress(&create.MintActor)
	p.UnpackAddress(&create.PauseUnpauseActor)
	p.UnpackAddress(&create.FreezeUnfreezeActor)
	p.UnpackAddress(&create.EnableDisableKYCAccountActor)
	p.UnpackAddress(&create.DeleteActor)
	return &create, p.Err()
}

func (*CreateAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
