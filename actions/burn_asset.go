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

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*BurnAsset)(nil)

type BurnAsset struct {
	// Asset is the AssetID of the asset.
	Asset ids.ID `json:"asset"`

	// For FT: Number of assets to burn.
	// For NFT: Unique ID of the NFT to burn.
	Value uint64 `json:"value"`

	// Is the asset FT or NFT
	IsNFT bool `json:"isNFT"`
}

func (*BurnAsset) GetTypeID() uint8 {
	return nconsts.BurnAssetID
}

func (b *BurnAsset) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	nftID := nchain.GenerateID(b.Asset, b.Value)
	return state.Keys{
		string(storage.AssetKey(b.Asset)):             state.Read | state.Write,
		string(storage.AssetNFTKey(b.Asset, b.Value)): state.Read | state.Write,
		string(storage.BalanceKey(actor, b.Asset)):    state.Read | state.Write,
		string(storage.BalanceKey(actor, nftID)):      state.Read | state.Write,
	}
}

func (*BurnAsset) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.AssetNFTChunks, storage.BalanceChunks}
}

func (b *BurnAsset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	if !b.IsNFT && b.Value == 0 {
		return nil, ErrOutputValueZero
	}

	exists, name, symbol, decimals, metadata, totalSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor, err := storage.GetAsset(ctx, mu, b.Asset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	// Check if the asset is NFT
	if b.IsNFT {
		exists, _, _, _, err := storage.GetAssetNFT(ctx, mu, b.Asset, b.Value)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrOutputAssetMissing
		}
	}

	newSupply, err := smath.Sub(totalSupply, b.Value)
	if err != nil {
		return nil, err
	}
	if err := storage.SetAsset(ctx, mu, b.Asset, name, symbol, decimals, metadata, newSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor); err != nil {
		return nil, err
	}

	// Handle for NFT
	if b.IsNFT {
		if err := storage.DeleteAssetNFT(ctx, mu, b.Asset, b.Value); err != nil {
			return nil, err
		}

		if err := storage.SubBalance(ctx, mu, actor, b.Asset, 1); err != nil {
			return nil, err
		}

		nftID := nchain.GenerateID(b.Asset, b.Value)
		if err := storage.SubBalance(ctx, mu, actor, nftID, 1); err != nil {
			return nil, err
		}
	} else {
		if err := storage.SubBalance(ctx, mu, actor, b.Asset, b.Value); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (*BurnAsset) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetComputeUnits
}

func (*BurnAsset) Size() int {
	return ids.IDLen + consts.Uint64Len + consts.BoolLen
}

func (b *BurnAsset) Marshal(p *codec.Packer) {
	p.PackID(b.Asset)
	p.PackUint64(b.Value)
	p.PackBool(b.IsNFT)
}

func UnmarshalBurnAsset(p *codec.Packer) (chain.Action, error) {
	var burn BurnAsset
	p.UnpackID(false, &burn.Asset) // can burn native asset
	burn.Value = p.UnpackUint64(false)
	burn.IsNFT = p.UnpackBool()
	return &burn, p.Err()
}

func (*BurnAsset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
