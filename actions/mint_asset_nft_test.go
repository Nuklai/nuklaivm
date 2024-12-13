// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"strings"
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

func TestMintAssetNFTAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)
	nftAddress := storage.AssetAddressNFT(assetAddress, []byte("metadata"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidMetadata",
			Actor: actor,
			Action: &MintAssetNFT{
				AssetAddress: assetAddress,
				Metadata:     strings.Repeat("a", storage.MaxAssetMetadataSize+1),
				To:           actor,
			},
			ExpectedErr: ErrMetadataInvalid,
		},
		{
			Name:  "NFTAlreadyExists",
			Actor: actor,
			Action: &MintAssetNFT{
				AssetAddress: assetAddress,
				Metadata:     "metadata",
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set an existing NFT in storage
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, nftAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrNFTAlreadyExists,
		},
		{
			Name:  "WrongAssetType",
			Actor: actor,
			Action: &MintAssetNFT{
				AssetAddress: assetAddress,
				Metadata:     "metadata",
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset type to fungible (invalid for NFT minting)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte(assetAddress.String()), 0, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrAssetTypeInvalid,
		},
		{
			Name:  "WrongMintAdmin",
			Actor: codec.CreateAddress(0, ids.GenerateTestID()), // Not the mint admin
			Action: &MintAssetNFT{
				AssetAddress: assetAddress,
				Metadata:     "metadata",
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting asset with a different mint admin
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte(assetAddress.String()), 0, 1, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrWrongMintAdmin,
		},
		{
			Name:  "AssetAlreadyHasParent",
			Actor: actor,
			Action: &MintAssetNFT{
				AssetAddress: nftAddress, // Asset already has a parent
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Setting correct asset details
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, nftAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte(assetAddress.String()), 0, 1000, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrCantFractionalizeFurther,
		},
		{
			Name:  "ExceedMaxSupply",
			Actor: actor,
			Action: &MintAssetNFT{
				AssetAddress: assetAddress,
				Metadata:     "metadata",
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with max supply 1
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte(assetAddress.String()), 1, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: storage.ErrMaxSupplyExceeded,
		},
		{
			Name:  "ValidNFTMint",
			Actor: actor,
			Action: &MintAssetNFT{
				AssetAddress: assetAddress,
				Metadata:     "metadata",
				To:           actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with max supply 1
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte(assetAddress.String()), 0, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the NFT was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, nftAddress)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetNonFungibleTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM-0", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, "metadata", string(metadata))
				require.Equal(t, assetAddress.String(), string(uri))
				require.Equal(t, uint64(1), totalSupply)
				require.Equal(t, uint64(1), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)

				// Check if the balance was updated
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)

				// Check if the total supply was reduced
				_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err = storage.GetAssetInfoNoController(ctx, store, assetAddress)
				require.NoError(t, err)
				require.Equal(t, uint64(1), totalSupply)
			},
			ExpectedOutputs: &MintAssetNFTResult{
				Actor:           actor.String(),
				Receiver:        actor.String(),
				AssetNftAddress: nftAddress.String(),
				OldBalance:      0,
				NewBalance:      1,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkMintAssetNFT(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), actor)
	nftAddress := storage.AssetAddressNFT(assetAddress, []byte("metadata"), actor)

	mintAssetNFTActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "MintAssetNFTBenchmark",
		Actor: actor,
		Action: &MintAssetNFT{
			AssetAddress: assetAddress,
			Metadata:     "metadata",
			To:           actor,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte(assetAddress.String()), 0, 1, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the NFT was created correctly
			assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, nftAddress)
			require.NoError(err)
			require.Equal(nconsts.AssetNonFungibleTokenID, assetType)
			require.Equal("name", string(name))
			require.Equal("SYM-0", string(symbol))
			require.Equal(uint8(0), decimals)
			require.Equal("metadata", string(metadata))
			require.Equal(assetAddress.String(), string(uri))
			require.Equal(uint64(1), totalSupply)
			require.Equal(uint64(1), maxSupply)
			require.Equal(actor, owner)
			require.Equal(codec.EmptyAddress, mintAdmin)
			require.Equal(codec.EmptyAddress, pauseUnpauseAdmin)
			require.Equal(codec.EmptyAddress, freezeUnfreezeAdmin)
			require.Equal(codec.EmptyAddress, enableDisableKYCAccountAdmin)

			// Check if the total supply was reduced
			_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err = storage.GetAssetInfoNoController(ctx, store, assetAddress)
			require.NoError(err)
			require.Equal(uint64(1), totalSupply)
		},
	}

	ctx := context.Background()
	mintAssetNFTActionBenchmark.Run(ctx, b)
}
