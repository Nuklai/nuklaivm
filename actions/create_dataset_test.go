// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestCreateDatasetAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	nftAddress := storage.AssetAddressNFT(datasetAddress, []byte("metadata"), actor)

	tests := []chaintest.ActionTest{
		{
			Name:  "NameTooShort",
			Actor: actor,
			Action: &CreateDataset{
				AssetAddress:  datasetAddress,
				Name:          "na", // Invalid name (too short)
				Description:   "Valid Description",
				Categories:    "Science",
				LicenseName:   "MIT",
				LicenseSymbol: "MIT",
				LicenseURL:    "http://license-url.com",
				Metadata:      "Metadata",
			},
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "NameTooLong",
			Actor: actor,
			Action: &CreateDataset{
				AssetAddress:  datasetAddress,
				Name:          strings.Repeat("n", storage.MaxNameSize+1), // Invalid name (too long)
				Description:   "Valid Description",
				Categories:    "Science",
				LicenseName:   "MIT",
				LicenseSymbol: "MIT",
				LicenseURL:    "http://license-url.com",
				Metadata:      "Metadata",
			},
			ExpectedErr: ErrNameInvalid,
		},
		{
			Name:  "ValidDatasetCreation",
			Actor: actor,
			Action: &CreateDataset{
				AssetAddress:       datasetAddress,
				Name:               "Valid Name",
				Description:        "Valid Description",
				Categories:         "Science",
				LicenseName:        "MIT",
				LicenseSymbol:      "MIT",
				LicenseURL:         "http://license-url.com",
				Metadata:           "Metadata",
				IsCommunityDataset: true,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Create the asset first
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, datasetAddress, nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), []byte("uri"), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the dataset was created correctly
				name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, marketplaceAssetAddress, baseAssetAddress, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, store, datasetAddress)
				require.NoError(t, err)
				require.Equal(t, "Valid Name", string(name))
				require.Equal(t, "Valid Description", string(description))
				require.Equal(t, "Science", string(categories))
				require.Equal(t, "MIT", string(licenseName))
				require.Equal(t, "MIT", string(licenseSymbol))
				require.Equal(t, "http://license-url.com", string(licenseURL))
				require.Equal(t, "Metadata", string(metadata))
				require.True(t, isCommunityDataset)
				require.Equal(t, codec.EmptyAddress, marketplaceAssetAddress)
				require.Equal(t, codec.EmptyAddress, baseAssetAddress)
				require.Equal(t, actor, owner)
			},
			ExpectedOutputs: &CreateDatasetResult{
				DatasetAddress:          datasetAddress.String(),
				DatasetParentNftAddress: nftAddress.String(),
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkCreateDataset(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	datasetAddress := storage.AssetAddress(nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), actor)
	// nftAddress := storage.AssetAddressNFT(datasetAddress, []byte("metadata"), actor)

	createDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "CreateDatasetBenchmark",
		Actor: actor,
		Action: &CreateDataset{
			AssetAddress:       datasetAddress,
			Name:               "Valid Name",
			Description:        "Valid Description",
			Categories:         "Science",
			LicenseName:        "MIT",
			LicenseSymbol:      "MIT",
			LicenseURL:         "http://license-url.com",
			Metadata:           "Metadata",
			IsCommunityDataset: true,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Create the asset first
			require.NoError(storage.SetAssetInfo(context.Background(), store, datasetAddress, nconsts.AssetFractionalTokenID, []byte("Valid Name"), []byte("DATASET"), 0, []byte("metadata"), []byte("uri"), 0, 0, actor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the dataset was created correctly
			// Check if the dataset was created correctly
			name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, marketplaceAssetAddress, baseAssetAddress, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, store, datasetAddress)
			require.NoError(err)
			require.Equal("Valid Name", string(name))
			require.Equal("Valid Description", string(description))
			require.Equal("Science", string(categories))
			require.Equal("MIT", string(licenseName))
			require.Equal("MIT", string(licenseSymbol))
			require.Equal("http://license-url.com", string(licenseURL))
			require.Equal("Metadata", string(metadata))
			require.True(isCommunityDataset)
			require.Equal(codec.EmptyAddress, marketplaceAssetAddress)
			require.Equal(codec.EmptyAddress, baseAssetAddress)
			require.Equal(actor, owner)
		},
	}

	ctx := context.Background()
	createDatasetBenchmark.Run(ctx, b)
}
