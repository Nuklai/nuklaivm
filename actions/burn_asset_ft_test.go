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
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestBurnAssetFTAction(t *testing.T) {
	addr := codec.CreateAddress(0, ids.GenerateTestID())
	assetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "ZeroValueBurn",
			Actor: addr,
			Action: &BurnAssetFT{
				AssetAddress: assetID,
				Value:   0, // Invalid zero value
			},
			ExpectedErr: ErrValueZero,
		},
		{
			Name:  "AssetMissing",
			Actor: addr,
			Action: &BurnAssetFT{
				AssetAddress: assetID, // Asset does not exist
				Value:   1000,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrAssetMissing,
		},
		{
			Name:  "WrongAssetType",
			Actor: addr,
			Action: &BurnAssetFT{
				AssetAddress: assetID,
				Value:   1000,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset type as non-fungible (invalid for FT burning)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 0, 1, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongAssetType,
		},
		{
			Name:  "InsufficientBalanceBurn",
			Actor: addr,
			Action: &BurnAssetFT{
				AssetAddress: assetID,
				Value:   1000, // Trying to burn more than available balance
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset with balance of 500
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Token"), []byte("TKN"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, assetID, 500))
				return store
			}(),
			ExpectedErr: storage.ErrInvalidBalance,
		},
		{
			Name:  "ValidBurn",
			Actor: addr,
			Action: &BurnAssetFT{
				AssetAddress: assetID,
				Value:   500,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset with balance of 1000
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Token"), []byte("TKN"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, assetID, 1000))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check new asset supply and balance
				exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, uint64(4500), totalSupply)

				// Check if balance updated
				balance, err := storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(500), balance)
			},
			ExpectedOutputs: &BurnAssetFTResult{
				From:             addr,
				OldBalance:       1000,
				NewBalance:       500,
				AssetTotalSupply: 4500,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkBurnAssetFT(b *testing.B) {
	require := require.New(b)
	actor := codec.CreateAddress(0, ids.GenerateTestID())
	assetID := ids.GenerateTestID()

	burnAssetFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "BurnAssetFTBenchmark",
		Actor: actor,
		Action: &BurnAssetFT{
			AssetAddress: assetID,
			Value:   500,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Benchmark Token"), []byte("BMT"), 9, []byte("benchmark metadata"), []byte("benchmark-uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			require.NoError(storage.SetBalance(context.Background(), store, actor, assetID, 1000))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check new asset supply and balance
			exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, uint64(4500), totalSupply)

			// Check if balance updated
			balance, err := storage.GetBalance(ctx, store, actor, assetID)
			require.NoError(err)
			require.Equal(b, uint64(500), balance)
		},
	}

	ctx := context.Background()
	burnAssetFTActionBenchmark.Run(ctx, b)
}
