// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// datasets_test.go
package integration

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ = ginkgo.Describe("datasets", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("create a dataset", func() {
		ginkgo.By("create a new dataset (no metadata)", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1001,
				},
				[]chain.Action{&actions.CreateDataset{
					AssetID:            ids.Empty,
					Name:               []byte("n00"),
					Description:        []byte("s00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					Metadata:           nil,
					IsCommunityDataset: false,
				}},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// too large)
			msg, err := tx.Digest()
			require.NoError(err)
			auth, err := factory.Sign(msg)
			require.NoError(err)
			tx.Auth = auth
			p := codec.NewWriter(0, consts.MaxInt) // test codec growth
			require.NoError(tx.Marshal(p))
			require.NoError(p.Err())
			_, err = instances[0].cli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			require.ErrorContains(err, "Bytes field is not populated")
		})

		ginkgo.By("create dataset with too long of metadata", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				[]chain.Action{&actions.CreateDataset{
					AssetID:            ids.Empty,
					Name:               []byte("n00"),
					Description:        []byte("d00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					Metadata:           make([]byte, actions.MaxDatasetMetadataSize*2),
					IsCommunityDataset: false,
				}},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// too large)
			msg, err := tx.Digest()
			require.NoError(err)
			auth, err := factory.Sign(msg)
			require.NoError(err)
			tx.Auth = auth
			p := codec.NewWriter(0, consts.MaxInt) // test codec growth
			require.NoError(tx.Marshal(p))
			require.NoError(p.Err())
			_, err = instances[0].cli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			require.ErrorContains(err, "size is larger than limit")
		})

		ginkgo.By("create a new dataset (solo contributor dataset)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateDataset{
					AssetID:            ids.Empty,
					Name:               asset1,
					Description:        []byte("d00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					Metadata:           asset1,
					IsCommunityDataset: false,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			asset1ID = chain.CreateActionID(tx.ID(), 0)

			// Check dataset info
			exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(description), []byte("d00"))
			require.Equal([]byte(categories), []byte("c00"))
			require.Equal([]byte(licenseName), []byte("l00"))
			require.Equal([]byte(licenseSymbol), []byte("ls00"))
			require.Equal([]byte(licenseURL), []byte("lu00"))
			require.Equal([]byte(metadata), asset1)
			require.False(isCommunityDataset)
			require.Equal(saleID, ids.Empty.String())
			require.Equal(baseAsset, ids.Empty.String())
			require.Zero(basePrice)
			require.Equal(revenueModelDataShare, uint8(100))
			require.Equal(revenueModelMetadataShare, uint8(0))
			require.Equal(revenueModelDataOwnerCut, uint8(100))
			require.Equal(revenueModelMetadataOwnerCut, uint8(0))
			require.Equal(owner, sender)

			// Check asset info
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), []byte("d00"))
			require.Equal([]byte(uri), []byte("d00"))
			require.Equal(totalSupply, uint64(1))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)

			// Check NFT info
			nftID := nchain.GenerateIDWithIndex(asset1ID, 0)
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))

			exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(collectionID, asset1ID.String())
			require.Equal(uniqueID, uint64(0))
			require.Equal([]byte(uri), []byte("d00"))
			require.Equal([]byte(metadata), []byte("d00"))
			require.Equal(owner, sender)
		})

		ginkgo.By("create a new dataset (community contributor dataset)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateDataset{
					AssetID:            ids.Empty,
					Name:               asset1,
					Description:        []byte("d00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					Metadata:           asset1,
					IsCommunityDataset: true,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			asset1ID = chain.CreateActionID(tx.ID(), 0)

			// Check dataset info
			exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(description), []byte("d00"))
			require.Equal([]byte(categories), []byte("c00"))
			require.Equal([]byte(licenseName), []byte("l00"))
			require.Equal([]byte(licenseSymbol), []byte("ls00"))
			require.Equal([]byte(licenseURL), []byte("lu00"))
			require.Equal([]byte(metadata), asset1)
			require.True(isCommunityDataset)
			require.Equal(saleID, ids.Empty.String())
			require.Equal(baseAsset, ids.Empty.String())
			require.Zero(basePrice)
			require.Equal(revenueModelDataShare, uint8(100))
			require.Equal(revenueModelMetadataShare, uint8(0))
			require.Equal(revenueModelDataOwnerCut, uint8(10))
			require.Equal(revenueModelMetadataOwnerCut, uint8(0))
			require.Equal(owner, sender)

			// Check asset info
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), []byte("d00"))
			require.Equal([]byte(uri), []byte("d00"))
			require.Equal(totalSupply, uint64(1))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)

			// Check NFT info
			nftID := nchain.GenerateIDWithIndex(asset1ID, 0)
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))

			exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(collectionID, asset1ID.String())
			require.Equal(uniqueID, uint64(0))
			require.Equal([]byte(uri), []byte("d00"))
			require.Equal([]byte(metadata), []byte("d00"))
			require.Equal(owner, sender)
		})

		ginkgo.By("create a new dataset (after an asset is already created)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetDatasetTokenID,
					Name:                         asset1,
					Symbol:                       asset1Symbol,
					Decimals:                     0,
					Metadata:                     asset1,
					URI:                          asset1,
					MaxSupply:                    uint64(0),
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			asset1ID = chain.CreateActionID(tx.ID(), 0)
			parser, err = instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err = instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateDataset{
					AssetID:            asset1ID,
					Name:               asset1,
					Description:        []byte("d00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					Metadata:           asset1,
					IsCommunityDataset: false,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept = expectBlk(instances[0])
			results = accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			// Check dataset info
			exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(description), []byte("d00"))
			require.Equal([]byte(categories), []byte("c00"))
			require.Equal([]byte(licenseName), []byte("l00"))
			require.Equal([]byte(licenseSymbol), []byte("ls00"))
			require.Equal([]byte(licenseURL), []byte("lu00"))
			require.Equal([]byte(metadata), asset1)
			require.False(isCommunityDataset)
			require.Equal(saleID, ids.Empty.String())
			require.Equal(baseAsset, ids.Empty.String())
			require.Zero(basePrice)
			require.Equal(revenueModelDataShare, uint8(100))
			require.Equal(revenueModelMetadataShare, uint8(0))
			require.Equal(revenueModelDataOwnerCut, uint8(100))
			require.Equal(revenueModelMetadataOwnerCut, uint8(0))
			require.Equal(owner, sender)

			// Check asset info
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset1)
			require.Equal([]byte(uri), asset1)
			require.Equal(totalSupply, uint64(1))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)

			// Check NFT info
			nftID := nchain.GenerateIDWithIndex(asset1ID, 0)
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))

			exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(collectionID, asset1ID.String())
			require.Equal(uniqueID, uint64(0))
			require.Equal([]byte(uri), asset1)
			require.Equal([]byte(metadata), asset1)
			require.Equal(owner, sender)
		})
	})

	ginkgo.It("update dataset", func() {
		ginkgo.By("update a dataset that doesn't exist", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.UpdateDataset{
					Dataset:            ids.GenerateTestID(),
					Name:               []byte("n00"),
					Description:        []byte("d00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					IsCommunityDataset: false,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			result := results[0]
			require.False(result.Success)
			require.Contains(string(result.Error), "dataset not found")
		})

		ginkgo.By("update a dataset(no field is updated)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.UpdateDataset{
					Dataset:            asset1ID,
					Name:               asset1,
					Description:        []byte("d00"),
					Categories:         []byte("c00"),
					LicenseName:        []byte("l00"),
					LicenseSymbol:      []byte("ls00"),
					LicenseURL:         []byte("lu00"),
					IsCommunityDataset: false,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			result := results[0]
			require.False(result.Success)
			require.Contains(string(result.Error), "must update at least one field")
		})

		ginkgo.By("update an existing dataset(convert solo contributor to community dataset)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.UpdateDataset{
					Dataset:            asset1ID,
					Name:               asset1,
					Description:        []byte("d00-updated"),
					Categories:         []byte("c00-updated"),
					LicenseName:        []byte("l00-updated"),
					LicenseSymbol:      []byte("ls00up"),
					LicenseURL:         []byte("lu00-updated"),
					IsCommunityDataset: true,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			// Check dataset info
			exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, _, _, _, _, _, revenueModelDataOwnerCut, _, _, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(description), []byte("d00-updated"))
			require.Equal([]byte(categories), []byte("c00-updated"))
			require.Equal([]byte(licenseName), []byte("l00-updated"))
			require.Equal([]byte(licenseSymbol), []byte("ls00up"))
			require.Equal([]byte(licenseURL), []byte("lu00-updated"))
			require.Equal([]byte(metadata), asset1)
			require.True(isCommunityDataset)
			require.Equal(revenueModelDataOwnerCut, uint8(10))
		})
	})
})
