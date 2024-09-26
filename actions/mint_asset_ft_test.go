// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestMintAssetFTAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	assetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "NativeAssetMint",
			Actor: addr,
			Action: &MintAssetFT{
				AssetID: ids.Empty, // Native asset, invalid for minting
				Value:   1000,
				To:      addr,
			},
			ExpectedErr: ErrOutputAssetIsNative,
		},
		{
			Name:  "ZeroValueMint",
			Actor: addr,
			Action: &MintAssetFT{
				AssetID: assetID,
				Value:   0, // Invalid zero value
				To:      addr,
			},
			ExpectedErr: ErrOutputValueZero,
		},
		{
			Name:  "AssetMissing",
			Actor: addr,
			Action: &MintAssetFT{
				AssetID: assetID, // Asset does not exist
				Value:   1000,
				To:      addr,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrOutputAssetMissing,
		},
		{
			Name:  "WrongAssetType",
			Actor: addr,
			Action: &MintAssetFT{
				AssetID: assetID,
				Value:   1000,
				To:      addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset type as non-fungible (invalid for FT minting)
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 0, 1, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongAssetType,
		},
		{
			Name:  "WrongMintAdmin",
			Actor: codec.CreateAddress(0, ids.GenerateTestID()), // Not the mint admin
			Action: &MintAssetFT{
				AssetID: assetID,
				Value:   1000,
				To:      addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Token"), []byte("TKN"), 9, []byte("metadata"), []byte("uri"), 0, 1000000, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongMintAdmin,
		},
		{
			Name:  "ExceedMaxSupply",
			Actor: addr,
			Action: &MintAssetFT{
				AssetID: assetID,
				Value:   1000001, // Exceeds max supply
				To:      addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details with max supply 1,000,000
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Token"), []byte("TKN"), 9, []byte("metadata"), []byte("uri"), 0, 1000000, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputMaxSupplyReached,
		},
		{
			Name:  "ValidMint",
			Actor: addr,
			Action: &MintAssetFT{
				AssetID: assetID,
				Value:   1000,
				To:      addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details with max supply 1,000,000
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Token"), []byte("TKN"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check new asset supply and balance
				exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAsset(ctx, store, assetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, uint64(6000), totalSupply)

				// Check if balance updated
				balance, err := storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(1000), balance)
			},
			ExpectedOutputs: &MintAssetFTResult{
				To:               addr.String(),
				OldBalance:       0,
				NewBalance:       1000,
				AssetTotalSupply: 6000,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkMintAssetFT(b *testing.B) {
	require := require.New(b)
	actor := codec.CreateAddress(0, ids.GenerateTestID())
	assetID := ids.GenerateTestID()

	mintAssetFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "MintAssetFTBenchmark",
		Actor: actor,
		Action: &MintAssetFT{
			AssetID: assetID,
			Value:   1000,
			To:      actor,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Benchmark Token"), []byte("BMT"), 9, []byte("benchmark metadata"), []byte("benchmark-uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check new asset supply and balance
			exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAsset(ctx, store, assetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, uint64(6000), totalSupply)

			// Check if balance updated
			balance, err := storage.GetBalance(ctx, store, actor, assetID)
			require.NoError(err)
			require.Equal(b, uint64(1000), balance)
		},
	}

	ctx := context.Background()
	mintAssetFTActionBenchmark.Run(ctx, b)
}
