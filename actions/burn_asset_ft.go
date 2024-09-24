// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	BurnAssetFTComputeUnits = 1
)

var (
	_ chain.Action = (*BurnAssetFT)(nil)
)

type BurnAssetFT struct {
	// AssetID ID of the asset to burn.
	AssetID ids.ID `serialize:"true" json:"asset_id"`

	// Number of assets to burn
	Value uint64 `serialize:"true" json:"value"`
}

func (*BurnAssetFT) GetTypeID() uint8 {
	return nconsts.BurnAssetFTID
}

func (b *BurnAssetFT) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(b.AssetID)):          state.Read | state.Write,
		string(storage.BalanceKey(actor, b.AssetID)): state.Read | state.Write,
	}
}

func (*BurnAssetFT) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.BalanceChunks}
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
		return nil, ErrOutputValueZero
	}

	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, b.AssetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetFungibleTokenID {
		return nil, ErrOutputWrongAssetType
	}

	newSupply, err := smath.Sub(totalSupply, b.Value)
	if err != nil {
		return nil, err
	}
	if err := storage.SetAsset(ctx, mu, b.AssetID, assetType, name, symbol, decimals, metadata, uri, newSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	newBalance, err := storage.SubBalance(ctx, mu, actor, b.AssetID, b.Value)
	if err != nil {
		return nil, err
	}

	return &BurnAssetFTResult{
		From:             actor.String(),
		OldBalance:       newBalance + b.Value,
		NewBalance:       newBalance,
		AssetTotalSupply: newSupply,
	}, nil
}

func (*BurnAssetFT) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetFTComputeUnits
}

func (*BurnAssetFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*BurnAssetFT)(nil)

func (*BurnAssetFT) Size() int {
	return ids.IDLen + consts.Uint64Len
}

func (b *BurnAssetFT) Marshal(p *codec.Packer) {
	p.PackID(b.AssetID)
	p.PackUint64(b.Value)
}

func UnmarshalBurnAssetFT(p *codec.Packer) (chain.Action, error) {
	var burn BurnAssetFT
	p.UnpackID(false, &burn.AssetID) // can burn native asset
	burn.Value = p.UnpackUint64(false)
	return &burn, p.Err()
}

var _ codec.Typed = (*BurnAssetFTResult)(nil)

type BurnAssetFTResult struct {
	From             string `serialize:"true" json:"from"`
	OldBalance       uint64 `serialize:"true" json:"old_balance"`
	NewBalance       uint64 `serialize:"true" json:"new_balance"`
	AssetTotalSupply uint64 `serialize:"true" json:"asset_total_supply"`
}

func (*BurnAssetFTResult) GetTypeID() uint8 {
	return nconsts.BurnAssetFTID
}
