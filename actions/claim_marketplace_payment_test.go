// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

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
	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	baseAssetAddress := storage.NAIAddress
	marketplaceAssetAddress := storage.AssetAddressFractional(datasetAddress)

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 100})

	tests := []chaintest.ActionTest{
		{
			Name:  "WrongOwner",
			Actor: codectest.NewRandomAddress(), // Not the owner of the asset
			Action: &ClaimMarketplacePayment{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     baseAssetAddress,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				metadata := map[string]string{
					"datasetAddress":          datasetAddress.String(),
					"marketplaceAssetAddress": marketplaceAssetAddress.String(),
					"datasetPricePerBlock":    "100",
					"paymentAssetAddress":     baseAssetAddress.String(),
					"publisher":               actor.String(),
					"lastClaimedBlock":        "0",
					"subscriptions":           "0",
					"paymentRemaining":        "0",
					"paymentClaimed":          "0",
				}
				metadataBytes, err := utils.MapToBytes(metadata)
				require.NoError(t, err)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrWrongOwner,
		},
		{
			Name:  "BaseAssetNotSupported",
			Actor: actor,
			Action: &ClaimMarketplacePayment{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     codec.EmptyAddress, // Invalid base asset ID
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				metadata := map[string]string{
					"datasetAddress":          datasetAddress.String(),
					"marketplaceAssetAddress": marketplaceAssetAddress.String(),
					"datasetPricePerBlock":    "100",
					"paymentAssetAddress":     baseAssetAddress.String(),
					"publisher":               actor.String(),
					"lastClaimedBlock":        "0",
					"subscriptions":           "0",
					"paymentRemaining":        "0",
					"paymentClaimed":          "0",
				}
				metadataBytes, err := utils.MapToBytes(metadata)
				require.NoError(t, err)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrPaymentAssetNotSupported,
		},
		{
			Name:  "NoPaymentRemaining",
			Actor: actor,
			Action: &ClaimMarketplacePayment{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     baseAssetAddress,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				metadata := map[string]string{
					"datasetAddress":          datasetAddress.String(),
					"marketplaceAssetAddress": marketplaceAssetAddress.String(),
					"datasetPricePerBlock":    "100",
					"paymentAssetAddress":     baseAssetAddress.String(),
					"publisher":               actor.String(),
					"lastClaimedBlock":        "50",
					"subscriptions":           "0",
					"paymentRemaining":        "0",
					"paymentClaimed":          "0",
				}
				metadataBytes, err := utils.MapToBytes(metadata)
				require.NoError(t, err)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrNoPaymentRemaining,
		},
		{
			Name:  "ValidPaymentClaim",
			Actor: actor,
			Action: &ClaimMarketplacePayment{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     baseAssetAddress,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				metadata := map[string]string{
					"datasetAddress":          datasetAddress.String(),
					"marketplaceAssetAddress": marketplaceAssetAddress.String(),
					"datasetPricePerBlock":    "100",
					"paymentAssetAddress":     baseAssetAddress.String(),
					"publisher":               actor.String(),
					"lastClaimedBlock":        "0",
					"subscriptions":           "0",
					"paymentRemaining":        "100",
					"paymentClaimed":          "0",
				}
				metadataBytes, err := utils.MapToBytes(metadata)
				require.NoError(t, err)
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the payment was correctly claimed
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, baseAssetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(100), balance) // 100 units claimed

				// Check if metadata was updated correctly
				_, _, _, _, metadata, _, _, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, marketplaceAssetAddress)
				require.NoError(t, err)
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
				DistributedTo:     actor.String(),
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
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	baseAssetAddress := storage.NAIAddress
	marketplaceAssetAddress := storage.AssetAddressFractional(datasetAddress)

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 100})

	claimMarketplacePaymentBenchmark := &chaintest.ActionBenchmark{
		Name:  "ClaimMarketplacePaymentBenchmark",
		Actor: actor,
		Action: &ClaimMarketplacePayment{
			MarketplaceAssetAddress: marketplaceAssetAddress,
			PaymentAssetAddress:     baseAssetAddress,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			metadata := map[string]string{
				"datasetAddress":          datasetAddress.String(),
				"marketplaceAssetAddress": marketplaceAssetAddress.String(),
				"datasetPricePerBlock":    "100",
				"paymentAssetAddress":     baseAssetAddress.String(),
				"publisher":               actor.String(),
				"lastClaimedBlock":        "0",
				"subscriptions":           "0",
				"paymentRemaining":        "100",
				"paymentClaimed":          "0",
			}
			metadataBytes, err := utils.MapToBytes(metadata)
			require.NoError(err)
			require.NoError(storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the payment was correctly claimed
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, baseAssetAddress, actor)
			require.NoError(err)
			require.Equal(uint64(100), balance) // 100 units claimed

			// Check if metadata was updated correctly
			_, _, _, _, metadata, _, _, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, marketplaceAssetAddress)
			require.NoError(err)
			metadataMap, err := utils.BytesToMap(metadata)
			require.NoError(err)
			require.Equal("0", metadataMap["paymentRemaining"]) // 1000 - 100 claimed
			require.Equal("100", metadataMap["paymentClaimed"])
			require.Equal("100", metadataMap["lastClaimedBlock"])
		},
	}

	ctx := context.Background()
	claimMarketplacePaymentBenchmark.Run(ctx, b)
}
