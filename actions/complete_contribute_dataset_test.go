// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/dataset"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestCompleteContributeDatasetAction(t *testing.T) {
	const (
		dataLocation   = "default"
		dataIdentifier = "data_id_1234"
	)

	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	datasetContributionID := storage.DatasetContributionID(datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor)
	nftAddress := codec.CreateAddress(nconsts.AssetFractionalTokenID, datasetContributionID)

	tests := []chaintest.ActionTest{
		{
			Name:  "WrongOwner",
			Actor: codectest.NewRandomAddress(), // Not the owner of the dataset
			Action: &CompleteContributeDataset{
				DatasetContributionID: datasetContributionID.String(),
				DatasetAddress:        datasetAddress,
				DatasetContributor:    actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set dataset with a different owner
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrWrongOwner,
		},
		{
			Name:  "DatasetAlreadyOnSale",
			Actor: actor,
			Action: &CompleteContributeDataset{
				DatasetContributionID: datasetContributionID.String(),
				DatasetAddress:        datasetAddress,
				DatasetContributor:    actor,
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
			Name:  "ContributionAlreadyCompleted",
			Actor: actor,
			Action: &CompleteContributeDataset{
				DatasetContributionID: datasetContributionID.String(),
				DatasetAddress:        datasetAddress,
				DatasetContributor:    actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid contribution
				require.NoError(t, storage.SetDatasetContributionInfo(context.Background(), store, datasetContributionID, datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor, true))
				// Set valid dataset
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrDatasetContributionAlreadyComplete,
		},
		{
			Name:  "DatasetAddressMismatch",
			Actor: actor,
			Action: &CompleteContributeDataset{
				DatasetContributionID: datasetContributionID.String(),
				DatasetAddress:        datasetAddress,
				DatasetContributor:    actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid contribution
				require.NoError(t, storage.SetDatasetContributionInfo(context.Background(), store, datasetContributionID, codectest.NewRandomAddress(), []byte(dataLocation), []byte(dataIdentifier), actor, false))
				// Set valid dataset
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				return store
			}(),
			ExpectedErr: ErrDatasetAddressMismatch,
		},
		{
			Name:  "DatasetContributorMismatch",
			Actor: actor,
			Action: &CompleteContributeDataset{
				DatasetContributionID: datasetContributionID.String(),
				DatasetAddress:        datasetAddress,
				DatasetContributor:    codectest.NewRandomAddress(),
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid contribution
				require.NoError(t, storage.SetDatasetContributionInfo(context.Background(), store, datasetContributionID, datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor, false))
				// Set valid dataset
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))

				return store
			}(),
			ExpectedErr: ErrDatasetContributorMismatch,
		},
		{
			Name:     "ValidCompletion",
			ActionID: ids.GenerateTestID(),
			Actor:    actor,
			Action: &CompleteContributeDataset{
				DatasetContributionID: datasetContributionID.String(),
				DatasetAddress:        datasetAddress,
				DatasetContributor:    actor,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set valid contribution
				require.NoError(t, storage.SetDatasetContributionInfo(context.Background(), store, datasetContributionID, datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor, false))
				// Set valid dataset
				require.NoError(t, storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
				// Create existing NFT
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, datasetAddress, nconsts.AssetFractionalTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(datasetAddress.String()), 1, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				// Set balance to 0
				config := dataset.GetDatasetConfig()
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, config.CollateralAssetAddressForDataContribution, actor, 0))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				config := dataset.GetDatasetConfig()

				// Check if the balance is correctly updated
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, config.CollateralAssetAddressForDataContribution, actor)
				require.NoError(t, err)
				require.Equal(t, config.CollateralAmountForDataContribution, balance) // Collateral refunded

				// Ensure total supply was increased
				_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, datasetAddress)
				require.NoError(t, err)
				require.Equal(t, uint64(2), totalSupply)

				// Check if NFT was created correctly
				assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, nftAddress)
				require.NoError(t, err)
				require.Equal(t, nconsts.AssetNonFungibleTokenID, assetType)
				require.Equal(t, "name", string(name))
				require.Equal(t, "SYM-1", string(symbol))
				require.Equal(t, uint8(0), decimals)
				require.Equal(t, datasetAddress.String(), string(uri))
				require.Equal(t, uint64(1), totalSupply)
				require.Equal(t, uint64(1), maxSupply)
				require.Equal(t, actor, owner)
				require.Equal(t, codec.EmptyAddress, mintAdmin)
				require.Equal(t, codec.EmptyAddress, pauseUnpauseAdmin)
				require.Equal(t, codec.EmptyAddress, freezeUnfreezeAdmin)
				require.Equal(t, codec.EmptyAddress, enableDisableKYCAccountAdmin)
				// Check metadata
				metadataMap, err := utils.BytesToMap(metadata)
				require.NoError(t, err)
				require.Equal(t, "default", metadataMap["dataLocation"])
				require.Equal(t, "data_id_1234", metadataMap["dataIdentifier"])
				// Check NFT balance
				balance, err = storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
			},
			ExpectedOutputs: &CompleteContributeDatasetResult{
				Actor:                    actor.String(),
				Receiver:                 actor.String(),
				CollateralAssetAddress:   dataset.GetDatasetConfig().CollateralAssetAddressForDataContribution.String(),
				CollateralAmountRefunded: dataset.GetDatasetConfig().CollateralAmountForDataContribution,
				DatasetChildNftAddress:   nftAddress.String(),
				To:                       actor.String(),
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
	const (
		dataLocation   = "default"
		dataIdentifier = "data_id_1234"
	)

	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	datasetContributionID := storage.DatasetContributionID(datasetAddress, []byte(dataLocation), []byte(dataIdentifier), actor)
	nftAddress := storage.AssetAddressNFT(datasetAddress, []byte("metadata"), actor)

	completeContributeDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "CompleteContributeDatasetBenchmark",
		Actor: actor,
		Action: &CompleteContributeDataset{
			DatasetContributionID: datasetContributionID.String(),
			DatasetAddress:        datasetAddress,
			DatasetContributor:    actor,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set valid dataset
			require.NoError(storage.SetDatasetInfo(context.Background(), store, datasetAddress, []byte("Valid Name"), []byte("Valid Description"), []byte("Science"), []byte("MIT"), []byte("MIT"), []byte("http://license-url.com"), []byte("Metadata"), true, codec.EmptyAddress, codec.EmptyAddress, 0, 100, 0, 100, 0, actor))
			require.NoError(storage.SetAssetInfo(context.Background(), store, datasetAddress, nconsts.AssetFractionalTokenID, []byte("name"), []byte("SYM"), 0, []byte("metadata"), []byte(datasetAddress.String()), 1, 0, actor, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			// Set balance to 0
			config := dataset.GetDatasetConfig()
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, config.CollateralAssetAddressForDataContribution, actor, 0))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			config := dataset.GetDatasetConfig()

			// Check if the balance is correctly updated
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, config.CollateralAssetAddressForDataContribution, actor)
			require.NoError(err)
			require.Equal(config.CollateralAmountForDataContribution, balance) // Collateral refunded

			// Ensure total supply was increased
			_, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, store, datasetAddress)
			require.NoError(err)
			require.Equal(uint64(2), totalSupply)

			// Check if NFT was created correctly
			assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoNoController(ctx, store, nftAddress)
			require.NoError(err)
			require.Equal(nconsts.AssetNonFungibleTokenID, assetType)
			require.Equal("name", string(name))
			require.Equal("SYM-1", string(symbol))
			require.Equal(uint8(0), decimals)
			require.Equal("metadata", string(metadata))
			require.Equal(datasetAddress.String(), string(uri))
			require.Equal(uint64(1), totalSupply)
			require.Equal(uint64(1), maxSupply)
			require.Equal(actor, owner)
			require.Equal(codec.EmptyAddress, mintAdmin)
			require.Equal(codec.EmptyAddress, pauseUnpauseAdmin)
			require.Equal(codec.EmptyAddress, freezeUnfreezeAdmin)
			require.Equal(codec.EmptyAddress, enableDisableKYCAccountAdmin)
			// Check NFT balance
			balance, err = storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, actor)
			require.NoError(err)
			require.Equal(uint64(1), balance)
		},
	}

	ctx := context.Background()
	completeContributeDatasetBenchmark.Run(ctx, b)
}
