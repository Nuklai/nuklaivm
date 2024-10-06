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

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	BurnAssetNFTComputeUnits = 1
)

var _ chain.Action = (*BurnAssetNFT)(nil)

type BurnAssetNFT struct {
	// AssetAddress of the asset to burn(this is the nft collection address)
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

	// AssetNftAddress  of the asset to burn
	AssetNftAddress codec.Address `serialize:"true" json:"asset_nft_address"`
}

func (*BurnAssetNFT) GetTypeID() uint8 {
	return nconsts.BurnAssetNFTID
}

func (b *BurnAssetNFT) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(b.AssetAddress)):      state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(b.AssetAddress, actor)): state.Read | state.Write,
		string(storage.AssetInfoKey(b.AssetNftAddress)):      state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(b.AssetNftAddress, actor)): state.Read | state.Write,
	}
}

func (b *BurnAssetNFT) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	assetType, _, _, _, _, _, _, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, b.AssetAddress)
	if err != nil {
		return nil, err
	}
	// Ensure that it's a non-fungible token
	if assetType != nconsts.AssetNonFungibleTokenID {
		return nil, ErrAssetTypeInvalid
	}
	// Ensure the asset exists
	if !storage.AssetExists(ctx, mu, b.AssetNftAddress) {
		return nil, ErrAssetNotFound
	}

	// Burning logic for non-fungible tokens
	newBalance, err := storage.BurnAsset(ctx, mu, b.AssetAddress, actor, 1)
	if err != nil {
		return nil, err
	}
	if err := storage.DeleteAsset(ctx, mu, b.AssetNftAddress); err != nil {
		return nil, err
	}

	return &BurnAssetNFTResult{
		OldBalance:       newBalance + 1,
		NewBalance:       newBalance,
	}, nil
}

func (*BurnAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return BurnAssetFTComputeUnits
}

func (*BurnAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalBurnAssetNFT(p *codec.Packer) (chain.Action, error) {
	var burn BurnAssetNFT
	p.UnpackAddress(&burn.AssetAddress)
	p.UnpackAddress(&burn.AssetNftAddress)
	return &burn, p.Err()
}

var (
	_ codec.Typed     = (*BurnAssetNFTResult)(nil)
	_ chain.Marshaler = (*BurnAssetNFTResult)(nil)
)

type BurnAssetNFTResult struct {
	OldBalance       uint64        `serialize:"true" json:"old_balance"`
	NewBalance       uint64        `serialize:"true" json:"new_balance"`
}

func (*BurnAssetNFTResult) GetTypeID() uint8 {
	return nconsts.BurnAssetNFTID
}

func (*BurnAssetNFTResult) Size() int {
	return consts.Uint64Len*2
}

func (r *BurnAssetNFTResult) Marshal(p *codec.Packer) {
	p.PackUint64(r.OldBalance)
	p.PackUint64(r.NewBalance)
}

func UnmarshalBurnAssetNFTResult(p *codec.Packer) (codec.Typed, error) {
	var result BurnAssetNFTResult
	result.OldBalance = p.UnpackUint64(true)
	result.NewBalance = p.UnpackUint64(false)
	return &result, p.Err()
}
