// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

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
	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	baseAssetAddress := storage.NAIAddress
	marketplaceAssetAddress := storage.AssetAddressFractional(datasetAddress)

	tests := []chaintest.ActionTest{
		{
			Name:  "MarketplaceAssetAlreadyExists",
			Actor: actor,
			Action: &PublishDatasetMarketplace{
				DatasetAddress:       datasetAddress,
				PaymentAssetAddress:  baseAssetAddress,
				DatasetPricePerBlock: 100,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, marketplaceAssetAddress, nconsts.AssetMarketplaceTokenID, []byte("name"), []byte("SYM"), 9, []byte("metadata"), []byte("uri"), 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			ExpectedErr: ErrAssetExists,
		},
		{
			Name:  "WrongOwner",
			Actor: codectest.NewRandomAddress(), // Not the owner of the dataset
			Action: &PublishDatasetMarketplace{
				DatasetAddress:       datasetAddress,
				PaymentAssetAddress:  baseAssetAddress,
				DatasetPricePerBlock: 100,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the dataset with a different owner
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrWrongOwner,
		},
		{
			Name:  "ValidPublishDataset",
			Actor: actor,
			Action: &PublishDatasetMarketplace{
				DatasetAddress:       datasetAddress,
				PaymentAssetAddress:  baseAssetAddress,
				DatasetPricePerBlock: 100,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the base asset with the required details
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the marketplace asset was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, marketplaceAssetAddress)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetMarketplaceTokenID, assetType)
				require.Equal(t, storage.MarketplaceAssetName, string(name))
				require.Equal(t, storage.MarketplaceAssetSymbol, string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, marketplaceAssetAddress.String(), string(uri))
				require.Equal(t, uint64(0), totalSupply)
				require.Equal(t, uint64(0), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)

				// Check metadata
				metadataMap, err := utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, datasetAddress.String(), metadataMap["datasetAddress"])
				require.Equal(t, marketplaceAssetAddress.String(), metadataMap["marketplaceAssetAddress"])
				require.Equal(t, "100", metadataMap["datasetPricePerBlock"])
				require.Equal(t, baseAssetAddress.String(), metadataMap["paymentAssetAddress"])
				require.Equal(t, actor.String(), metadataMap["publisher"])
				require.Equal(t, "0", metadataMap["lastClaimedBlock"])
				require.Equal(t, "0", metadataMap["subscriptions"])
				require.Equal(t, "0", metadataMap["paymentRemaining"])
				require.Equal(t, "0", metadataMap["paymentClaimed"])

				// Check if the dataset was updated correctly
				_, _, _, _, _, _, _, _, mAddr, baseAsset, basePrice, _, _, _, _, _, err := storage.GetDatasetInfoNoController(ctx, store, datasetAddress)
				require.NoError(t, err)
				require.Equal(t, marketplaceAssetAddress, mAddr)
				require.Equal(t, baseAssetAddress, baseAsset)
				require.Equal(t, uint64(100), basePrice)
			},
			ExpectedOutputs: &PublishDatasetMarketplaceResult{
				MarketplaceAssetAddress: marketplaceAssetAddress,
				PaymentAssetAddress:     baseAssetAddress,
				DatasetPricePerBlock:    100,
				Publisher:               actor,
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
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	baseAssetAddress := storage.NAIAddress
	marketplaceAssetAddress := storage.AssetAddressFractional(datasetAddress)

	publishDatasetMarketplaceBenchmark := &chaintest.ActionBenchmark{
		Name:  "PublishDatasetMarketplaceBenchmark",
		Actor: actor,
		Action: &PublishDatasetMarketplace{
			DatasetAddress:       datasetAddress,
			PaymentAssetAddress:  baseAssetAddress,
			DatasetPricePerBlock: 100,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the marketplace asset was created correctly
			assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, marketplaceAssetAddress)
			require.NoError(err)
			require.Equal(nconsts.AssetMarketplaceTokenID, assetType)
			require.Equal(storage.MarketplaceAssetName, string(name))
			require.Equal(storage.MarketplaceAssetSymbol, string(symbol))
			require.Equal(uint8(0), decimals)
			require.Equal(marketplaceAssetAddress.String(), string(uri))
			require.Equal(uint64(0), totalSupply)
			require.Equal(uint64(0), maxSupply)
			require.Equal(actor, owner)
			require.Equal(codec.EmptyAddress, mintAdmin)
			require.Equal(codec.EmptyAddress, pauseUnpauseAdmin)
			require.Equal(codec.EmptyAddress, freezeUnfreezeAdmin)
			require.Equal(codec.EmptyAddress, enableDisableKYCAccountAdmin)

			// Check metadata
			metadataMap, err := utils.BytesToMap(metadata)
			require.NoError(err)
			require.Equal(datasetAddress.String(), metadataMap["datasetAddress"])
			require.Equal(marketplaceAssetAddress.String(), metadataMap["marketplaceAssetAddress"])
			require.Equal("100", metadataMap["datasetPricePerBlock"])
			require.Equal(baseAssetAddress.String(), metadataMap["paymentAssetAddress"])
			require.Equal(actor.String(), metadataMap["publisher"])
			require.Equal("0", metadataMap["lastClaimedBlock"])
			require.Equal("0", metadataMap["subscriptions"])
			require.Equal("0", metadataMap["paymentRemaining"])
			require.Equal("0", metadataMap["paymentClaimed"])

			// Check if the dataset was updated correctly
			_, _, _, _, _, _, _, _, mAddr, baseAsset, basePrice, _, _, _, _, _, err := storage.GetDatasetInfoNoController(ctx, store, datasetAddress)
			require.NoError(err)
			require.Equal(marketplaceAssetAddress, mAddr)
			require.Equal(baseAssetAddress, baseAsset)
			require.Equal(uint64(100), basePrice)
		},
	}

	ctx := context.Background()
	publishDatasetMarketplaceBenchmark.Run(ctx, b)
}
