// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
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

func TestBurnAssetNFTAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	assetID := ids.GenerateTestID()
	nftID := utils.GenerateIDWithIndex(assetID, 0)

	tests := []chaintest.ActionTest{
		{
			Name:  "AssetMissing",
			Actor: addr,
			Action: &BurnAssetNFT{
				AssetAddress: assetID, // NFT collection does not exist
				AssetNftAddress:   nftID,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrAssetMissing,
		},
		{
			Name:  "WrongAssetType",
			Actor: addr,
			Action: &BurnAssetNFT{
				AssetAddress: assetID,
				AssetNftAddress:   nftID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset type to fungible (invalid for NFTs)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("FT"), []byte("FT"), 0, []byte("metadata"), []byte("uri"), 1000, 1000000, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongAssetType,
		},
		{
			Name:  "NFTMissing",
			Actor: addr,
			Action: &BurnAssetNFT{
				AssetAddress: assetID,
				AssetNftAddress:   nftID, // NFT does not exist
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set a valid NFT collection
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT Collection"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 10, 100, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrAssetMissing,
		},
		{
			Name:  "ValidNFTBurn",
			Actor: addr,
			Action: &BurnAssetNFT{
				AssetAddress: assetID,
				AssetNftAddress:   nftID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set a valid NFT collection
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT Collection"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 10, 100, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

				// Set the NFT
				require.NoError(t, storage.SetAssetNFT(context.Background(), store, assetID, 0, nftID, []byte("nft-uri"), []byte("NFT Metadata"), addr))

				// Set balances
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, assetID, 1))
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, nftID, 1))

				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the NFT was removed correctly
				exists, _, _, _, _, _, _ := storage.GetAssetNFT(ctx, store, nftID)
				require.False(t, exists)

				// Check if balances were updated
				balance, err := storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(0), balance)

				// Check if the total supply was reduced
				exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, uint64(9), totalSupply)
			},
			ExpectedOutputs: &BurnAssetNFTResult{
				From:             addr,
				OldBalance:       1,
				NewBalance:       0,
				AssetTotalSupply: 9,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkBurnAssetNFT(b *testing.B) {
	require := require.New(b)
	actor := codec.CreateAddress(0, ids.GenerateTestID())
	assetID := ids.GenerateTestID()
	nftID := utils.GenerateIDWithIndex(assetID, 0)

	burnAssetNFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "BurnAssetNFTBenchmark",
		Actor: actor,
		Action: &BurnAssetNFT{
			AssetAddress: assetID,
			AssetNftAddress:   nftID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set a valid NFT collection and NFT
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("Benchmark NFT Collection"), []byte("BNFT"), 0, []byte("benchmark metadata"), []byte("benchmark-uri"), 1, 10, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

			require.NoError(storage.SetAssetNFT(context.Background(), store, assetID, 0, nftID, []byte("nft-uri"), []byte("NFT Metadata"), actor))

			// Set balances
			require.NoError(storage.SetBalance(context.Background(), store, actor, assetID, 1))
			require.NoError(storage.SetBalance(context.Background(), store, actor, nftID, 1))

			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if NFT was burned correctly
			nftExists, _, _, _, _, _, _ := storage.GetAssetNFT(ctx, store, nftID)
			require.False(nftExists)

			// Check if balance was updated
			balance, err := storage.GetBalance(ctx, store, actor, nftID)
			require.NoError(err)
			require.Equal(b, uint64(0), balance)

			// Check if total supply was reduced
			exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, assetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, uint64(9), totalSupply)
		},
	}

	ctx := context.Background()
	burnAssetNFTActionBenchmark.Run(ctx, b)
}
