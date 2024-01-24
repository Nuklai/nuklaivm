// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	hmath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*BurnAsset)(nil)

type BurnAsset struct {
	// Asset is the [TxID] that created the asset.
	Asset ids.ID `json:"asset"`

	// Number of assets to mint to [To].
	Value uint64 `json:"value"`
}

func (*BurnAsset) GetTypeID() uint8 {
	return nconsts.BurnAssetID
}

func (b *BurnAsset) StateKeys(actor codec.Address, _ ids.ID) []string {
	return []string{
		string(storage.AssetKey(b.Asset)),
		string(storage.BalanceKey(actor, b.Asset)),
	}
}

func (*BurnAsset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.BalanceChunks}
}

func (*BurnAsset) OutputsWarpMessage() bool {
	return false
}

func (b *BurnAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	if b.Value == 0 {
		return false, BurnAssetComputeUnits, OutputValueZero, nil, nil
	}
	if err := storage.SubBalance(ctx, mu, actor, b.Asset, b.Value); err != nil {
		return false, BurnAssetComputeUnits, utils.ErrBytes(err), nil, nil
	}
	exists, symbol, decimals, metadata, supply, owner, warp, err := storage.GetAsset(ctx, mu, b.Asset)
	if err != nil {
		return false, BurnAssetComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if !exists {
		return false, BurnAssetComputeUnits, OutputAssetMissing, nil, nil
	}
	newSupply, err := hmath.Sub(supply, b.Value)
	if err != nil {
		return false, BurnAssetComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.SetAsset(ctx, mu, b.Asset, symbol, decimals, metadata, newSupply, owner, warp); err != nil {
		return false, BurnAssetComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, BurnAssetComputeUnits, nil, nil, nil
}

func (*BurnAsset) MaxComputeUnits(chain.Rules) uint64 {
	return BurnAssetComputeUnits
}

func (*BurnAsset) Size() int {
	return hconsts.IDLen + hconsts.Uint64Len
}

func (b *BurnAsset) Marshal(p *codec.Packer) {
	p.PackID(b.Asset)
	p.PackUint64(b.Value)
}

func UnmarshalBurnAsset(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var burn BurnAsset
	p.UnpackID(false, &burn.Asset) // can burn native asset
	burn.Value = p.UnpackUint64(true)
	return &burn, p.Err()
}

func (*BurnAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
