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
)

func TestUpdateDatasetAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "DatasetNotFound",
			Actor: addr,
			Action: &UpdateDataset{
				DatasetAddress: datasetID, // Dataset does not exist
				Name:      []byte("Updated Name"),
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrDatasetNotFound,
		},
		{
			Name:  "NoFieldsUpdated",
			Actor: addr,
			Action: &UpdateDataset{
				DatasetAddress: datasetID, // No fields changed
				Name:      []byte("Dataset Name"),
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 0, addr))
				return store
			}(),
			ExpectedErr: ErrOutputMustUpdateAtLeastOneField,
		},
		{
			Name:  "InvalidNameUpdate",
			Actor: addr,
			Action: &UpdateDataset{
				DatasetAddress: datasetID,
				Name:      []byte("Na"), // Invalid name (too short)
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 0, addr))
				return store
			}(),
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "ValidUpdateDataset",
			Actor: addr,
			Action: &UpdateDataset{
				DatasetAddress:   datasetID,
				Name:        []byte("Updated Name"),
				Description: []byte("Updated Description"),
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 0, addr))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the dataset was updated correctly
				exists, name, description, _, _, _, _, _, _, _, _, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, store, datasetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, "Updated Name", string(name))
				require.Equal(t, "Updated Description", string(description))
				require.Equal(t, addr, owner)
			},
			ExpectedOutputs: &UpdateDatasetResult{
				Name:        []byte("Updated Name"),
				Description: []byte("Updated Description"),
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkUpdateDataset(b *testing.B) {
	require := require.New(b)
	actor := codec.CreateAddress(0, ids.GenerateTestID())
	datasetID := ids.GenerateTestID()

	updateDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "UpdateDatasetBenchmark",
		Actor: actor,
		Action: &UpdateDataset{
			DatasetAddress:   datasetID,
			Name:        []byte("Benchmark Dataset Name"),
			Description: []byte("Benchmark Description"),
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 0, actor))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the dataset was updated correctly
			exists, name, description, _, _, _, _, _, _, _, _, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, store, datasetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, "Benchmark Dataset Name", string(name))
			require.Equal(b, "Benchmark Description", string(description))
			require.Equal(b, actor, owner)
		},
	}

	ctx := context.Background()
	updateDatasetBenchmark.Run(ctx, b)
}
