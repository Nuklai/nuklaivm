// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/chain"
	"github.com/nuklai/nuklaivm/storage"
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
	nftID := chain.GenerateIDWithIndex(assetID, 0)

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidAssetType",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: 255, // Invalid type
				Name:      []byte("My Asset"),
				Symbol:    []byte("MYA"),
				Decimals:  0,
				Metadata:  []byte("Metadata"),
				URI:       []byte("uri"),
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputAssetTypeInvalid,
		},
		{
			Name:  "InvalidName",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      []byte("As"), // Invalid name, too short
				Symbol:    []byte("MYA"),
				Decimals:  0,
				Metadata:  []byte("Metadata"),
				URI:       []byte("uri"),
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputNameInvalid,
		},
		{
			Name:  "InvalidSymbol",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      []byte("My Asset"),
				Symbol:    []byte("SY"), // Invalid symbol, too short
				Decimals:  0,
				Metadata:  []byte("Metadata"),
				URI:       []byte("uri"),
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputSymbolInvalid,
		},
		{
			Name:  "InvalidDecimalsForFungible",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      []byte("My Asset"),
				Symbol:    []byte("MYA"),
				Decimals:  19, // Invalid decimals, exceeds max
				Metadata:  []byte("Metadata"),
				URI:       []byte("uri"),
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputDecimalsInvalid,
		},
		{
			Name:  "InvalidDecimalsForNonFungible",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: nconsts.AssetNonFungibleTokenID,
				Name:      []byte("My NFT"),
				Symbol:    []byte("NFT"),
				Decimals:  1, // Invalid decimals for non-fungible token
				Metadata:  []byte("Metadata"),
				URI:       []byte("uri"),
				MaxSupply: 1,
			},
			ExpectedErr: ErrOutputDecimalsInvalid,
		},
		{
			Name:  "InvalidMetadata",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      []byte("My Asset"),
				Symbol:    []byte("MYA"),
				Decimals:  0,
				Metadata:  []byte("Me"), // Invalid metadata, too short
				URI:       []byte("uri"),
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputMetadataInvalid,
		},
		{
			Name:  "InvalidURI",
			Actor: addr,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      []byte("My Asset"),
				Symbol:    []byte("MYA"),
				Decimals:  0,
				Metadata:  []byte("Metadata"),
				URI:       []byte("ur"), // Invalid URI, too short
				MaxSupply: 1000,
			},
			ExpectedErr: ErrOutputURIInvalid,
		},
		{
			Name:     "ValidFungibleTokenCreation",
			Actor:    addr,
			ActionID: assetID,
			Action: &CreateAsset{
				AssetType: nconsts.AssetFungibleTokenID,
				Name:      []byte("My Token"),
				Symbol:    []byte("MYT"),
				Decimals:  9,
				Metadata:  []byte("Metadata"),
				URI:       []byte("uri"),
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
				AssetType: nconsts.AssetNonFungibleTokenID,
				Name:      []byte("My NFT"),
				Symbol:    []byte("NFT"),
				Decimals:  0,
				Metadata:  []byte("NFT Metadata"),
				URI:       []byte("nft-uri"),
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
				AssetType: nconsts.AssetDatasetTokenID,
				Name:      []byte("Dataset Asset"),
				Symbol:    []byte("DAT"),
				Decimals:  0,
				Metadata:  []byte("Dataset Metadata"),
				URI:       []byte("dataset-uri"),
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
			AssetType: nconsts.AssetFungibleTokenID,
			Name:      []byte("Benchmark Asset"),
			Symbol:    []byte("BMA"),
			Decimals:  9,
			Metadata:  []byte("Benchmark Metadata"),
			URI:       []byte("benchmark-uri"),
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
