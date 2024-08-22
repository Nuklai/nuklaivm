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

var _ chain.Action = (*BurnAssetFT)(nil)

type BurnAssetFT struct {
	// Asset ID of the asset to burn.
	Asset ids.ID `json:"asset"`

	// Number of assets to burn
	Value uint64 `json:"value"`
}

func (*BurnAssetFT) GetTypeID() uint8 {
	return nconsts.BurnAssetFTID
}

func (b *BurnAssetFT) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(b.Asset)):          state.Read | state.Write,
		string(storage.BalanceKey(actor, b.Asset)): state.Read | state.Write,
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
) ([][]byte, error) {
	if b.Value == 0 {
		return nil, ErrOutputValueZero
	}

	exists, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor, err := storage.GetAsset(ctx, mu, b.Asset)
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
	if err := storage.SetAsset(ctx, mu, b.Asset, assetType, name, symbol, decimals, metadata, newSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor); err != nil {
		return nil, err
	}

	if err := storage.SubBalance(ctx, mu, actor, b.Asset, b.Value); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*BurnAssetFT) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetComputeUnits
}

func (*BurnAssetFT) Size() int {
	return ids.IDLen + consts.Uint64Len
}

func (b *BurnAssetFT) Marshal(p *codec.Packer) {
	p.PackID(b.Asset)
	p.PackUint64(b.Value)
}

func UnmarshalBurnAssetFT(p *codec.Packer) (chain.Action, error) {
	var burn BurnAssetFT
	p.UnpackID(false, &burn.Asset) // can burn native asset
	burn.Value = p.UnpackUint64(false)
	return &burn, p.Err()
}

func (*BurnAssetFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
