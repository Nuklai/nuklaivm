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

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	MintAssetNFTComputeUnits = 5
)

var (
	ErrOutputNFTAlreadyExists                          = errors.New("NFT already exists")
	ErrOutputUniqueIDGreaterThanMaxSupply              = errors.New("unique ID is greater than or equal to max supply")
	_                                     chain.Action = (*MintAssetNFT)(nil)
)

type MintAssetNFT struct {
	// AssetID is the AssetID(NFT Collection ID) of the asset to mint.
	AssetID ids.ID `serialize:"true" json:"asset_id"`

	// Unique ID to assign to the NFT
	UniqueID uint64 `serialize:"true" json:"unique_id"`

	// URI of the NFT
	URI []byte `serialize:"true" json:"uri"`

	// Metadata of the NFT
	Metadata []byte `serialize:"true" json:"metadata"`

	// To is the recipient of the [Value].
	To codec.Address `serialize:"true" json:"to"`
}

func (*MintAssetNFT) GetTypeID() uint8 {
	return nconsts.MintAssetNFTID
}

func (m *MintAssetNFT) StateKeys(codec.Address) state.Keys {
	nftID := utils.GenerateIDWithIndex(m.AssetID, m.UniqueID)
	return state.Keys{
		string(storage.AssetInfoKey(m.AssetID)):     state.Read | state.Write,
		string(storage.AssetNFTKey(nftID)):          state.Allocate | state.Write,
		string(storage.BalanceKey(m.To, m.AssetID)): state.Allocate | state.Write,
		string(storage.BalanceKey(m.To, nftID)):     state.Allocate | state.Write,
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
	if m.AssetID == ids.Empty {
		return nil, ErrOutputAssetIsNative
	}
	if len(m.URI) < 3 || len(m.URI) > MaxMetadataSize {
		return nil, ErrOutputURIInvalid
	}
	if len(m.Metadata) < 3 || len(m.Metadata) > MaxMetadataSize {
		return nil, ErrMetadataInvalid
	}

	// Check if the nftID already exists
	nftID := utils.GenerateIDWithIndex(m.AssetID, m.UniqueID)
	exists, _, _, _, _, _, _ := storage.GetAssetNFT(ctx, mu, nftID)
	if exists {
		return nil, ErrOutputNFTAlreadyExists
	}

	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, mu, m.AssetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetNonFungibleTokenID {
		return nil, ErrOutputWrongAssetType
	}
	if mintAdmin != actor {
		return nil, ErrOutputWrongMintAdmin
	}
	// Check if the unique ID is greater than or equal to the max supply
	if maxSupply != 0 && m.UniqueID >= maxSupply {
		return nil, ErrOutputUniqueIDGreaterThanMaxSupply
	}

	// Minting logic for non-fungible tokens
	amountOfToken := uint64(1)
	newSupply, err := smath.Add64(totalSupply, amountOfToken)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}

	if err := storage.SetAssetNFT(ctx, mu, m.AssetID, m.UniqueID, nftID, m.URI, m.Metadata, m.To); err != nil {
		return nil, err
	}

	if err := storage.SetAssetInfo(ctx, mu, m.AssetID, assetType, name, symbol, decimals, metadata, uri, newSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	newBalance, err := storage.AddBalance(ctx, mu, m.To, m.AssetID, amountOfToken, true)
	if err != nil {
		return nil, err
	}

	// Add the balance to individual NFT
	if _, err := storage.AddBalance(ctx, mu, m.To, nftID, amountOfToken, true); err != nil {
		return nil, err
	}

	return &MintAssetNFTResult{
		NftID:            nftID,
		To:               m.To,
		OldBalance:       newBalance - amountOfToken,
		NewBalance:       newBalance,
		AssetTotalSupply: newSupply,
	}, nil
}

func (*MintAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetNFTComputeUnits
}

func (*MintAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*MintAssetNFT)(nil)

func (m *MintAssetNFT) Size() int {
	return ids.IDLen + consts.Uint64Len + codec.BytesLen(m.URI) + codec.BytesLen(m.Metadata) + codec.AddressLen
}

func (m *MintAssetNFT) Marshal(p *codec.Packer) {
	p.PackID(m.AssetID)
	p.PackUint64(m.UniqueID)
	p.PackBytes(m.URI)
	p.PackBytes(m.Metadata)
	p.PackAddress(m.To)
}

func UnmarshalMintAssetNFT(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetNFT
	p.UnpackID(true, &mint.AssetID) // empty ID is the native asset
	mint.UniqueID = p.UnpackUint64(false)
	p.UnpackBytes(MaxMetadataSize, true, &mint.URI)
	p.UnpackBytes(MaxMetadataSize, true, &mint.Metadata)
	p.UnpackAddress(&mint.To)
	return &mint, p.Err()
}

var (
	_ codec.Typed     = (*MintAssetNFTResult)(nil)
	_ chain.Marshaler = (*MintAssetNFTResult)(nil)
)

type MintAssetNFTResult struct {
	NftID            ids.ID        `serialize:"true" json:"nft_id"`
	To               codec.Address `serialize:"true" json:"to"`
	OldBalance       uint64        `serialize:"true" json:"old_balance"`
	NewBalance       uint64        `serialize:"true" json:"new_balance"`
	AssetTotalSupply uint64        `serialize:"true" json:"asset_total_supply"`
}

func (*MintAssetNFTResult) GetTypeID() uint8 {
	return nconsts.MintAssetNFTID
}

func (*MintAssetNFTResult) Size() int {
	return ids.IDLen + codec.AddressLen + consts.Uint64Len*3
}

func (r *MintAssetNFTResult) Marshal(p *codec.Packer) {
	p.PackID(r.NftID)
	p.PackAddress(r.To)
	p.PackUint64(r.OldBalance)
	p.PackUint64(r.NewBalance)
	p.PackUint64(r.AssetTotalSupply)
}

func UnmarshalMintAssetNFTResult(p *codec.Packer) (codec.Typed, error) {
	var result MintAssetNFTResult
	p.UnpackID(true, &result.NftID)
	p.UnpackAddress(&result.To)
	result.OldBalance = p.UnpackUint64(false)
	result.NewBalance = p.UnpackUint64(false)
	result.AssetTotalSupply = p.UnpackUint64(false)
	return &result, p.Err()
}
