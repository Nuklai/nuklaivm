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

func TestUpdateAssetAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "WrongOwner",
			Actor: codec.CreateAddress(0, ids.GenerateTestID()), // Not the actual owner
			Action: &UpdateAsset{
				AssetAddress: assetAddress,
				Name:         "New Name",
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrWrongOwner,
		},
		{
			Name:  "NoFieldUpdated",
			Actor: actor,
			Action: &UpdateAsset{
				AssetAddress: assetAddress,
				Name:         "My Token", // Same as current name
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputMustUpdateAtLeastOneField,
		},
		{
			Name:  "InvalidName",
			Actor: actor,
			Action: &UpdateAsset{
				AssetAddress: assetAddress,
				Name:         "Up", // Invalid name (too short)
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "UpdateNameAndSymbol",
			Actor: actor,
			Action: &UpdateAsset{
				AssetAddress: assetAddress,
				Name:         "Updated Name",
				Symbol:       "UPD",
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was updated correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetFungibleTokenID, assetType)
				require.Equal(t, "Updated Name", string(name))
				require.Equal(t, "UPD", string(symbol))
				require.Equal(t, uint8(9), decimals)
				require.Equal(t, "Metadata", string(metadata))
				require.Equal(t, "uri", string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(1000), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
			},
			ExpectedOutputs: &UpdateAssetResult{
				Name:      "Updated Name",
				Symbol:    "UPD",
				MaxSupply: 1000,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkUpdateAsset(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	assetAddress := storage.AssetAddress(nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), actor)

	updateAssetActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "UpdateAssetBenchmark",
		Actor: actor,
		Action: &UpdateAsset{
			AssetAddress: assetAddress,
			Name:         "Benchmark Updated Asset",
			Symbol:       "BUP",
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAssetInfo(context.Background(), store, assetAddress, nconsts.AssetFungibleTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the asset was updated correctly
			assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, assetAddress)
			require.NoError(err)
			require.Equal(nconsts.AssetFungibleTokenID, assetType)
			require.Equal("Benchmark Updated Asset", string(name))
			require.Equal("BUP", string(symbol))
			require.Equal(uint8(9), decimals)
			require.Equal("Metadata", string(metadata))
			require.Equal("uri", string(uri))
			require.Equal(uint64(0), totalSupply)
			require.Equal(uint64(1000), maxSupply)
			require.Equal(actor, owner)
			require.Equal(codec.EmptyAddress, mintAdmin)
			require.Equal(codec.EmptyAddress, pauseUnpauseAdmin)
			require.Equal(codec.EmptyAddress, freezeUnfreezeAdmin)
			require.Equal(codec.EmptyAddress, enableDisableKYCAccountAdmin)
		},
	}

	ctx := context.Background()
	updateAssetActionBenchmark.Run(ctx, b)
}
