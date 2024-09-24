// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

const (
	MintAssetFTComputeUnits = 5
)

var (
	ErrOutputAssetIsNative                 = errors.New("asset is native")
	ErrOutputAssetMissing                  = errors.New("asset missing")
	ErrOutputWrongAssetType                = errors.New("asset is not of correct type")
	ErrOutputWrongMintAdmin                = errors.New("mint admin is not correct")
	ErrOutputMaxSupplyReached              = errors.New("max supply reached")
	_                         chain.Action = (*MintAssetFT)(nil)
)

type MintAssetFT struct {
	// AssetID is the AssetID of the asset to mint.
	AssetID ids.ID `serialize:"true" json:"asset_id"`

	// Number of assets to mint to [To].
	Value uint64 `serialize:"true" json:"value"`

	// To is the recipient of the [Value].
	To codec.Address `serialize:"true" json:"to"`
}

func (*MintAssetFT) GetTypeID() uint8 {
	return nconsts.MintAssetFTID
}

func (m *MintAssetFT) StateKeys(codec.Address, ids.ID) state.Keys {
	return state.Keys{
		string(storage.AssetKey(m.AssetID)):         state.Read | state.Write,
		string(storage.BalanceKey(m.To, m.AssetID)): state.All,
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
) (codec.Typed, error) {
	if m.AssetID == ids.Empty {
		return nil, ErrOutputAssetIsNative
	}
	if m.Value == 0 {
		return nil, ErrOutputValueZero
	}
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, mu, m.AssetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetFungibleTokenID {
		return nil, ErrOutputWrongAssetType
	}
	if mintAdmin != actor {
		return nil, ErrOutputWrongMintAdmin
	}

	// Minting logic for fungible tokens
	newSupply, err := smath.Add(totalSupply, m.Value)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}

	if err := storage.SetAsset(ctx, mu, m.AssetID, assetType, name, symbol, decimals, metadata, uri, newSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return nil, err
	}
	newBalance, err := storage.AddBalance(ctx, mu, m.To, m.AssetID, m.Value, true)
	if err != nil {
		return nil, err
	}

	return &MintAssetFTResult{
		To:               m.To.String(),
		OldBalance:       newBalance - m.Value,
		NewBalance:       newBalance,
		AssetTotalSupply: newSupply,
	}, nil
}

func (*MintAssetFT) ComputeUnits(chain.Rules) uint64 {
	return MintAssetFTComputeUnits
}

func (*MintAssetFT) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*MintAssetFT)(nil)

func (*MintAssetFT) Size() int {
	return codec.AddressLen + ids.IDLen + consts.Uint64Len
}

func (m *MintAssetFT) Marshal(p *codec.Packer) {
	p.PackID(m.AssetID)
	p.PackLong(m.Value)
	p.PackAddress(m.To)
}

func UnmarshalMintAssetFT(p *codec.Packer) (chain.Action, error) {
	var mint MintAssetFT
	p.UnpackID(true, &mint.AssetID) // empty ID is the native asset
	mint.Value = p.UnpackUint64(true)
	p.UnpackAddress(&mint.To)
	return &mint, p.Err()
}

var _ codec.Typed = (*MintAssetFTResult)(nil)

type MintAssetFTResult struct {
	To               string `serialize:"true" json:"to"`
	OldBalance       uint64 `serialize:"true" json:"old_balance"`
	NewBalance       uint64 `serialize:"true" json:"new_balance"`
	AssetTotalSupply uint64 `serialize:"true" json:"asset_total_supply"`
}

func (*MintAssetFTResult) GetTypeID() uint8 {
	return nconsts.MintAssetFTID
}
