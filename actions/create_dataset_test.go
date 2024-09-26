// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"

	nchain "github.com/nuklai/nuklaivm/chain"
)

func TestCreateDatasetAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	datasetID := ids.GenerateTestID()

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidName",
			Actor: addr,
			Action: &CreateDataset{
				Name:          []byte("Na"), // Invalid name (too short)
				Description:   []byte("Valid Description"),
				Categories:    []byte("Science"),
				LicenseName:   []byte("MIT"),
				LicenseSymbol: []byte("MIT"),
				LicenseURL:    []byte("http://license-url.com"),
				Metadata:      []byte("Metadata"),
			},
			ExpectedErr: ErrOutputNameInvalid,
		},
		{
			Name:  "InvalidDescription",
			Actor: addr,
			Action: &CreateDataset{
				Name:          []byte("Valid Name"),
				Description:   []byte("De"), // Invalid description (too short)
				Categories:    []byte("Science"),
				LicenseName:   []byte("MIT"),
				LicenseSymbol: []byte("MIT"),
				LicenseURL:    []byte("http://license-url.com"),
				Metadata:      []byte("Metadata"),
			},
			ExpectedErr: ErrOutputDescriptionInvalid,
		},
		{
			Name:  "InvalidCategories",
			Actor: addr,
			Action: &CreateDataset{
				Name:          []byte("Valid Name"),
				Description:   []byte("Valid Description"),
				Categories:    []byte("Ca"), // Invalid categories (too short)
				LicenseName:   []byte("MIT"),
				LicenseSymbol: []byte("MIT"),
				LicenseURL:    []byte("http://license-url.com"),
				Metadata:      []byte("Metadata"),
			},
			ExpectedErr: ErrOutputCategoriesInvalid,
		},
		{
			Name:  "InvalidLicenseName",
			Actor: addr,
			Action: &CreateDataset{
				Name:          []byte("Valid Name"),
				Description:   []byte("Valid Description"),
				Categories:    []byte("Science"),
				LicenseName:   []byte("Li"), // Invalid license name (too short)
				LicenseSymbol: []byte("MIT"),
				LicenseURL:    []byte("http://license-url.com"),
				Metadata:      []byte("Metadata"),
			},
			ExpectedErr: ErrOutputLicenseNameInvalid,
		},
		{
			Name:  "InvalidLicenseSymbol",
			Actor: addr,
			Action: &CreateDataset{
				Name:          []byte("Valid Name"),
				Description:   []byte("Valid Description"),
				Categories:    []byte("Science"),
				LicenseName:   []byte("MIT"),
				LicenseSymbol: []byte("Mi"), // Invalid license symbol (too short)
				LicenseURL:    []byte("http://license-url.com"),
				Metadata:      []byte("Metadata"),
			},
			ExpectedErr: ErrOutputLicenseSymbolInvalid,
		},
		{
			Name:  "InvalidLicenseURL",
			Actor: addr,
			Action: &CreateDataset{
				Name:          []byte("Valid Name"),
				Description:   []byte("Valid Description"),
				Categories:    []byte("Science"),
				LicenseName:   []byte("MIT"),
				LicenseSymbol: []byte("MIT"),
				LicenseURL:    []byte("ur"), // Invalid license URL (too short)
				Metadata:      []byte("Metadata"),
			},
			ExpectedErr: ErrOutputLicenseURLInvalid,
		},
		{
			Name:     "ValidDatasetCreation",
			ActionID: datasetID,
			Actor:    addr,
			Action: &CreateDataset{
				Name:               []byte("Dataset Name"),
				Description:        []byte("This is a dataset"),
				Categories:         []byte("Science"),
				LicenseName:        []byte("MIT"),
				LicenseSymbol:      []byte("MIT"),
				LicenseURL:         []byte("http://license-url.com"),
				Metadata:           []byte("Dataset Metadata"),
				IsCommunityDataset: true,
			},
			State: chaintest.NewInMemoryStore(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if the dataset was created correctly
				nftID := nchain.GenerateIDWithIndex(datasetID, 0)
				exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, _, _, _, _, _, owner, err := storage.GetDataset(ctx, store, datasetID)
				require.NoError(t, err)
				require.True(t, exists)
				require.Equal(t, "Dataset Name", string(name))
				require.Equal(t, "This is a dataset", string(description))
				require.Equal(t, "Science", string(categories))
				require.Equal(t, "MIT", string(licenseName))
				require.Equal(t, "MIT", string(licenseSymbol))
				require.Equal(t, "http://license-url.com", string(licenseURL))
				require.Equal(t, "Dataset Metadata", string(metadata))
				require.True(t, isCommunityDataset)
				require.Equal(t, ids.Empty, saleID)
				require.Equal(t, ids.Empty, baseAsset)
				require.Equal(t, addr, owner)

				// Check if balance was updated correctly
				balance, err := storage.GetBalance(ctx, store, addr, datasetID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)

				// Check if NFT was created correctly
				nftExists, _, _, _, _, nftOwner, _ := storage.GetAssetNFT(ctx, store, nftID)
				require.True(t, nftExists)
				require.Equal(t, addr, nftOwner)
			},
			ExpectedOutputs: &CreateDatasetResult{
				DatasetID:   datasetID.String(),
				ParentNftID: nchain.GenerateIDWithIndex(datasetID, 0).String(),
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
	assetID := ids.GenerateTestID()

	createDatasetBenchmark := &chaintest.ActionBenchmark{
		Name:  "CreateDatasetBenchmark",
		Actor: actor,
		Action: &CreateDataset{
			Name:               []byte("Benchmark Dataset"),
			Description:        []byte("This is a benchmark dataset"),
			Categories:         []byte("Science"),
			LicenseName:        []byte("MIT"),
			LicenseSymbol:      []byte("MIT"),
			LicenseURL:         []byte("http://license-url.com"),
			Metadata:           []byte("Benchmark Metadata"),
			IsCommunityDataset: true,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if the dataset was created correctly
			nftID := nchain.GenerateIDWithIndex(assetID, 0)
			exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, _, _, _, _, _, owner, err := storage.GetDataset(ctx, store, assetID)
			require.NoError(err)
			require.True(exists)
			require.Equal(b, "Benchmark Dataset", string(name))
			require.Equal(b, "This is a benchmark dataset", string(description))
			require.Equal(b, "Science", string(categories))
			require.Equal(b, "MIT", string(licenseName))
			require.Equal(b, "MIT", string(licenseSymbol))
			require.Equal(b, "http://license-url.com", string(licenseURL))
			require.Equal(b, "Benchmark Metadata", string(metadata))
			require.True(isCommunityDataset)
			require.Equal(b, ids.Empty, saleID)
			require.Equal(b, ids.Empty, baseAsset)
			require.Equal(b, actor, owner)

			// Check if balance was updated correctly
			balance, err := storage.GetBalance(ctx, store, actor, assetID)
			require.NoError(err)
			require.Equal(b, uint64(1), balance)

			// Check if NFT was created correctly
			nftExists, _, _, _, _, nftOwner, _ := storage.GetAssetNFT(ctx, store, nftID)
			require.True(nftExists)
			require.Equal(b, actor, nftOwner)
		},
	}

	ctx := context.Background()
	createDatasetBenchmark.Run(ctx, b)
}
