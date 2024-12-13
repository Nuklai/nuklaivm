// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestBurnAssetFTAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "ZeroValueBurn",
			Actor: actor,
			Action: &BurnAssetFT{
				AssetAddress: assetAddress,
				Value:        0, // Invalid zero value
			},
			ExpectedErr: ErrValueZero,
		},
		{
			Name:  "WrongAssetType",
			Actor: actor,
			Action: &BurnAssetFT{
				AssetAddress: assetAddress,
				Value:        1000,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset type as non-fungible (invalid for FT burning)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrAssetTypeInvalid,
		},
		{
			Name:  "InsufficientBalanceBurn",
			Actor: actor,
			Action: &BurnAssetFT{
				AssetAddress: assetAddress,
				Value:        1000, // Trying to burn more than available balance
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset with balance of 500
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 500))
				return store
			}(),
			ExpectedErr: storage.ErrInsufficientAssetBalance,
		},
		{
			Name:  "ValidBurn",
			Actor: actor,
			Action: &BurnAssetFT{
				AssetAddress: assetAddress,
				Value:        500,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset with balance of 1000
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 1000))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check new asset supply and balance
				_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
				require.NoError(t, err)
				require.Equal(t, uint64(4500), totalSupply)

				// Check if balance updated
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(500), balance)
			},
			ExpectedOutputs: &BurnAssetFTResult{
				Actor:      actor.String(),
				Receiver:   "",
				OldBalance: 1000,
				NewBalance: 500,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkBurnAssetFT(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), actor)

	burnAssetFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "BurnAssetFTBenchmark",
		Actor: actor,
		Action: &BurnAssetFT{
			AssetAddress: assetAddress,
			Value:        500,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Setting asset with balance of 1000
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 1000))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check new asset supply and balance
			_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
			require.NoError(err)
			require.Equal(uint64(4500), totalSupply)

			// Check if balance updated
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
			require.NoError(err)
			require.Equal(uint64(500), balance)
		},
	}

	ctx := context.Background()
	burnAssetFTActionBenchmark.Run(ctx, b)
}
