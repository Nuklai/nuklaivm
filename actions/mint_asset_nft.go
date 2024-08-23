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

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*MintAssetNFT)(nil)

type MintAssetNFT struct {
	// To is the recipient of the [Value].
	To codec.Address `json:"to"`

	// Asset is the AssetID(NFT Collection ID) of the asset to mint.
	Asset ids.ID `json:"asset"`

	// Unique ID to assign to the NFT
	UniqueID uint64 `json:"uniqueID"`

	// URI of the NFT
	URI []byte `json:"uri"`
}

func (*MintAssetNFT) GetTypeID() uint8 {
	return nconsts.MintAssetNFTID
}

func (m *MintAssetNFT) StateKeys(codec.Address, ids.ID) state.Keys {
	nftID := nchain.GenerateID(m.Asset, m.UniqueID)
	return state.Keys{
		string(storage.AssetKey(m.Asset)):         state.Read | state.Write,
		string(storage.AssetNFTKey(nftID)):        state.Allocate | state.Write,
		string(storage.BalanceKey(m.To, m.Asset)): state.Allocate | state.Write,
		string(storage.BalanceKey(m.To, nftID)):   state.Allocate | state.Write,
	}
}

func (*MintAssetNFT) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.AssetNFTChunks, storage.BalanceChunks}
}

func (m *MintAssetNFT) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	if m.Asset == ids.Empty {
		return nil, ErrOutputAssetIsNative
	}
	if len(m.URI) < 3 || len(m.URI) > MaxMetadataSize {
		return nil, ErrOutputURIInvalid
	}

	// Check if the nftID already exists
	nftID := nchain.GenerateID(m.Asset, m.UniqueID)
	exists, _, _, _, _, _ := storage.GetAssetNFT(ctx, mu, nftID)
	if exists {
		return nil, ErrOutputNFTAlreadyExists
	}

	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, m.Asset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetNonFungibleTokenID {
		return nil, ErrOutputWrongAssetType
	}
	if mintActor != actor {
		return nil, ErrOutputWrongMintActor
	}
	// Check if the unique ID is greater than or equal to the max supply
	if maxSupply != 0 && m.UniqueID >= maxSupply {
		return nil, ErrOutputIDGreaterThanMaxSupply
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
	totalSupply = newSupply

	if err := storage.SetAssetNFT(ctx, mu, m.Asset, m.UniqueID, nftID, m.URI, m.To); err != nil {
		return nil, err
	}

	if err := storage.SetAsset(ctx, mu, m.Asset, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	if err := storage.AddBalance(ctx, mu, m.To, m.Asset, amountOfToken, true); err != nil {
		return nil, err
	}

	// Add the balance to individual NFT
	if err := storage.AddBalance(ctx, mu, m.To, nftID, amountOfToken, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*MintAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetNFTComputeUnits
}

func (m *MintAssetNFT) Size() int {
	return codec.AddressLen + ids.IDLen + consts.Uint64Len + codec.BytesLen(m.URI)
}

func (m *MintAssetNFT) Marshal(p *codec.Packer) {
	p.PackAddress(m.To)
	p.PackID(m.Asset)
	p.PackUint64(m.UniqueID)
	p.PackBytes(m.URI)
}

func UnmarshalMintAssetNFT(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetNFT
	p.UnpackAddress(&mint.To)
	p.UnpackID(true, &mint.Asset) // empty ID is the native asset
	mint.UniqueID = p.UnpackUint64(false)
	p.UnpackBytes(MaxTextSize, true, &mint.URI)
	return &mint, p.Err()
}

func (*MintAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
