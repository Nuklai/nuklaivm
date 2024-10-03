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
	addr := codectest.NewRandomAddress()
	assetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "WrongOwner",
			Actor: codec.CreateAddress(0, ids.GenerateTestID()), // Not the actual owner
			Action: &UpdateAsset{
				AssetID: assetID,
				Name:    []byte("New Name"),
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 0, 1000, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputWrongOwner,
		},
		{
			Name:  "NoFieldUpdated",
			Actor: addr,
			Action: &UpdateAsset{
				AssetID: assetID,
				Name:    []byte("My Token"), // Same as current name
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 0, 1000, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrOutputMustUpdateAtLeastOneField,
		},
		{
			Name:  "InvalidName",
			Actor: addr,
			Action: &UpdateAsset{
				AssetID: assetID,
				Name:    []byte("Up"), // Invalid name (too short)
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 0, 1000, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "UpdateNameAndSymbol",
			Actor: addr,
			Action: &UpdateAsset{
				AssetID: assetID,
				Name:    []byte("Updated Name"),
				Symbol:  []byte("UPD"),
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set asset with `addr` as the owner
				require.NoError(t, storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 0, 1000, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the asset was updated correctly
				exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, store, assetID)
				require.True(t, exists)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetFungibleTokenID, assetType)
				require.Equal(t, "Updated Name", string(name))
				require.Equal(t, "UPD", string(symbol))
				require.Equal(t, uint8(9), decimals)
				require.Equal(t, "Metadata", string(metadata))
				require.Equal(t, "uri", string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(1000), maxSupply)
				require.Equal(t, addr, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
			},
			ExpectedOutputs: &UpdateAssetResult{
				Name:      []byte("Updated Name"),
				Symbol:    []byte("UPD"),
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
	assetID := ids.GenerateTestID()

	updateAssetActionBenchmark := &chaintest.ActionBenchmark{
		Name:  "UpdateAssetBenchmark",
		Actor: actor,
		Action: &UpdateAsset{
			AssetID: assetID,
			Name:    []byte("Benchmark Updated Asset"),
			Symbol:  []byte("BUP"),
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAsset(context.Background(), store, assetID, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 0, 1000, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the asset was updated correctly
			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAsset(ctx, store, assetID)
			require.True(exists)
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
