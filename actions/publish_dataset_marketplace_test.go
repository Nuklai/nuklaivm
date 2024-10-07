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

func TestPublishDatasetMarketplaceAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	baseAssetID := ids.Empty
	marketplaceAssetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "DatasetNotFound",
			Actor: addr,
			Action: &PublishDatasetMarketplace{
				MarketplaceAssetID:   marketplaceAssetID,
				DatasetAddress:       datasetID, // Non-existent dataset ID
				PaymentAssetAddress:  baseAssetID,
				DatasetPricePerBlock: 100,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrDatasetNotFound,
		},
		{
			Name:  "WrongOwner",
			Actor: codectest.NewRandomAddress(), // Not the owner of the dataset
			Action: &PublishDatasetMarketplace{
				MarketplaceAssetID:   marketplaceAssetID,
				DatasetAddress:       datasetID,
				PaymentAssetAddress:  baseAssetID,
				DatasetPricePerBlock: 100,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with a different owner
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrWrongOwner,
		},
		{
			Name:  "AssetNotFound",
			Actor: addr,
			Action: &PublishDatasetMarketplace{
				MarketplaceAssetID:   marketplaceAssetID,
				DatasetAddress:       datasetID,
				PaymentAssetAddress:  baseAssetID, // Non-existent base asset ID
				DatasetPricePerBlock: 100,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with the correct owner
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrAssetNotFound,
		},
		{
			Name:     "ValidPublishDataset",
			ActionID: marketplaceAssetID,
			Actor:    addr,
			Action: &PublishDatasetMarketplace{
				MarketplaceAssetID:   marketplaceAssetID,
				DatasetAddress:       datasetID,
				PaymentAssetAddress:  baseAssetID,
				DatasetPricePerBlock: 100,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the base asset with the required details
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				// Set the dataset with the correct owner
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, datasetID, nconsts.AssetFractionalTokenID, []byte("Base Asset"), []byte("BA"), 0, []byte("Metadata"), []byte("uri"), 1, 0, addr, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the dataset was updated correctly
				exists, _, _, _, _, _, _, _, _, saleID, baseAsset, basePrice, _, _, _, _, _, err := storage.GetDatasetInfoNoController(ctx, store, datasetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, marketplaceAssetID, saleID)
				require.Equal(t, baseAssetID, baseAsset)
				require.Equal(t, uint64(100), basePrice)

				// Check if the marketplace asset was created correctly
				exists, assetType, name, symbol, _, metadata, _, _, _, owner, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, marketplaceAssetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, nconsts.AssetMarketplaceTokenID, assetType)
				require.Contains(t, string(name), "Dataset-Marketplace-")
				require.Contains(t, string(symbol), "DM-")
				require.Equal(t, codec.EmptyAddress, owner)

				// Check metadata
				metadataMap, err := utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, datasetID.String(), metadataMap["datasetID"])
				require.Equal(t, marketplaceAssetID.String(), metadataMap["marketplaceAssetID"])
				require.Equal(t, "100", metadataMap["datasetPricePerBlock"])
				require.Equal(t, baseAssetID.String(), metadataMap["assetForPayment"])
				require.Equal(t, addr.String(), metadataMap["publisher"])
			},
			ExpectedOutputs: &PublishDatasetMarketplaceResult{
				MarketplaceAssetAddress: marketplaceAssetID,
				PaymentAssetAddress:     baseAssetID,
				DatasetPricePerBlock:    100,
				Publisher:               addr,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkPublishDatasetMarketplace(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	baseAssetID := ids.GenerateTestID()
	marketplaceAssetID := ids.GenerateTestID()

	publishDatasetMarketplaceBenchmark := &chaintest.ActionBenchmark{
		Name:  "PublishDatasetMarketplaceBenchmark",
		Actor: actor,
		Action: &PublishDatasetMarketplace{
			MarketplaceAssetID:   marketplaceAssetID,
			DatasetAddress:       datasetID,
			PaymentAssetAddress:  baseAssetID,
			DatasetPricePerBlock: 100,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set the base asset with the required details
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, actor))
			// Set the dataset with the correct owner
			require.NoError(storage.SetAssetInfo(context.Background(), store, datasetID, nconsts.AssetFractionalTokenID, []byte("Base Asset"), []byte("BA"), 0, []byte("Metadata"), []byte("uri"), 1, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the dataset was updated correctly
			exists, _, _, _, _, _, _, _, _, saleID, baseAsset, basePrice, _, _, _, _, _, err := storage.GetDatasetInfoNoController(ctx, store, datasetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, marketplaceAssetID, saleID)
			require.Equal(b, baseAssetID, baseAsset)
			require.Equal(b, uint64(100), basePrice)

			// Check if the marketplace asset was created correctly
			exists, assetType, name, symbol, _, metadata, _, _, _, owner, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, marketplaceAssetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, nconsts.AssetMarketplaceTokenID, assetType)
			require.Contains(b, string(name), "Dataset-Marketplace-")
			require.Contains(b, string(symbol), "DM-")
			require.Equal(b, codec.EmptyAddress, owner)

			// Check metadata
			metadataMap, err := utils.BytesToMap(metadata)
			require.NoError(err)
			require.Equal(b, datasetID.String(), metadataMap["datasetID"])
			require.Equal(b, marketplaceAssetID.String(), metadataMap["marketplaceAssetID"])
			require.Equal(b, "100", metadataMap["datasetPricePerBlock"])
			require.Equal(b, baseAssetID.String(), metadataMap["assetForPayment"])
			require.Equal(b, actor.String(), metadataMap["publisher"])
		},
	}

	ctx := context.Background()
	publishDatasetMarketplaceBenchmark.Run(ctx, b)
}
