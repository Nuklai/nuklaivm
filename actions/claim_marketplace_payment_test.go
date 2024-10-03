// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
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

func TestClaimMarketplacePaymentAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	marketplaceAssetID := ids.GenerateTestID()
	baseAssetID := ids.Empty

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 100})

	tests := []chaintest.ActionTest{
		{
			Name:  "DatasetNotFound",
			Actor: addr,
			Action: &ClaimMarketplacePayment{
				DatasetID:          datasetID, // Non-existent dataset ID
				MarketplaceAssetID: marketplaceAssetID,
				AssetForPayment:    baseAssetID,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrDatasetNotFound,
		},
		{
			Name:  "WrongOwner",
			Actor: codectest.NewRandomAddress(), // Not the owner of the dataset
			Action: &ClaimMarketplacePayment{
				DatasetID:          datasetID,
				MarketplaceAssetID: marketplaceAssetID,
				AssetForPayment:    baseAssetID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with a different owner
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrOutputWrongOwner,
		},
		{
			Name:  "AssetNotSupported",
			Actor: addr,
			Action: &ClaimMarketplacePayment{
				DatasetID:          datasetID,
				MarketplaceAssetID: marketplaceAssetID,
				AssetForPayment:    ids.GenerateTestID(), // Non-existent base asset ID
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with the correct owner
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), false, ids.Empty, ids.Empty, 0, 100, 0, 100, 100, addr))
				return store
			}(),
			ExpectedErr: ErrBaseAssetNotSupported,
		},
		{
			Name:  "NoPaymentRemaining",
			Actor: addr,
			Action: &ClaimMarketplacePayment{
				DatasetID:          datasetID,
				MarketplaceAssetID: marketplaceAssetID,
				AssetForPayment:    baseAssetID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with the correct owner and metadata
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, addr))
				// Set the marketplace asset with no payment remaining
				metadataMap := map[string]string{
					"paymentRemaining": "0",
					"paymentClaimed":   "0",
					"lastClaimedBlock": "50",
				}
				metadata, err := utils.MapToBytes(metadataMap)
				require.NoError(t, err)
				require.NoError(t, storage.SetAsset(context.Background(), store, marketplaceAssetID, nconsts.AssetMarketplaceTokenID, []byte("Marketplace Token"), []byte("MKT"), 0, metadata, []byte(datasetID.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrNoPaymentRemaining,
		},
		{
			Name:     "ValidPaymentClaim",
			ActionID: marketplaceAssetID,
			Actor:    addr,
			Action: &ClaimMarketplacePayment{
				DatasetID:          datasetID,
				MarketplaceAssetID: marketplaceAssetID,
				AssetForPayment:    baseAssetID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with the correct owner
				require.NoError(t, storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, addr))
				// Set the marketplace asset with payment remaining
				metadataMap := map[string]string{
					"paymentRemaining": "100", // 100 units remaining
					"paymentClaimed":   "0",   // No payments claimed yet
					"lastClaimedBlock": "0",   // Last claimed at block 0
				}
				metadata, err := utils.MapToBytes(metadataMap)
				require.NoError(t, err)
				require.NoError(t, storage.SetAsset(context.Background(), store, marketplaceAssetID, nconsts.AssetMarketplaceTokenID, []byte("Marketplace Token"), []byte("MKT"), 0, metadata, []byte(datasetID.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the payment was correctly claimed
				balance, err := storage.GetBalance(ctx, store, addr, baseAssetID)
				require.NoError(t, err)
				require.Equal(t, uint64(100), balance) // 100 units claimed

				// Check if metadata was updated correctly
				exists, _, _, _, _, metadata, _, _, _, _, _, _, _, _, err := storage.GetAsset(ctx, store, marketplaceAssetID)
				require.NoError(t, err)
				require.True(t, exists)

				metadataMap, err := utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, "0", metadataMap["paymentRemaining"]) // 1000 - 100 claimed
				require.Equal(t, "100", metadataMap["paymentClaimed"])
				require.Equal(t, "100", metadataMap["lastClaimedBlock"])
			},
			ExpectedOutputs: &ClaimMarketplacePaymentResult{
				LastClaimedBlock:  100,
				PaymentClaimed:    100,
				PaymentRemaining:  0,
				DistributedReward: 100,
				DistributedTo:     addr,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkClaimMarketplacePayment(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()
	marketplaceAssetID := ids.GenerateTestID()
	baseAssetID := ids.Empty

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 100})

	claimMarketplacePaymentBenchmark := &chaintest.ActionBenchmark{
		Name:  "ClaimMarketplacePaymentBenchmark",
		Actor: actor,
		Action: &ClaimMarketplacePayment{
			DatasetID:          datasetID,
			MarketplaceAssetID: marketplaceAssetID,
			AssetForPayment:    baseAssetID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set the dataset with the correct owner
			require.NoError(storage.SetDataset(context.Background(), store, datasetID, []byte("Dataset Name"), []byte("Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, marketplaceAssetID, baseAssetID, 100, 100, 0, 100, 100, actor))
			// Set the marketplace asset with payment remaining
			metadataMap := map[string]string{
				"paymentRemaining": "1000",
				"paymentClaimed":   "0",
				"lastClaimedBlock": "50",
			}
			metadata, err := utils.MapToBytes(metadataMap)
			require.NoError(err)
			require.NoError(storage.SetAsset(context.Background(), store, marketplaceAssetID, nconsts.AssetMarketplaceTokenID, []byte("Marketplace Token"), []byte("MKT"), 0, metadata, []byte(datasetID.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly updated
			balance, err := storage.GetBalance(ctx, store, actor, baseAssetID)
			require.NoError(err)
			require.Equal(b, uint64(100), balance) // 100 units claimed
		},
	}

	ctx := context.Background()
	claimMarketplacePaymentBenchmark.Run(ctx, b)
}
