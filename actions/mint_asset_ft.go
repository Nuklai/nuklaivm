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

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*MintAssetFT)(nil)

type MintAssetFT struct {
	// To is the recipient of the [Value].
	To codec.Address `json:"to"`

	// Asset is the AssetID of the asset to mint.
	Asset ids.ID `json:"asset"`

	// Number of assets to mint to [To].
	Value uint64 `json:"value"`

	// TODO: add boolean to indicate whether sender will
	// create recipient account
}

func (*MintAssetFT) GetTypeID() uint8 {
	return nconsts.MintAssetFTID
}

func (m *MintAssetFT) StateKeys(codec.Address, ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(m.Asset)):         state.Read | state.Write,
		string(storage.BalanceKey(m.To, m.Asset)): state.All,
	}
}

func (*MintAssetFT) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.AssetChunks, storage.BalanceChunks}
}

func (m *MintAssetFT) Execute(
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
	if m.Value == 0 {
		return nil, ErrOutputValueZero
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

	// Minting logic for fungible tokens
	newSupply, err := smath.Add64(totalSupply, m.Value)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}
	totalSupply = newSupply

	if err := storage.SetAsset(ctx, mu, m.Asset, name, symbol, decimals, metadata, totalSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor); err != nil {
		return nil, err
	}
	if err := storage.AddBalance(ctx, mu, m.To, m.Asset, m.Value, true); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*MintAssetFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetComputeUnits
}

func (*MintAssetFT) Size() int {
	return codec.AddressLen + ids.IDLen + consts.Uint64Len
}

func (m *MintAssetFT) Marshal(p *codec.Packer) {
	p.PackAddress(m.To)
	p.PackID(m.Asset)
	p.PackUint64(m.Value)
}

func UnmarshalMintAsset(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetFT
	p.UnpackAddress(&mint.To)
	p.UnpackID(true, &mint.Asset) // empty ID is the native asset
	mint.Value = p.UnpackUint64(true)
	return &mint, p.Err()
}

func (*MintAssetFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
