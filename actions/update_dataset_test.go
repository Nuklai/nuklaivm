// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"
)

func TestUpdateDatasetAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), []byte("uri"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "NoFieldsUpdated",
			Actor: actor,
			Action: &UpdateDataset{
				DatasetAddress: datasetAddress, // No fields changed
				Name:           "Valid Name",
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrOutputMustUpdateAtLeastOneField,
		},
		{
			Name:  "InvalidNameUpdate",
			Actor: actor,
			Action: &UpdateDataset{
				DatasetAddress: datasetAddress,
				Name:           "na", // Invalid name (too short)
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "ValidUpdateDataset",
			Actor: actor,
			Action: &UpdateDataset{
				DatasetAddress: datasetAddress,
				Name:           "Updated Name",
				Description:    "Updated Description",
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the dataset was updated correctly
				name, description, _, _, _, _, _, _, _, _, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, store, datasetAddress)
				require.NoError(t, err)
				require.Equal(t, "Updated Name", string(name))
				require.Equal(t, "Updated Description", string(description))
				require.Equal(t, actor, owner)
			},
			ExpectedOutputs: &UpdateDatasetResult{
				Name:        "Updated Name",
				Description: "Updated Description",
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkUpdateDataset(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), []byte("uri"), actor)

	updateDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "UpdateDatasetBenchmark",
		Actor: actor,
		Action: &UpdateDataset{
			DatasetAddress: datasetAddress,
			Name:           "Updated Name",
			Description:    "Updated Description",
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the dataset was updated correctly
			name, description, _, _, _, _, _, _, _, _, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, store, datasetAddress)
			require.NoError(err)
			require.Equal(b, "Benchmark Dataset Name", string(name))
			require.Equal(b, "Benchmark Description", string(description))
			require.Equal(b, actor, owner)
		},
	}

	ctx := context.Background()
	updateDatasetBenchmark.Run(ctx, b)
}
