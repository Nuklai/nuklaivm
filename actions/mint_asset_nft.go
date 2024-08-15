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
	Uri []byte `json:"uri"`
}

func (*MintAssetNFT) GetTypeID() uint8 {
	return nconsts.MintAssetNFTID
}

func (m *MintAssetNFT) StateKeys(codec.Address, ids.ID) state.Keys {
	nftID := nchain.GenerateID(m.Asset, m.UniqueID)
	return state.Keys{
		string(storage.AssetKey(m.Asset)):                state.Read | state.Write,
		string(storage.AssetNFTKey(m.Asset, m.UniqueID)): state.All,
		string(storage.BalanceKey(m.To, m.Asset)):        state.All,
		string(storage.BalanceKey(m.To, nftID)):          state.All,
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
	if len(m.Uri) == 0 || len(m.Uri) > MaxTextSize {
		return nil, ErrOutputURIInvalid
	}

	// Check if the unique ID already exists
	exists, _, _, _, _ := storage.GetAssetNFT(ctx, mu, m.Asset, m.UniqueID)
	if exists {
		return nil, ErrOutputNFTAlreadyExists
	}

	exists, name, symbol, decimals, metadata, totalSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor, err := storage.GetAsset(ctx, mu, m.Asset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if mintActor != actor {
		return nil, ErrOutputWrongMintActor
	}
	// Check if the unique ID is greater than or equal to the max supply
	if maxSupply != 0 && m.UniqueID >= maxSupply {
		return nil, ErrOutputIDGreaterThanMaxSupply
	}

	// Minting logic for non-fungible tokens
	newSupply, err := smath.Add64(totalSupply, 1)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}
	totalSupply = newSupply

	nftID := nchain.GenerateID(m.Asset, m.UniqueID)
	if err := storage.SetAssetNFT(ctx, mu, m.Asset, m.UniqueID, nftID, m.Uri, m.To); err != nil {
		return nil, err
	}

	if err := storage.SetAsset(ctx, mu, m.Asset, name, symbol, decimals, metadata, totalSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	if err := storage.AddBalance(ctx, mu, m.To, m.Asset, 1, true); err != nil {
		return nil, err
	}
	// Add the balance to individual NFT
	if err := storage.AddBalance(ctx, mu, m.To, nftID, 1, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func (*MintAssetNFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetNFTComputeUnits
}

func (m *MintAssetNFT) Size() int {
	return codec.AddressLen + ids.IDLen + consts.Uint64Len + codec.BytesLen(m.Uri)
}

func (m *MintAssetNFT) Marshal(p *codec.Packer) {
	p.PackAddress(m.To)
	p.PackID(m.Asset)
	p.PackUint64(m.UniqueID)
	p.PackBytes(m.Uri)
}

func UnmarshalMintAssetNFT(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetNFT
	p.UnpackAddress(&mint.To)
	p.UnpackID(true, &mint.Asset) // empty ID is the native asset
	mint.UniqueID = p.UnpackUint64(false)
	p.UnpackBytes(MaxTextSize, true, &mint.Uri)
	return &mint, p.Err()
}

func (*MintAssetNFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
