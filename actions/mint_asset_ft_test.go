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
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "NativeAssetMint",
			Actor: actor,
			Action: &MintAssetFT{
				AssetAddress: storage.NAIAddress, // Native asset, invalid for minting
				Value:        1000,
				To:           actor,
			},
			ExpectedErr: ErrAssetIsNative,
		},
		{
			Name:  "ZeroValueMint",
			Actor: actor,
			Action: &MintAssetFT{
				AssetAddress: assetAddress,
				Value:        0, // Invalid zero value
				To:           actor,
			},
			ExpectedErr: ErrValueZero,
		},
		{
			Name:  "WrongAssetType",
			Actor: actor,
			Action: &MintAssetFT{
				AssetAddress: assetAddress,
				Value:        1000,
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset type as non-fungible (invalid for FT minting)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrAssetTypeInvalid,
		},
		{
			Name:  "WrongMintAdmin",
			Actor: codec.CreateAddress(0, ids.GenerateTestID()), // Not the mint admin
			Action: &MintAssetFT{
				AssetAddress: assetAddress,
				Value:        1000,
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrWrongMintAdmin,
		},
		{
			Name:  "ExceedMaxSupply",
			Actor: actor,
			Action: &MintAssetFT{
				AssetAddress: assetAddress,
				Value:        1000001, // Exceeds max supply
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details with max supply 1,000,000
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: storage.ErrMaxSupplyExceeded,
		},
		{
			Name:  "ValidMint",
			Actor: actor,
			Action: &MintAssetFT{
				AssetAddress: assetAddress,
				Value:        1000,
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details with max supply 1,000,000
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check new asset supply and balance
				_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
				require.NoError(t, err)
				require.Equal(t, uint64(6000), totalSupply)

				// Check if balance updated
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(1000), balance)
			},
			ExpectedOutputs: &MintAssetFTResult{
				OldBalance: 0,
				NewBalance: 1000,
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
	assetAddress := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), actor)

	mintAssetFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "MintAssetFTBenchmark",
		Actor: actor,
		Action: &MintAssetFT{
			AssetAddress: assetAddress,
			Value:        1000,
			To:           actor,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 5000, 1000000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check new asset supply and balance
			_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
			require.NoError(err)
			require.Equal(b, uint64(6000), totalSupply)

			// Check if balance updated
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
			require.NoError(err)
			require.Equal(b, uint64(1000), balance)
		},
	}

	ctx := context.Background()
	mintAssetFTActionBenchmark.Run(ctx, b)
}
