// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/marketplace"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"
)

func TestInitiateContributeDatasetAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	dataLocation := []byte("default")
	dataIdentifier := []byte("data_id_1234")

	mockMarketplace := marketplace.MockNewMarketplace(&marketplace.MockMarketplace{
		DataContribution: marketplace.DataContribution{
			DataLocation:   dataLocation,
			DataIdentifier: dataIdentifier,
			Contributor:    addr,
		},
	})

	tests := []chaintest.ActionTest{
		{
			Name:  "DatasetNotFound",
			Actor: addr,
			Action: &InitiateContributeDataset{
				DatasetID:      datasetID, // Non-existent dataset ID
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrDatasetNotFound,
		},
		{
			Name:  "DatasetNotOpenForContribution",
			Actor: addr,
			Action: &InitiateContributeDataset{
				DatasetID:      datasetID,
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset that is not open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrDatasetNotOpenForContribution,
		},
		{
			Name:  "DatasetAlreadyOnSale",
			Actor: addr,
			Action: &InitiateContributeDataset{
				DatasetID:      datasetID,
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset that is already on sale
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.GenerateTestID(), ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrDatasetAlreadyOnSale,
		},
		{
			Name:  "InvalidDataLocation",
			Actor: addr,
			Action: &InitiateContributeDataset{
				DatasetID:      datasetID,
				DataLocation:   []byte("d"), // Invalid data location (too short)
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrOutputDataLocationInvalid,
		},
		{
			Name:  "InvalidDataIdentifier",
			Actor: addr,
			Action: &InitiateContributeDataset{
				DatasetID:      datasetID,
				DataLocation:   dataLocation,
				DataIdentifier: []byte(""), // Invalid data identifier (empty)
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrURIInvalid,
		},
		{
			Name:     "ValidContribution",
			ActionID: ids.GenerateTestID(),
			Actor:    addr,
			Action: &InitiateContributeDataset{
				DatasetID:      datasetID,
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				// Set sufficient balance for collateral
				config := marketplace.GetDatasetConfig()
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, config.CollateralAssetIDForDataContribution, config.CollateralAmountForDataContribution))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				config := marketplace.GetDatasetConfig()

				// Check if balance is correctly deducted
				balance, err := storage.GetBalance(ctx, store, addr, config.CollateralAssetIDForDataContribution)
				require.NoError(t, err)
				require.Equal(t, uint64(0), balance) // Initial collateral balance should be zero after deduction

				// Verify that the contribution is initiated correctly in the marketplace
				contributions, err := mockMarketplace.GetDataContribution(datasetID, addr)
				require.NoError(t, err)
				require.Equal(t, string(dataLocation), string(contributions[0].DataLocation))
				require.Equal(t, string(dataIdentifier), string(contributions[0].DataIdentifier))
			},
			ExpectedOutputs: &InitiateContributeDatasetResult{
				CollateralAssetID:     marketplace.GetDatasetConfig().CollateralAssetIDForDataContribution,
				CollateralAmountTaken: marketplace.GetDatasetConfig().CollateralAmountForDataContribution,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkInitiateContributeDataset(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	dataLocation := []byte("default")
	dataIdentifier := []byte("data_id_1234")

	mockMarketplace := marketplace.MockNewMarketplace(&marketplace.MockMarketplace{
		DataContribution: marketplace.DataContribution{
			DataLocation:   dataLocation,
			DataIdentifier: dataIdentifier,
			Contributor:    actor,
		},
	})

	initiateContributeDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "InitiateContributeDatasetBenchmark",
		Actor: actor,
		Action: &InitiateContributeDataset{
			DatasetID:      datasetID,
			DataLocation:   dataLocation,
			DataIdentifier: dataIdentifier,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set valid dataset open for contributions
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, actor))
			// Set sufficient balance for collateral
			config := marketplace.GetDatasetConfig()
			require.NoError(storage.SetBalance(context.Background(), store, actor, config.CollateralAssetIDForDataContribution, config.CollateralAmountForDataContribution))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			config := marketplace.GetDatasetConfig()

			// Check if balance is correctly deducted
			balance, err := storage.GetBalance(ctx, store, actor, config.CollateralAssetIDForDataContribution)
			require.NoError(err)
			require.Equal(b, uint64(0), balance) // Initial collateral balance should be zero after deduction

			// Verify that the contribution is initiated correctly in the marketplace
			contributions, err := mockMarketplace.GetDataContribution(datasetID, actor)
			require.NoError(err)
			require.Equal(string(dataLocation), string(contributions[0].DataLocation))
			require.Equal(string(dataIdentifier), string(contributions[0].DataIdentifier))
		},
	}

	ctx := context.Background()
	initiateContributeDatasetBenchmark.Run(ctx, b)
}
