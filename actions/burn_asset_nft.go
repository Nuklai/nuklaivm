// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	BurnAssetNFTComputeUnits = 1
)

var _ chain.Action = (*BurnAssetNFT)(nil)

type BurnAssetNFT struct {
	// AssetID ID of the asset to burn(this is the nft collection ID)
	AssetID ids.ID `serialize:"true" json:"asset_id"`

	// NFT ID of the asset to burn
	NftID ids.ID `serialize:"true" json:"nftID"`
}

func (*BurnAssetNFT) GetTypeID() uint8 {
	return nconsts.BurnAssetNFTID
}

func (b *BurnAssetNFT) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(b.AssetID)):          state.Read | state.Write,
		string(storage.AssetNFTKey(b.NftID)):         state.Read | state.Write,
		string(storage.BalanceKey(actor, b.AssetID)): state.Read | state.Write,
		string(storage.BalanceKey(actor, b.NftID)):   state.Read | state.Write,
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
) (codec.Typed, error) {
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, mu, b.AssetID)
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
	if err := storage.SetAsset(ctx, mu, b.AssetID, assetType, name, symbol, decimals, metadata, uri, newSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	// Sub balance from individual NFT
	if _, err := storage.SubBalance(ctx, mu, actor, b.NftID, 1); err != nil {
		return nil, err
	}

	// Sub balance from collection
	newBalance, err := storage.SubBalance(ctx, mu, actor, b.AssetID, amountOfToken)
	if err != nil {
		return nil, err
	}

	if err := storage.DeleteAssetNFT(ctx, mu, b.NftID); err != nil {
		return nil, err
	}

	return &BurnAssetNFTResult{
		From:             actor,
		OldBalance:       newBalance + amountOfToken,
		NewBalance:       newBalance,
		AssetTotalSupply: newSupply,
	}, nil
}

func (*BurnAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetFTComputeUnits
}

func (*BurnAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*BurnAssetNFT)(nil)

func (*BurnAssetNFT) Size() int {
	return ids.IDLen * 2
}

func (b *BurnAssetNFT) Marshal(p *codec.Packer) {
	p.PackID(b.AssetID)
	p.PackID(b.NftID)
}

func UnmarshalBurnAssetNFT(p *codec.Packer) (chain.Action, error) {
	var burn BurnAssetNFT
	p.UnpackID(false, &burn.AssetID)
	p.UnpackID(false, &burn.NftID)
	return &burn, p.Err()
}

var (
	_ codec.Typed     = (*BurnAssetNFTResult)(nil)
	_ chain.Marshaler = (*BurnAssetNFTResult)(nil)
)

type BurnAssetNFTResult struct {
	From             codec.Address `serialize:"true" json:"from"`
	OldBalance       uint64        `serialize:"true" json:"old_balance"`
	NewBalance       uint64        `serialize:"true" json:"new_balance"`
	AssetTotalSupply uint64        `serialize:"true" json:"asset_total_supply"`
}

func (*BurnAssetNFTResult) GetTypeID() uint8 {
	return nconsts.BurnAssetNFTID
}

func (*BurnAssetNFTResult) Size() int {
	return codec.AddressLen + consts.Uint64Len*3
}

func (r *BurnAssetNFTResult) Marshal(p *codec.Packer) {
	p.PackAddress(r.From)
	p.PackUint64(r.OldBalance)
	p.PackUint64(r.NewBalance)
	p.PackUint64(r.AssetTotalSupply)
}

func UnmarshalBurnAssetNFTResult(p *codec.Packer) (codec.Typed, error) {
	var result BurnAssetNFTResult
	p.UnpackAddress(&result.From)
	result.OldBalance = p.UnpackUint64(false)
	result.NewBalance = p.UnpackUint64(false)
	result.AssetTotalSupply = p.UnpackUint64(false)
	return &result, p.Err()
}
