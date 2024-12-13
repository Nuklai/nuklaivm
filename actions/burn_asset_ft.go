// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	BurnAssetFTComputeUnits = 1
)

var _ chain.Action = (*BurnAssetFT)(nil)

type BurnAssetFT struct {
	// AssetAddress of the asset to burn.
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

	// Number of assets to burn
	Value uint64 `serialize:"true" json:"value"`
}

func (*BurnAssetFT) GetTypeID() uint8 {
	return nconsts.BurnAssetFTID
}

func (b *BurnAssetFT) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(b.AssetAddress)):                  state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(b.AssetAddress, actor)): state.Read | state.Write,
	}
}

func (b *BurnAssetFT) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	if b.Value == 0 {
		return nil, ErrValueZero
	}

	assetType, _, _, _, _, _, _, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, b.AssetAddress)
	if err != nil {
		return nil, err
	}
	// Ensure that it's a fungible token
	if assetType != nconsts.AssetFungibleTokenID {
		return nil, ErrAssetTypeInvalid
	}

	// Burning logic for fungible tokens
	newBalance, err := storage.BurnAsset(ctx, mu, b.AssetAddress, actor, b.Value)
	if err != nil {
		return nil, err
	}

	return &BurnAssetFTResult{
		Actor:      actor.String(),
		Receiver:   "",
		OldBalance: newBalance + b.Value,
		NewBalance: newBalance,
	}, nil
}

func (*BurnAssetFT) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetFTComputeUnits
}

func (*BurnAssetFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalBurnAssetFT(p *codec.Packer) (chain.Action, error) {
	var burn BurnAssetFT
	p.UnpackAddress(&burn.AssetAddress)
	burn.Value = p.UnpackUint64(true)
	return &burn, p.Err()
}

var _ codec.Typed = (*BurnAssetFTResult)(nil)

type BurnAssetFTResult struct {
	Actor      string `serialize:"true" json:"actor"`
	Receiver   string `serialize:"true" json:"receiver"`
	OldBalance uint64 `serialize:"true" json:"old_balance"`
	NewBalance uint64 `serialize:"true" json:"new_balance"`
}

func (*BurnAssetFTResult) GetTypeID() uint8 {
	return nconsts.BurnAssetFTID
}

func UnmarshalBurnAssetFTResult(p *codec.Packer) (codec.Typed, error) {
	var result BurnAssetFTResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.OldBalance = p.UnpackUint64(true)
	result.NewBalance = p.UnpackUint64(false)
	return &result, p.Err()
}
