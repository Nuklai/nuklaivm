// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"fmt"
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

func TestSubscribeDatasetMarketplaceAction(t *testing.T) {
	mockEmission := emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 1})

	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	baseAssetAddress := storage.NAIAddress
	marketplaceAssetAddress := storage.AssetAddressFractional(datasetAddress)
	nftAddress := storage.AssetAddressNFT(marketplaceAssetAddress, nil, actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "UserAlreadySubscribed",
			Actor: actor,
			Action: &SubscribeDatasetMarketplace{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     baseAssetAddress,
				NumBlocksToSubscribe:    10,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, nftAddress, nconsts.AssetNonFungibleTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(marketplaceAssetAddress.String()), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrUserAlreadySubscribed,
		},
		{
			Name:  "BaseAssetNotSupported",
			Actor: actor,
			Action: &SubscribeDatasetMarketplace{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     codec.EmptyAddress, // Invalid base asset ID
				NumBlocksToSubscribe:    10,
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
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrPaymentAssetNotSupported,
		},
		{
			Name:  "ValidSubscription",
			Actor: actor,
			Action: &SubscribeDatasetMarketplace{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     baseAssetAddress,
				NumBlocksToSubscribe:    10,
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
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				// Set base asset balance to sufficient amount
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, baseAssetAddress, actor, 5000))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly deducted
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, baseAssetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(4000), balance) // 5000 - 1000 = 4000

				// Check if the subscription NFT was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, nftAddress)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetNonFungibleTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM-0", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, marketplaceAssetAddress.String(), string(uri))
				require.Equal(t, uint64(1), totalSupply)
				require.Equal(t, uint64(1), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
				// Check metadata of NFT
				metadataMap, err := utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, datasetAddress.String(), metadataMap["datasetAddress"])
				require.Equal(t, marketplaceAssetAddress.String(), metadataMap["marketplaceAssetAddress"])
				require.Equal(t, "100", metadataMap["datasetPricePerBlock"])
				require.Equal(t, baseAssetAddress.String(), metadataMap["paymentAssetAddress"])
				require.Equal(t, "1000", metadataMap["totalCost"])
				require.Equal(t, fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()), metadataMap["issuanceBlock"])
				require.Equal(t, "10", metadataMap["numBlocksToSubscribe"])
				require.Equal(t, fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()+10), metadataMap["expirationBlock"])

				// Check if the marketplace asset metadata was updated correctly
				_, _, _, _, metadata, _, _, _, _, _, _, _, _, err = storage.GetAssetInfoNoController(ctx, store, marketplaceAssetAddress)
				require.NoError(t, err)
				metadataMap, err = utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, "1000", metadataMap["paymentRemaining"])
				require.Equal(t, "1", metadataMap["subscriptions"])
				require.Equal(t, fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()), metadataMap["lastClaimedBlock"])
			},
			ExpectedOutputs: &SubscribeDatasetMarketplaceResult{
				MarketplaceAssetAddress:          marketplaceAssetAddress.String(),
				MarketplaceAssetNumSubscriptions: 1,
				SubscriptionNftAddress:           nftAddress.String(),
				PaymentAssetAddress:              baseAssetAddress.String(),
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
	mockEmission := emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 1})

	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	baseAssetAddress := storage.NAIAddress
	marketplaceAssetAddress := storage.AssetAddressFractional(datasetAddress)
	nftAddress := storage.AssetAddressNFT(marketplaceAssetAddress, nil, actor)

	subscribeDatasetMarketplaceBenchmark := &chaintest.ActionBenchmark{
		Name:  "SubscribeDatasetMarketplaceBenchmark",
		Actor: actor,
		Action: &SubscribeDatasetMarketplace{
			MarketplaceAssetAddress: marketplaceAssetAddress,
			PaymentAssetAddress:     baseAssetAddress,
			NumBlocksToSubscribe:    10,
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
				"paymentRemaining":        "0",
				"paymentClaimed":          "0",
			}
			metadataBytes, err := utils.MapToBytes(metadata)
			require.NoError(err)
			require.NoError(storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 0, metadataBytes, []byte(marketplaceAssetAddress.String()), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			// Set base asset balance to sufficient amount
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, baseAssetAddress, actor, 5000))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly deducted
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, baseAssetAddress, actor)
			require.NoError(err)
			require.Equal(uint64(4000), balance) // 5000 - 1000 = 4000

			// Check if the subscription NFT was created correctly
			assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, nftAddress)
			require.NoError(err)
			require.Equal(nconsts.AssetNonFungibleTokenID, assetType)
			require.Equal("name", string(name))
			require.Equal("SYM-0", string(symbol))
			require.Equal(uint8(0), decimals)
			require.Equal(marketplaceAssetAddress.String(), string(uri))
			require.Equal(uint64(1), totalSupply)
			require.Equal(uint64(1), maxSupply)
			require.Equal(actor, owner)
			require.Equal(codec.EmptyAddress, mintAdmin)
			require.Equal(codec.EmptyAddress, pauseUnpauseAdmin)
			require.Equal(codec.EmptyAddress, freezeUnfreezeAdmin)
			require.Equal(codec.EmptyAddress, enableDisableKYCAccountAdmin)
			// Check metadata of NFT
			metadataMap, err := utils.BytesToMap(metadata)
			require.NoError(err)
			require.Equal(datasetAddress.String(), metadataMap["datasetAddress"])
			require.Equal(marketplaceAssetAddress.String(), metadataMap["marketplaceAssetAddress"])
			require.Equal("100", metadataMap["datasetPricePerBlock"])
			require.Equal(baseAssetAddress.String(), metadataMap["paymentAssetAddress"])
			require.Equal("1000", metadataMap["totalCost"])
			require.Equal(fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()), metadataMap["issuanceBlock"])
			require.Equal("10", metadataMap["numBlocksToSubscribe"])
			require.Equal(fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()+10), metadataMap["expirationBlock"])

			// Check if the marketplace asset metadata was updated correctly
			_, _, _, _, metadata, _, _, _, _, _, _, _, _, err = storage.GetAssetInfoNoController(ctx, store, marketplaceAssetAddress)
			require.NoError(err)
			metadataMap, err = utils.BytesToMap(metadata)
			require.NoError(err)
			require.Equal("1000", metadataMap["paymentRemaining"])
			require.Equal("1", metadataMap["subscriptions"])
			require.Equal(fmt.Sprint(mockEmission.GetLastAcceptedBlockHeight()), metadataMap["lastClaimedBlock"])
		},
	}

	ctx := context.Background()
	subscribeDatasetMarketplaceBenchmark.Run(ctx, b)
}
