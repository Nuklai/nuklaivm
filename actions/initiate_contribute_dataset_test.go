// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/nuklai/nuklaivm/dataset"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestInitiateContributeDatasetAction(t *testing.T) {
	dataLocation := "default"
	dataIdentifier := "data_id_1234"

	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	datasetContributionID := storage.DatasetContributionID(datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "DatasetNotOpenForContribution",
			Actor: actor,
			Action: &InitiateContributeDataset{
				DatasetAddress: datasetAddress,
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset that is not open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrDatasetNotOpenForContribution,
		},
		{
			Name:  "DatasetAlreadyOnSale",
			Actor: actor,
			Action: &InitiateContributeDataset{
				DatasetAddress: datasetAddress,
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset that is already on sale
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codectest.NewRandomAddress(), codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrDatasetAlreadyOnSale,
		},
		{
			Name:  "InvalidDataLocation",
			Actor: actor,
			Action: &InitiateContributeDataset{
				DatasetAddress: datasetAddress,
				DataLocation:   strings.Repeat("d", 65), // Invalid data location (too long)
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrDataLocationInvalid,
		},
		{
			Name:  "InvalidDataIdentifier",
			Actor: actor,
			Action: &InitiateContributeDataset{
				DatasetAddress: datasetAddress,
				DataLocation:   dataLocation,
				DataIdentifier: "", // Invalid data identifier (empty)
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrDataIdentifierInvalid,
		},
		{
			Name:  "ValidContribution",
			Actor: actor,
			Action: &InitiateContributeDataset{
				DatasetAddress: datasetAddress,
				DataLocation:   dataLocation,
				DataIdentifier: dataIdentifier,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset open for contributions
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				// Set sufficient balance for collateral
				config := dataset.GetDatasetConfig()
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, config.CollateralAssetAddressForDataContribution, actor, config.CollateralAmountForDataContribution))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				config := dataset.GetDatasetConfig()

				// Check if balance is correctly deducted
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, config.CollateralAssetAddressForDataContribution, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(0), balance) // Initial collateral balance should be zero after deduction

				// Verify that the contribution is initiated correctly
				datasetAddress, dataLocation, dataIdentifier, contributor, active, err := storage.GetDatasetContributionInfoNoController(ctx, store, datasetContributionID)
				require.NoError(t, err)
				require.Equal(t, datasetAddress, datasetAddress)
				require.Equal(t, "default", string(dataLocation))
				require.Equal(t, "data_id_1234", string(dataIdentifier))
				require.Equal(t, actor, contributor)
				require.False(t, active)
			},
			ExpectedOutputs: &InitiateContributeDatasetResult{
				Actor:                  actor.String(),
				Receiver:               "",
				DatasetContributionID:  datasetContributionID.String(),
				CollateralAssetAddress: dataset.GetDatasetConfig().CollateralAssetAddressForDataContribution.String(),
				CollateralAmountTaken:  dataset.GetDatasetConfig().CollateralAmountForDataContribution,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkInitiateContributeDataset(b *testing.B) {
	require := require.New(b)
	dataLocation := "default"
	dataIdentifier := "data_id_1234"

	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	datasetContributionID := storage.DatasetContributionID(datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor)

	initiateContributeDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "InitiateContributeDatasetBenchmark",
		Actor: actor,
		Action: &InitiateContributeDataset{
			DatasetAddress: datasetAddress,
			DataLocation:   dataLocation,
			DataIdentifier: dataIdentifier,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set valid dataset open for contributions
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
			// Set sufficient balance for collateral
			config := dataset.GetDatasetConfig()
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, config.CollateralAssetAddressForDataContribution, actor, config.CollateralAmountForDataContribution))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			config := dataset.GetDatasetConfig()

			// Check if balance is correctly deducted
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, config.CollateralAssetAddressForDataContribution, actor)
			require.NoError(err)
			require.Equal(uint64(0), balance) // Initial collateral balance should be zero after deduction

			// Verify that the contribution is initiated correctly
			datasetAddress, dataLocation, dataIdentifier, contributor, active, err := storage.GetDatasetContributionInfoNoController(ctx, store, datasetContributionID)
			require.NoError(err)
			require.Equal(datasetAddress, datasetAddress)
			require.Equal("default", string(dataLocation))
			require.Equal("data_id_1234", string(dataIdentifier))
			require.Equal(actor, contributor)
			require.False(active)
		},
	}

	ctx := context.Background()
	initiateContributeDatasetBenchmark.Run(ctx, b)
}
