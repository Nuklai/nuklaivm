// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"fmt"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestSubscribeDatasetMarketplaceAction(t *testing.T) {
	mockEmission := emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 1})

	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	marketplaceAssetID := ids.GenerateTestID()
	baseAssetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "DatasetNotFound",
			Actor: addr,
			Action: &SubscribeDatasetMarketplace{
				DatasetID:            datasetID, // Non-existent dataset ID
				MarketplaceAssetID:   marketplaceAssetID,
				AssetForPayment:      baseAssetID,
				NumBlocksToSubscribe: 10,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrDatasetNotFound,
		},
		{
			Name:  "DatasetNotOnSale",
			Actor: addr,
			Action: &SubscribeDatasetMarketplace{
				DatasetID:            datasetID,
				MarketplaceAssetID:   marketplaceAssetID,
				AssetForPayment:      baseAssetID,
				NumBlocksToSubscribe: 10,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset without a sale ID
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrDatasetNotOnSale,
		},
		{
			Name:  "InvalidMarketplaceAssetID",
			Actor: addr,
			Action: &SubscribeDatasetMarketplace{
				DatasetID:            datasetID,
				MarketplaceAssetID:   ids.GenerateTestID(), // Incorrect marketplace asset ID
				AssetForPayment:      baseAssetID,
				NumBlocksToSubscribe: 10,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset with a valid sale ID
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrMarketplaceAssetIDInvalid,
		},
		{
			Name:  "BaseAssetNotSupported",
			Actor: addr,
			Action: &SubscribeDatasetMarketplace{
				DatasetID:            datasetID,
				MarketplaceAssetID:   marketplaceAssetID,
				AssetForPayment:      ids.GenerateTestID(), // Invalid base asset ID
				NumBlocksToSubscribe: 10,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset with a valid sale ID and base asset
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrBaseAssetNotSupported,
		},
		{
			Name:     "ValidSubscription",
			ActionID: marketplaceAssetID,
			Actor:    addr,
			Action: &SubscribeDatasetMarketplace{
				DatasetID:            datasetID,
				MarketplaceAssetID:   marketplaceAssetID,
				AssetForPayment:      baseAssetID,
				NumBlocksToSubscribe: 10,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset with a valid sale ID and base asset
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, addr))
				require.NoError(t, storage.SetAsset(context.Background(), store, datasetID, nconsts.AssetDatasetTokenID, []byte("Base Token"), []byte("BASE"), 0, []byte("Metadata"), []byte("uri"), 0, 0, addr, addr, addr, addr, addr))
				// Set base asset balance to sufficient amount
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, baseAssetID, 5000))
				// Set the marketplace asset for the dataset
				metadataMap := make(map[string]string, 0)
				metadataMap["datasetID"] = datasetID.String()
				metadataMap["marketplaceAssetID"] = marketplaceAssetID.String()
				metadataMap["datasetPricePerBlock"] = "100"
				metadataMap["assetForPayment"] = baseAssetID.String()
				metadataMap["publisher"] = addr.String()
				metadataMap["lastClaimedBlock"] = "0"
				metadataMap["subscriptions"] = "0"
				metadataMap["paymentRemaining"] = "0"
				metadataMap["paymentClaimed"] = "0"
				metadata, err := utils.MapToBytes(metadataMap)
				require.NoError(t, err)
				require.NoError(t, storage.SetAsset(context.Background(), store, marketplaceAssetID, nconsts.AssetMarketplaceTokenID, []byte("Marketplace Token"), []byte("MKT"), 0, metadata, []byte(datasetID.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly deducted
				balance, err := storage.GetBalance(ctx, store, addr, baseAssetID)
				require.NoError(t, err)
				require.Equal(t, uint64(4000), balance) // 5000 - 1000 = 4000

				// Check if the subscription NFT was created correctly
				nftID := utils.GenerateIDWithAddress(marketplaceAssetID, addr)
				nftExists, _, _, _, _, owner, _ := storage.GetAssetNFT(ctx, store, nftID)
				require.True(t, nftExists)
				require.Equal(t, addr, owner)

				// Check if the marketplace asset metadata was updated correctly
				exists, _, _, _, _, metadata, _, _, _, _, _, _, _, _, err := storage.GetAsset(ctx, store, marketplaceAssetID)
				require.NoError(t, err)
				require.True(t, exists)

				metadataMap, err := utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, "1000", metadataMap["paymentRemaining"])
				require.Equal(t, "1", metadataMap["subscriptions"])
				require.Equal(t, fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()), metadataMap["lastClaimedBlock"])
				require.Equal(t, addr.String(), metadataMap["publisher"])
			},
			ExpectedOutputs: &SubscribeDatasetMarketplaceResult{
				MarketplaceAssetID:               marketplaceAssetID,
				MarketplaceAssetNumSubscriptions: 1,
				SubscriptionNftID:                utils.GenerateIDWithAddress(marketplaceAssetID, addr),
				AssetForPayment:                  baseAssetID,
				DatasetPricePerBlock:             100,
				TotalCost:                        1000,
				NumBlocksToSubscribe:             10,
				IssuanceBlock:                    mockEmission.GetLastAcceptedBlockHeight(),
				ExpirationBlock:                  mockEmission.GetLastAcceptedBlockHeight() + 10,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkSubscribeDatasetMarketplace(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	marketplaceAssetID := ids.GenerateTestID()
	baseAssetID := ids.GenerateTestID()

	mockEmission := emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 1,
	})

	subscribeDatasetMarketplaceBenchmark := &chaintest.ActionBenchmark{
		Name:  "SubscribeDatasetMarketplaceBenchmark",
		Actor: actor,
		Action: &SubscribeDatasetMarketplace{
			DatasetID:            datasetID,
			MarketplaceAssetID:   marketplaceAssetID,
			AssetForPayment:      baseAssetID,
			NumBlocksToSubscribe: 10,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set dataset with a valid sale ID and base asset
			require.NoError(storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, actor))
			require.NoError(storage.SetAsset(context.Background(), store, datasetID, nconsts.AssetDatasetTokenID, []byte("Base Token"), []byte("BASE"), 0, []byte("Metadata"), []byte("uri"), 0, 0, actor, actor, actor, actor, actor))
			// Set base asset balance to sufficient amount
			require.NoError(storage.SetBalance(context.Background(), store, actor, baseAssetID, 5000))
			// Set the marketplace asset for the dataset
			metadataMap := make(map[string]string, 0)
			metadataMap["datasetID"] = datasetID.String()
			metadataMap["marketplaceAssetID"] = marketplaceAssetID.String()
			metadataMap["datasetPricePerBlock"] = "100"
			metadataMap["assetForPayment"] = baseAssetID.String()
			metadataMap["publisher"] = actor.String()
			metadataMap["lastClaimedBlock"] = "0"
			metadataMap["subscriptions"] = "0"
			metadataMap["paymentRemaining"] = "0"
			metadataMap["paymentClaimed"] = "0"
			metadata, err := utils.MapToBytes(metadataMap)
			require.NoError(err)
			require.NoError(storage.SetAsset(context.Background(), store, marketplaceAssetID, nconsts.AssetMarketplaceTokenID, []byte("Marketplace Token"), []byte("MKT"), 0, metadata, []byte(datasetID.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly deducted
			balance, err := storage.GetBalance(ctx, store, actor, baseAssetID)
			require.NoError(err)
			require.Equal(b, uint64(4000), balance) // 5000 - 1000 = 4000

			// Check if the subscription NFT was created correctly
			nftID := utils.GenerateIDWithAddress(marketplaceAssetID, actor)
			nftExists, _, _, _, _, owner, _ := storage.GetAssetNFT(ctx, store, nftID)
			require.True(nftExists)
			require.Equal(b, actor, owner)

			// Check if the marketplace asset metadata was updated correctly
			exists, _, _, _, _, metadata, _, _, _, _, _, _, _, _, err := storage.GetAsset(ctx, store, marketplaceAssetID)
			require.NoError(err)
			require.True(exists)

			metadataMap, err := utils.BytesToMap(metadata)
			require.NoError(err)
			require.Equal(b, "1000", metadataMap["paymentRemaining"])
			require.Equal(b, "1", metadataMap["subscriptions"])
			require.Equal(fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()), metadataMap["lastClaimedBlock"])
			require.Equal(b, actor.String(), metadataMap["publisher"])
		},
	}

	ctx := context.Background()
	subscribeDatasetMarketplaceBenchmark.Run(ctx, b)
}
