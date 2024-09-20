// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*BurnAssetNFT)(nil)

type BurnAssetNFT struct {
	// Asset ID of the asset to burn(this is the nft collection ID)
	Asset ids.ID `json:"asset"`

	// NFT ID of the asset to burn
	NftID ids.ID `json:"nftID"`
}

func (*BurnAssetNFT) GetTypeID() uint8 {
	return nconsts.BurnAssetNFTID
}

func (b *BurnAssetNFT) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(b.Asset)):          state.Read | state.Write,
		string(storage.AssetNFTKey(b.NftID)):       state.Read | state.Write,
		string(storage.BalanceKey(actor, b.Asset)): state.Read | state.Write,
		string(storage.BalanceKey(actor, b.NftID)): state.Read | state.Write,
	}
}

func (*BurnAssetNFT) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.AssetNFTChunks, storage.BalanceChunks, storage.BalanceChunks}
}

func (b *BurnAssetNFT) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, b.Asset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetNonFungibleTokenID {
		return nil, ErrOutputWrongAssetType
	}

	exists, _, _, _, _, _, err = storage.GetAssetNFT(ctx, mu, b.NftID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}

	amountOfToken := uint64(1)
	newSupply, err := smath.Sub(totalSupply, amountOfToken)
	if err != nil {
		return nil, err
	}
	if err := storage.SetAsset(ctx, mu, b.Asset, assetType, name, symbol, decimals, metadata, uri, newSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	if err := storage.DeleteAssetNFT(ctx, mu, b.NftID); err != nil {
		return nil, err
	}
	if err := storage.SubBalance(ctx, mu, actor, b.NftID, 1); err != nil {
		return nil, err
	}

	// Sub balance from collection
	if err := storage.SubBalance(ctx, mu, actor, b.Asset, amountOfToken); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*BurnAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetComputeUnits
}

func (*BurnAssetNFT) Size() int {
	return ids.IDLen * 2
}

func (b *BurnAssetNFT) Marshal(p *codec.Packer) {
	p.PackID(b.Asset)
	p.PackID(b.NftID)
}

func UnmarshalBurnAssetNFT(p *codec.Packer) (chain.Action, error) {
	var burn BurnAssetNFT
	p.UnpackID(false, &burn.Asset)
	p.UnpackID(false, &burn.NftID)
	return &burn, p.Err()
}

func (*BurnAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
