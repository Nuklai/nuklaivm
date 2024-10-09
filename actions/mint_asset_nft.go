// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	MintAssetNFTComputeUnits = 5
)

var (
	ErrNFTAlreadyExists                      = errors.New("NFT already exists")
	ErrCantFractionalizeFurther              = errors.New("asset already has a parent")
	_                           chain.Action = (*MintAssetNFT)(nil)
)

type MintAssetNFT struct {
	// AssetAddress is the AssetAddress(NFT Collection ID) of the asset to mint.
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

	// Metadata of the NFT
	Metadata string `serialize:"true" json:"metadata"`

	// To is the recipient of the [Value].
	To codec.Address `serialize:"true" json:"to"`
}

func (*MintAssetNFT) GetTypeID() uint8 {
	return nconsts.MintAssetNFTID
}

func (m *MintAssetNFT) StateKeys(codec.Address) state.Keys {
	nftAddress := storage.AssetAddressNFT(m.AssetAddress, []byte(m.Metadata), m.To)
	return state.Keys{
		string(storage.AssetInfoKey(m.AssetAddress)):                 state.Read | state.Write,
		string(storage.AssetInfoKey(nftAddress)):                     state.All,
		string(storage.AssetAccountBalanceKey(m.AssetAddress, m.To)): state.All,
		string(storage.AssetAccountBalanceKey(nftAddress, m.To)):     state.All,
	}
}

func (m *MintAssetNFT) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	if len(m.Metadata) > storage.MaxAssetMetadataSize {
		return nil, ErrMetadataInvalid
	}

	assetType, name, symbol, _, metadata, uri, totalSupply, _, _, mintAdmin, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, m.AssetAddress)
	if err != nil {
		return nil, err
	}
	// Ensure that it's a non-fungible token
	if assetType != nconsts.AssetNonFungibleTokenID {
		return nil, ErrAssetTypeInvalid
	}
	if mintAdmin != actor {
		return nil, ErrWrongMintAdmin
	}
	// Ensure that m.AssetAddress is not the same as uri
	if m.AssetAddress.String() != string(uri) {
		return nil, ErrCantFractionalizeFurther
	}

	// Check if the nftAddress already exists
	nftAddress := storage.AssetAddressNFT(m.AssetAddress, []byte(m.Metadata), m.To)
	if storage.AssetExists(ctx, mu, nftAddress) {
		return nil, ErrNFTAlreadyExists
	}

	// Minting logic for non-fungible tokens
	newBalance, err := storage.MintAsset(ctx, mu, m.AssetAddress, m.To, 1)
	if err != nil {
		return nil, err
	}
	symbol = utils.CombineWithSuffix(symbol, totalSupply, storage.MaxSymbolSize)
	if err := storage.SetAssetInfo(ctx, mu, nftAddress, assetType, name, symbol, 0, metadata, []byte(m.AssetAddress.String()), 0, 1, m.To, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}
	if _, err := storage.MintAsset(ctx, mu, nftAddress, m.To, 1); err != nil {
		return nil, err
	}

	return &MintAssetNFTResult{
		AssetNftAddress: nftAddress.String(),
		OldBalance:      newBalance - 1,
		NewBalance:      newBalance,
	}, nil
}

func (*MintAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetNFTComputeUnits
}

func (*MintAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalMintAssetNFT(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetNFT
	p.UnpackAddress(&mint.AssetAddress)
	p.UnpackString(false)
	p.UnpackAddress(&mint.To)
	return &mint, p.Err()
}

var (
	_ codec.Typed     = (*MintAssetNFTResult)(nil)
	_ chain.Marshaler = (*MintAssetNFTResult)(nil)
)

type MintAssetNFTResult struct {
	AssetNftAddress string `serialize:"true" json:"asset_nft_address"`
	OldBalance      uint64        `serialize:"true" json:"old_balance"`
	NewBalance      uint64        `serialize:"true" json:"new_balance"`
}

func (*MintAssetNFTResult) GetTypeID() uint8 {
	return nconsts.MintAssetNFTID
}

func (m *MintAssetNFTResult) Size() int {
	return codec.StringLen(m.AssetNftAddress) + consts.Uint64Len*2
}

func (r *MintAssetNFTResult) Marshal(p *codec.Packer) {
	p.PackString(r.AssetNftAddress)
	p.PackUint64(r.OldBalance)
	p.PackUint64(r.NewBalance)
}

func UnmarshalMintAssetNFTResult(p *codec.Packer) (codec.Typed, error) {
	var result MintAssetNFTResult
	result.AssetNftAddress = p.UnpackString(true)
	result.OldBalance = p.UnpackUint64(false)
	result.NewBalance = p.UnpackUint64(true)
	return &result, p.Err()
}
