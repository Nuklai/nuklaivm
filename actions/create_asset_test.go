// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestCreateAssetAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	assetFT := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), actor)
	assetNFT := storage.AssetAddress(nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)
	assetFractional := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)
	nftAddress := storage.AssetAddressNFT(assetFractional, []byte("metadata"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidAssetType",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: 255, // Invalid type
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  "Metadata",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrAssetTypeInvalid,
		},
		{
			Name:  "InvalidName",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "As", // Invalid name, too short
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  "Metadata",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "InvalidSymbol",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "SY", // Invalid symbol, too short
				Decimals:  0,
				Metadata:  "Metadata",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrSymbolInvalid,
		},
		{
			Name:  "InvalidDecimalsForFungible",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  19, // Invalid decimals, exceeds max
				Metadata:  "Metadata",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrDecimalsInvalid,
		},
		{
			Name:  "InvalidDecimalsForNonFungible",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetNonFungibleTokenID,
				Name:      "My NFT",
				Symbol:    "NFT",
				Decimals:  1, // Invalid decimals for non-fungible token
				Metadata:  "Metadata",
				MaxSupply: 1,
			},
			ExpectedErr: ErrDecimalsInvalid,
		},
		{
			Name:  "InvalidMetadata",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  strings.Repeat("a", storage.MaxAssetMetadataSize+1), // Invalid metadata, too long
				MaxSupply: 1000,
			},
			ExpectedErr: ErrMetadataInvalid,
		},
		{
			Name:  "ValidFungibleTokenCreation",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "name",
				Symbol:    "SYM",
				Decimals:  9,
				Metadata:  "metadata",
				MaxSupply: 1000000,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, assetFT)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetFungibleTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM", string(symbol))
				require.Equal(t, uint8(9), decimals)
				require.Equal(t, "metadata", string(metadata))
				require.Equal(t, assetFT.String(), string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(1000000), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
			},
			ExpectedOutputs: &CreateAssetResult{
				Actor:        actor.String(),
				Receiver:     "",
				AssetAddress: assetFT.String(),
				AssetBalance: 0,
			},
		},
		{
			Name:  "ValidNonFungibleTokenCreation",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetNonFungibleTokenID,
				Name:      "name",
				Symbol:    "SYM",
				Decimals:  0,
				Metadata:  "metadata",
				MaxSupply: 10,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, assetNFT)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetNonFungibleTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, "metadata", string(metadata))
				require.Equal(t, assetNFT.String(), string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(10), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
			},
			ExpectedOutputs: &CreateAssetResult{
				Actor:        actor.String(),
				Receiver:     "",
				AssetAddress: assetNFT.String(),
				AssetBalance: 0,
			},
		},
		{
			Name:  "ValidFractionalTokenCreation",
			Actor: actor,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFractionalTokenID,
				Name:      "name",
				Symbol:    "SYM",
				Decimals:  0,
				Metadata:  "metadata",
				MaxSupply: 100,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, assetFractional)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetFractionalTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, "metadata", string(metadata))
				require.Equal(t, assetFractional.String(), string(uri))
				require.Equal(t, uint64(1), totalSupply)
				require.Equal(t, uint64(100), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
				// Check  balance for addr
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetFractional, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)

				// Check if NFT was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err = storage.GetAssetInfoNoController(ctx, store, nftAddress)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetNonFungibleTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM-0", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, "metadata", string(metadata))
				require.Equal(t, assetFractional.String(), string(uri))
				require.Equal(t, uint64(1), totalSupply)
				require.Equal(t, uint64(1), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
				// Check NFT balance
				balance, err = storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)

				// Check if the NFT has been minted correctly
				require.True(t, storage.AssetExists(ctx, store, nftAddress))
			},
			ExpectedOutputs: &CreateAssetResult{
				Actor:                   actor.String(),
				Receiver:                actor.String(),
				AssetAddress:            assetFractional.String(),
				AssetBalance:            1,
				DatasetParentNftAddress: nftAddress.String(),
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkCreateAsset(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	assetFT := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)

	createAssetActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "CreateAssetBenchmark",
		Actor: actor,
		Action: &CreateAsset{
			AssetType: nconsts.AssetFungibleTokenID,
			Name:      "name",
			Symbol:    "SYM",
			Decimals:  9,
			Metadata:  "metadata",
			MaxSupply: 1000000,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, assetFT, actor, 1000))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the asset was created correctly
			assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, assetFT)
			require.NoError(err)
			require.Equal(nconsts.AssetNonFungibleTokenID, assetType)
			require.Equal("name", string(name))
			require.Equal("SYM", string(symbol))
			require.Equal(uint8(9), decimals)
			require.Equal("metadata", string(metadata))
			require.Equal(assetFT.String(), string(uri))
			require.Equal(uint64(0), totalSupply)
			require.Equal(uint64(1000000), maxSupply)
			require.Equal(actor, owner)
			require.Equal(codec.EmptyAddress, mintAdmin)
			require.Equal(codec.EmptyAddress, pauseUnpauseAdmin)
			require.Equal(codec.EmptyAddress, freezeUnfreezeAdmin)
			require.Equal(codec.EmptyAddress, enableDisableKYCAccountAdmin)
		},
	}

	ctx := context.Background()
	createAssetActionBenchmark.Run(ctx, b)
}
