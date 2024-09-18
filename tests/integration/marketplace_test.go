// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// contributions_test.go
package integration

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"

	"github.com/nuklai/nuklaivm/actions"
	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ = ginkgo.Describe("marketplace", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("initiate data contribution to the dataset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateDataset{
				AssetID:            ids.Empty,
				Name:               asset1,
				Description:        []byte("d01"),
				Categories:         []byte("c01"),
				LicenseName:        []byte("l01"),
				LicenseSymbol:      []byte("ls01"),
				LicenseURL:         []byte("lu01"),
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

		dataset1ID = chain.CreateActionID(tx.ID(), 0)

		// Check asset info
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), dataset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), []byte("d01"))
		require.Equal([]byte(uri), []byte("d01"))
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)

		// Save balance before contribution
		balanceBefore, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)

		// Initiate contribution to dataset
		submit, _, _, err = instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.InitiateContributeDataset{
				Dataset:        dataset1ID,
				DataLocation:   []byte("default"),
				DataIdentifier: []byte("id1"),
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[0])
		results = accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Check balance after contribution
		balanceAfter, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)
		require.Less(balanceAfter, balanceBefore-uint64(1_000_000_000)) // 1 NAI is deducted + fees

		// Check contribution info by interacting with marketplace directly
		dataContributions, err := instances[0].marketplace.GetDataContribution(dataset1ID, rsender)
		require.NoError(err)
		require.NotEmpty(dataContributions)
		require.Equal(len(dataContributions), 1)
		require.Equal([]byte("default"), dataContributions[0].DataLocation)
		require.Equal([]byte("id1"), dataContributions[0].DataIdentifier)

		// Check contribution info by interacting with RPC node
		dataContributionsNew, err := instances[0].ncli.DataContributionPending(context.TODO(), dataset1ID.String())
		require.NoError(err)
		require.NotEmpty(dataContributionsNew)
		require.Equal(len(dataContributionsNew), 1)
		require.Equal("default", dataContributionsNew[0].DataLocation)
		require.Equal("id1", dataContributionsNew[0].DataIdentifier)
	})

	ginkgo.It("complete data contribution to the dataset", func() {
		// Save balance before contribution
		balanceBefore, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)

		// Check asset info before an NFT is minted for data contribution
		exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), dataset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(totalSupply, uint64(1))

		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		// Complete contribution to dataset
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CompleteContributeDataset{
				Dataset:     dataset1ID,
				Contributor: rsender,
				UniqueNFTID: totalSupply,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Check balance after getting the collateral refunded after the contribution is complete
		balanceAfter, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)
		require.GreaterOrEqual(balanceAfter, balanceBefore+uint64(1_000_000_000)-uint64(100_000)) // 1 NAI is refunded but fees is taken

		// Check contribution info
		_, err = instances[0].marketplace.GetDataContribution(dataset1ID, rsender)
		require.Equal(err.Error(), "contribution not found")

		// Check asset info
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, dataset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(2))

		// Check asset info after an NFT is minted for data contribution
		exists, _, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err = instances[0].ncli.Asset(context.TODO(), dataset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(totalSupply, uint64(2))

		// Check NFT that was created for data contribution to the dataset
		nftID := nchain.GenerateIDWithIndex(dataset1ID, totalSupply-1)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		// Check NFT info
		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, dataset1ID.String())
		require.Equal(uniqueID, totalSupply-1)
		require.Equal([]byte(uri), []byte("d01"))
		require.Equal([]byte(metadata), []byte("{\"dataLocation\":\"default\",\"dataIdentifier\":\"id1\"}"))
		require.Equal(owner, sender)
	})

	ginkgo.It("publish dataset to marketplace", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		// Complete contribution to dataset
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.PublishDatasetMarketplace{
				Dataset:   dataset1ID,
				BaseAsset: ids.Empty,
				BasePrice: 100,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		marketplace1ID = chain.CreateActionID(tx.ID(), 0)

		// Check updated dataset info
		exists, _, _, _, _, _, _, _, _, saleID, baseAsset, basePrice, _, _, _, _, _, err := instances[0].ncli.Dataset(context.TODO(), dataset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(saleID, marketplace1ID.String())
		require.Equal(baseAsset, ids.Empty.String())
		require.Equal(basePrice, uint64(100))

		// Check the newly created marketplace asset
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), marketplace1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetMarketplaceTokenDesc)
		require.Equal([]byte(name), nchain.CombineWithPrefix([]byte("Dataset-Marketplace-"), []byte(asset1), 256))
		symbolBytes := nchain.CombineWithPrefix([]byte(""), asset1, 8)
		symbolBytes = nchain.CombineWithPrefix([]byte("DM-"), symbolBytes, 8)
		require.Equal([]byte(symbol), symbolBytes)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), []byte("{\"dataset\":\""+dataset1ID.String()+"\",\"datasetPricePerBlock\":\""+"100"+"\",\"assetForPayment\":\""+ids.Empty.String()+"\",\"publisher\":\""+sender+"\"}"))
		require.Equal(uri, dataset1ID.String())
		require.Zero(totalSupply)
		require.Zero(maxSupply)
		emptyAddress := codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress)
		require.Equal(admin, emptyAddress)
		require.Equal(mintActor, emptyAddress)
		require.Equal(pauseUnpauseActor, emptyAddress)
		require.Equal(freezeUnfreezeActor, emptyAddress)
		require.Equal(enableDisableKYCAccountActor, emptyAddress)
	})

	ginkgo.It("subscribe to a dataset in the marketplace", func() {
		// Save balance before contribution
		balanceBefore, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)

		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		// Complete contribution to dataset
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.SubscribeDatasetMarketplace{
				Dataset:              dataset1ID,
				MarketplaceID:        marketplace1ID,
				AssetForPayment:      ids.Empty,
				NumBlocksToSubscribe: 5,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Check updated dataset info
		exists, _, _, _, _, _, _, _, _, _, _, basePrice, _, _, _, _, _, err := instances[0].ncli.Dataset(context.TODO(), dataset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(basePrice, uint64(100))

		totalCost := 5 * basePrice

		// Check balance after for the asset used for payment
		balanceAfter, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)
		require.Less(balanceAfter, balanceBefore-uint64(100)) // totalCost is deducted + fees

		// Check NFT balance
		nftID := nchain.GenerateIDWithAddress(marketplace1ID, rsender)
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		// Check the  marketplace asset info
		exists, assetType, _, _, _, _, _, totalSupply, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), marketplace1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetMarketplaceTokenDesc)
		require.Equal(totalSupply, uint64(1))

		// Check NFT info
		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, marketplace1ID.String())
		require.Equal(uniqueID, totalSupply)
		require.Equal(uri, dataset1ID.String())
		require.Equal([]byte(metadata), []byte("{\"dataset\":\""+dataset1ID.String()+"\",\"marketplaceID\":\""+marketplace1ID.String()+"\",\"datasetPricePerBlock\":\""+fmt.Sprint(basePrice)+"\",\"totalCost\":\""+fmt.Sprint(totalCost)+"\",\"assetForPayment\":\""+ids.Empty.String()+"\",\"numBlocksToSubscribe\":\""+"5"+"\"}"))
		require.Equal(owner, sender)
	})
})
