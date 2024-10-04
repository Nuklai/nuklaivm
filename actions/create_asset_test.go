// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestCreateAssetAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	assetID := ids.GenerateTestID()
	nftID := utils.GenerateIDWithIndex(assetID, 0)

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidAssetType",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: 255, // Invalid type
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  "Metadata",
				URI:       "uri",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputAssetTypeInvalid,
		},
		{
			Name:  "InvalidName",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "As", // Invalid name, too short
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  "Metadata",
				URI:       "uri",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "InvalidSymbol",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "SY", // Invalid symbol, too short
				Decimals:  0,
				Metadata:  "Metadata",
				URI:       "uri",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputSymbolInvalid,
		},
		{
			Name:  "InvalidDecimalsForFungible",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  19, // Invalid decimals, exceeds max
				Metadata:  "Metadata",
				URI:       "uri",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputDecimalsInvalid,
		},
		{
			Name:  "InvalidDecimalsForNonFungible",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetNonFungibleTokenID,
				Name:      "My NFT",
				Symbol:    "NFT",
				Decimals:  1, // Invalid decimals for non-fungible token
				Metadata:  "Metadata",
				URI:       "uri",
				MaxSupply: 1,
			},
			ExpectedErr: ErrOutputDecimalsInvalid,
		},
		{
			Name:  "InvalidMetadata",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  strings.Repeat("a", MaxMetadataSize+1), // Invalid metadata, too long
				URI:       "uri",
				MaxSupply: 1000,
			},
			ExpectedErr: ErrMetadataInvalid,
		},
		{
			Name:  "InvalidURI",
			Actor: addr,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Asset",
				Symbol:    "MYA",
				Decimals:  0,
				Metadata:  "Metadata",
				URI:       strings.Repeat("a", MaxMetadataSize+1), // Invalid URI, too long
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputURIInvalid,
		},
		{
			Name:     "ValidFungibleTokenCreation",
			Actor:    addr,
			ActionID: assetID,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      "My Token",
				Symbol:    "MYT",
				Decimals:  9,
				Metadata:  "Metadata",
				URI:       "uri",
				MaxSupply: 1000000,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was created correctly
				exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, store, assetID)
				require.True(t, exists)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetFungibleTokenID, assetType)
				require.Equal(t, "My Token", string(name))
				require.Equal(t, "MYT", string(symbol))
				require.Equal(t, uint8(9), decimals)
				require.Equal(t, "Metadata", string(metadata))
				require.Equal(t, "uri", string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(1000000), maxSupply)
				require.Equal(t, addr, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
			},
			ExpectedOutputs: &CreateAssetResult{
				AssetID:            assetID,
				AssetBalance:       0,
				DatasetParentNftID: ids.Empty,
			},
		},
		{
			Name:     "ValidNonFungibleTokenCreation",
			Actor:    addr,
			ActionID: assetID,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetNonFungibleTokenID,
				Name:      "My NFT",
				Symbol:    "NFT",
				Decimals:  0,
				Metadata:  "NFT Metadata",
				URI:       "nft-uri",
				MaxSupply: 10,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was created correctly
				exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, store, assetID)
				require.True(t, exists)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetNonFungibleTokenID, assetType)
				require.Equal(t, "My NFT", string(name))
				require.Equal(t, "NFT", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, "NFT Metadata", string(metadata))
				require.Equal(t, "nft-uri", string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(10), maxSupply)
				require.Equal(t, addr, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
			},
			ExpectedOutputs: &CreateAssetResult{
				AssetID:            assetID,
				AssetBalance:       0,
				DatasetParentNftID: ids.Empty,
			},
		},
		{
			Name:     "ValidDatasetCreation",
			Actor:    addr,
			ActionID: assetID,
			Action: &CreateAsset{
				AssetID:   assetID.String(),
				AssetType: nconsts.AssetDatasetTokenID,
				Name:      "Dataset Asset",
				Symbol:    "DAT",
				Decimals:  0,
				Metadata:  "Dataset Metadata",
				URI:       "dataset-uri",
				MaxSupply: 100,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was created correctly
				exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, store, assetID)
				require.True(t, exists)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetDatasetTokenID, assetType)
				require.Equal(t, "Dataset Asset", string(name))
				require.Equal(t, "DAT", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, "Dataset Metadata", string(metadata))
				require.Equal(t, "dataset-uri", string(uri))
				require.Equal(t, uint64(1), totalSupply)
				require.Equal(t, uint64(100), maxSupply)
				require.Equal(t, addr, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
				// Check NFT balance for addr
				balance, err := storage.GetBalance(ctx, store, addr, nftID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
				// Check collectionID balance for addr
				balance, err = storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
				// Check if the NFT has been transferred correctly
				exists, _, _, _, _, owner, _ = storage.GetAssetNFT(ctx, store, nftID)
				require.True(t, exists)
				require.Equal(t, addr.String(), owner.String())
			},
			ExpectedOutputs: &CreateAssetResult{
				AssetID:            assetID,
				AssetBalance:       1,
				DatasetParentNftID: nftID,
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
	assetID := ids.GenerateTestID()

	createAssetActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "CreateAssetBenchmark",
		Actor: actor,
		Action: &CreateAsset{
			AssetID:   assetID.String(),
			AssetType: nconsts.AssetFungibleTokenID,
			Name:      "Benchmark Asset",
			Symbol:    "BMA",
			Decimals:  9,
			Metadata:  "Benchmark Metadata",
			URI:       "benchmark-uri",
			MaxSupply: 1000000,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetBalance(context.Background(), store, actor, assetID, 1000))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the asset was created correctly
			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, store, assetID)
			require.True(exists)
			require.NoError(err)
			require.Equal(nconsts.AssetFungibleTokenID, assetType)
			require.Equal("Benchmark Asset", string(name))
			require.Equal("BMA", string(symbol))
			require.Equal(uint8(9), decimals)
			require.Equal("Benchmark Metadata", string(metadata))
			require.Equal("benchmark-uri", string(uri))
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
