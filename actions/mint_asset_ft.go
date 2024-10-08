// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
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
	MintAssetFTComputeUnits = 5
)

var (
	ErrAssetIsNative               = errors.New("asset is native")
	ErrAssetMissing                = errors.New("asset missing")
	ErrWrongMintAdmin              = errors.New("mint admin is not correct")
	_                 chain.Action = (*MintAssetFT)(nil)
)

type MintAssetFT struct {
	// AssetAddress is the AssetAddress of the asset to mint.
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

	// Number of assets to mint to [To].
	Value uint64 `serialize:"true" json:"value"`

	// To is the recipient of the [Value].
	To codec.Address `serialize:"true" json:"to"`
}

func (*MintAssetFT) GetTypeID() uint8 {
	return nconsts.MintAssetFTID
}

func (m *MintAssetFT) StateKeys(codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(m.AssetAddress)):                 state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(m.AssetAddress, m.To)): state.All,
	}
}

func (m *MintAssetFT) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	if m.AssetAddress == storage.NAIAddress {
		return nil, ErrAssetIsNative
	}
	if m.Value == 0 {
		return nil, ErrValueZero
	}

	assetType, _, _, _, _, _, _, _, _, mintAdmin, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, m.AssetAddress)
	if err != nil {
		return nil, err
	}
	// Ensure that it's a fungible token
	if assetType != nconsts.AssetFungibleTokenID {
		return nil, ErrAssetTypeInvalid
	}
	if mintAdmin != actor {
		return nil, ErrWrongMintAdmin
	}

	// Minting logic for fungible tokens
	newBalance, err := storage.MintAsset(ctx, mu, m.AssetAddress, m.To, m.Value)
	if err != nil {
		return nil, err
	}

	return &MintAssetFTResult{
		OldBalance: newBalance - m.Value,
		NewBalance: newBalance,
	}, nil
}

func (*MintAssetFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetFTComputeUnits
}

func (*MintAssetFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalMintAssetFT(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetFT
	p.UnpackAddress(&mint.AssetAddress)
	mint.Value = p.UnpackUint64(true)
	p.UnpackAddress(&mint.To)
	return &mint, p.Err()
}

var (
	_ codec.Typed     = (*MintAssetFTResult)(nil)
	_ chain.Marshaler = (*MintAssetFTResult)(nil)
)

type MintAssetFTResult struct {
	OldBalance uint64 `serialize:"true" json:"old_balance"`
	NewBalance uint64 `serialize:"true" json:"new_balance"`
}

func (*MintAssetFTResult) GetTypeID() uint8 {
	return nconsts.MintAssetFTID
}

func (*MintAssetFTResult) Size() int {
	return consts.Uint64Len * 2
}

func (r *MintAssetFTResult) Marshal(p *codec.Packer) {
	p.PackUint64(r.OldBalance)
	p.PackUint64(r.NewBalance)
}

func UnmarshalMintAssetFTResult(p *codec.Packer) (codec.Typed, error) {
	var result MintAssetFTResult
	result.OldBalance = p.UnpackUint64(false)
	result.NewBalance = p.UnpackUint64(true)
	return &result, p.Err()
}
