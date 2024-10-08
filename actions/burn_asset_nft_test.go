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

func TestBurnAssetNFTAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)
	nftAddress := storage.AssetAddressNFT(assetAddress, []byte("metadata"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "WrongAssetType",
			Actor: actor,
			Action: &BurnAssetNFT{
				AssetAddress:    assetAddress,
				AssetNftAddress: nftAddress,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset type to fungible (invalid for NFTs)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, nftAddress, actor, 1))
				return store
			}(),
			ExpectedErr: ErrAssetTypeInvalid,
		},
		{
			Name:  "CantBurnNFTCollection",
			Actor: actor,
			Action: &BurnAssetNFT{
				AssetAddress:    assetAddress,
				AssetNftAddress: assetAddress,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset type to fungible (invalid for NFTs)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 1))
				return store
			}(),
			ExpectedErr: ErrNFTDoesNotBelongToTheCollection,
		},
		{
			Name:  "ValidNFTBurn",
			Actor: actor,
			Action: &BurnAssetNFT{
				AssetAddress:    assetAddress,
				AssetNftAddress: nftAddress,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set a valid NFT collection
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

				// Set the NFT
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, nftAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

				// Set balances
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 1))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, nftAddress, actor, 0))

				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the NFT was removed correctly
				require.False(t, storage.AssetExists(ctx, store, nftAddress))
				// Ensure NFT collection still exists
				require.True(t, storage.AssetExists(ctx, store, assetAddress))

				// Check if balances were updated
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(0), balance)

				// Check if the total supply was reduced
				_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
				require.NoError(t, err)
				require.Equal(t, uint64(0), totalSupply)
			},
			ExpectedOutputs: &BurnAssetNFTResult{
				OldBalance: 1,
				NewBalance: 0,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkBurnAssetNFT(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)
	nftAddress := storage.AssetAddressNFT(assetAddress, []byte("metadata"), actor)

	burnAssetNFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "BurnAssetNFTBenchmark",
		Actor: actor,
		Action: &BurnAssetNFT{
			AssetAddress:    assetAddress,
			AssetNftAddress: nftAddress,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set a valid NFT collection
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

			// Set the NFT
			require.NoError(storage.SetAssetInfo(context.Background(), store, nftAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

			// Set balances
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 1))
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, nftAddress, actor, 0))

			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the NFT was removed correctly
			require.False(storage.AssetExists(ctx, store, nftAddress))
			// Ensure NFT collection still exists
			require.True(storage.AssetExists(ctx, store, assetAddress))

			// Check if balances were updated
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
			require.NoError(err)
			require.Equal(uint64(0), balance)

			// Check if the total supply was reduced
			_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
			require.NoError(err)
			require.Equal(uint64(0), totalSupply)
		},
	}

	ctx := context.Background()
	burnAssetNFTActionBenchmark.Run(ctx, b)
}
