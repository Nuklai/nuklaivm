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

func TestMintAssetNFTAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	assetID := ids.GenerateTestID()
	nftID := utils.GenerateIDWithIndex(assetID, 0)

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidURI",
			Actor: addr,
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 0,
				URI:      []byte("u"), // Invalid URI (too short)
				Metadata: []byte("NFT Metadata"),
				To:       addr,
			},
			ExpectedErr: ErrOutputURIInvalid,
		},
		{
			Name:  "InvalidMetadata",
			Actor: addr,
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 0,
				URI:      []byte("nft-uri"),
				Metadata: []byte("m"), // Invalid metadata (too short)
				To:       addr,
			},
			ExpectedErr: ErrOutputMetadataInvalid,
		},
		{
			Name:  "NFTAlreadyExists",
			Actor: addr,
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 0,
				URI:      []byte("nft-uri"),
				Metadata: []byte("NFT Metadata"),
				To:       addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set an existing NFT in storage
				require.NoError(t, storage.SetAssetNFT(context.Background(), store, assetID, 0, nftID, []byte("nft-uri"), []byte("NFT Metadata"), addr))
				return store
			}(),
			ExpectedErr: ErrOutputNFTAlreadyExists,
		},
		{
			Name:  "WrongAssetType",
			Actor: addr,
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 0,
				URI:      []byte("nft-uri"),
				Metadata: []byte("NFT Metadata"),
				To:       addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset type to fungible (invalid for NFT minting)
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("Fungible"), []byte("FT"), 0, []byte("metadata"), []byte("uri"), 0, 100, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongAssetType,
		},
		{
			Name:  "WrongMintAdmin",
			Actor: codec.CreateAddress(0, ids.GenerateTestID()), // Not the mint admin
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 0,
				URI:      []byte("nft-uri"),
				Metadata: []byte("NFT Metadata"),
				To:       addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset with a different mint admin
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT Collection"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 0, 100, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongMintAdmin,
		},
		{
			Name:  "ExceedMaxSupply",
			Actor: addr,
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 100, // Exceeds max supply
				URI:      []byte("nft-uri"),
				Metadata: []byte("NFT Metadata"),
				To:       addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with max supply 100
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT Collection"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 0, 100, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputUniqueIDGreaterThanMaxSupply,
		},
		{
			Name:  "ValidNFTMint",
			Actor: addr,
			Action: &MintAssetNFT{
				AssetID:  assetID,
				UniqueID: 0,
				URI:      []byte("nft-uri"),
				Metadata: []byte("NFT Metadata"),
				To:       addr,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with max supply 100
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("NFT Collection"), []byte("NFT"), 0, []byte("metadata"), []byte("uri"), 0, 100, addr, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the NFT was created correctly
				exists, _, _, nftURI, nftMetadata, owner, _ := storage.GetAssetNFT(ctx, store, nftID)
				require.True(t, exists)
				require.Equal(t, "nft-uri", string(nftURI))
				require.Equal(t, "NFT Metadata", string(nftMetadata))
				require.Equal(t, addr.String(), owner.String())

				// Check if the balance was updated
				balance, err := storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
			},
			ExpectedOutputs: &MintAssetNFTResult{
				NftID:            nftID,
				To:               addr,
				OldBalance:       0,
				NewBalance:       1,
				AssetTotalSupply: 1,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkMintAssetNFT(b *testing.B) {
	require := require.New(b)
	actor := codec.CreateAddress(0, ids.GenerateTestID())
	assetID := ids.GenerateTestID()

	mintAssetNFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "MintAssetNFTBenchmark",
		Actor: actor,
		Action: &MintAssetNFT{
			AssetID:  assetID,
			UniqueID: 0,
			URI:      []byte("nft-uri"),
			Metadata: []byte("NFT Metadata"),
			To:       actor,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAsset(context.Background(), store, assetID, nconsts.AssetNonFungibleTokenID, []byte("Benchmark NFT Collection"), []byte("BNFT"), 0, []byte("benchmark metadata"), []byte("benchmark-uri"), 0, 100, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the NFT was created correctly
			nftID := utils.GenerateIDWithIndex(assetID, 0)
			exists, _, _, nftURI, nftMetadata, owner, _ := storage.GetAssetNFT(ctx, store, nftID)
			require.True(exists)
			require.Equal(b, "nft-uri", string(nftURI))
			require.Equal(b, "NFT Metadata", string(nftMetadata))
			require.Equal(b, actor.String(), owner.String())

			// Check if balance updated
			balance, err := storage.GetBalance(ctx, store, actor, assetID)
			require.NoError(err)
			require.Equal(b, uint64(1), balance)
		},
	}

	ctx := context.Background()
	mintAssetNFTActionBenchmark.Run(ctx, b)
}
