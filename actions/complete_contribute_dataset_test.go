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

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestCompleteContributeDatasetAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	uniqueNFTID := uint64(1)
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
			Action: &CompleteContributeDataset{
				DatasetID:                 datasetID, // Non-existent dataset ID
				Contributor:               addr,
				UniqueNFTIDForContributor: uniqueNFTID,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrDatasetNotFound,
		},
		{
			Name:  "WrongOwner",
			Actor: codectest.NewRandomAddress(), // Not the owner of the dataset
			Action: &CompleteContributeDataset{
				DatasetID:                 datasetID,
				Contributor:               addr,
				UniqueNFTIDForContributor: uniqueNFTID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset with a different owner
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrOutputWrongOwner,
		},
		{
			Name:  "DatasetAlreadyOnSale",
			Actor: addr,
			Action: &CompleteContributeDataset{
				DatasetID:                 datasetID,
				Contributor:               addr,
				UniqueNFTIDForContributor: uniqueNFTID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset that is already on sale
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.GenerateTestID(), ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrDatasetAlreadyOnSale,
		},
		{
			Name:  "NFTAlreadyExists",
			Actor: addr,
			Action: &CompleteContributeDataset{
				DatasetID:                 datasetID,
				Contributor:               addr,
				UniqueNFTIDForContributor: uniqueNFTID, // NFT already exists
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				// Create existing NFT
				nftID := nchain.GenerateIDWithIndex(datasetID, uniqueNFTID)
				require.NoError(t, storage.SetAssetNFT(context.Background(), store, datasetID, uniqueNFTID, nftID, []byte("Dataset NFT"), []byte("Metadata"), addr))
				return store
			}(),
			ExpectedErr: ErrOutputNFTAlreadyExists,
		},
		{
			Name:     "ValidCompletion",
			ActionID: ids.GenerateTestID(),
			Actor:    addr,
			Action: &CompleteContributeDataset{
				DatasetID:                 datasetID,
				Contributor:               addr,
				UniqueNFTIDForContributor: uniqueNFTID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid dataset
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				require.NoError(t, storage.SetAsset(context.Background(), store, datasetID, nconsts.AssetDatasetTokenID, []byte("Base Token"), []byte("BASE"), 0, []byte("Metadata"), []byte("uri"), 0, 0, addr, addr, addr, addr, addr))
				// Set balance to 0
				config := marketplace.GetDatasetConfig()
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, config.CollateralAssetIDForDataContribution, 0))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				config := marketplace.GetDatasetConfig()
				nftID := nchain.GenerateIDWithIndex(datasetID, uniqueNFTID)

				// Check if the balance is correctly updated
				balance, err := storage.GetBalance(ctx, store, addr, config.CollateralAssetIDForDataContribution)
				require.NoError(t, err)
				require.Equal(t, config.CollateralAmountForDataContribution, balance) // Collateral refunded

				// Check if the NFT was created correctly
				nftExists, _, _, _, _, owner, _ := storage.GetAssetNFT(ctx, store, nftID)
				require.True(t, nftExists)
				require.Equal(t, addr, owner)

				// Verify marketplace contribution
				contribution, err := mockMarketplace.CompleteContributeDataset(datasetID, addr)
				require.NoError(t, err)
				require.Equal(t, string(dataLocation), string(contribution.DataLocation))
				require.Equal(t, string(dataIdentifier), string(contribution.DataIdentifier))
			},
			ExpectedOutputs: &CompleteContributeDatasetResult{
				CollateralAssetID:        marketplace.GetDatasetConfig().CollateralAssetIDForDataContribution,
				CollateralAmountRefunded: marketplace.GetDatasetConfig().CollateralAmountForDataContribution,
				DatasetID:                datasetID,
				DatasetChildNftID:        nchain.GenerateIDWithIndex(datasetID, uniqueNFTID),
				To:                       addr,
				DataLocation:             dataLocation,
				DataIdentifier:           dataIdentifier,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkCompleteContributeDataset(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	uniqueNFTID := uint64(1)
	dataLocation := []byte("default")
	dataIdentifier := []byte("data_id_1234")

	mockMarketplace := marketplace.MockNewMarketplace(&marketplace.MockMarketplace{
		DataContribution: marketplace.DataContribution{
			DataLocation:   dataLocation,
			DataIdentifier: dataIdentifier,
			Contributor:    actor,
		},
	})

	completeContributeDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "CompleteContributeDatasetBenchmark",
		Actor: actor,
		Action: &CompleteContributeDataset{
			DatasetID:                 datasetID,
			Contributor:               actor,
			UniqueNFTIDForContributor: uniqueNFTID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set valid dataset
			require.NoError(storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, actor))
			require.NoError(storage.SetAsset(context.Background(), store, datasetID, nconsts.AssetDatasetTokenID, []byte("Base Token"), []byte("BASE"), 0, []byte("Metadata"), []byte("uri"), 0, 0, actor, actor, actor, actor, actor))
			// Set balance to 0
			config := marketplace.GetDatasetConfig()
			require.NoError(storage.SetBalance(context.Background(), store, actor, config.CollateralAssetIDForDataContribution, 0))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			config := marketplace.GetDatasetConfig()

			// Check if the balance is correctly updated
			balance, err := storage.GetBalance(ctx, store, actor, config.CollateralAssetIDForDataContribution)
			require.NoError(err)
			require.Equal(b, config.CollateralAmountForDataContribution, balance) // Collateral refunded

			// Verify marketplace contribution
			contribution, err := mockMarketplace.CompleteContributeDataset(datasetID, actor)
			require.NoError(err)
			require.Equal(string(dataLocation), string(contribution.DataLocation))
			require.Equal(string(dataIdentifier), string(contribution.DataIdentifier))
		},
	}

	ctx := context.Background()
	completeContributeDatasetBenchmark.Run(ctx, b)
}
