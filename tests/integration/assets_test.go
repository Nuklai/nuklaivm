// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// assets_test.go
package integration

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ = ginkgo.Describe("assets", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("mint asset that doesn't exist", func() {
		ginkgo.By("mint a fungible asset that doesn't exist", func() {
			other, err := ed25519.GeneratePrivateKey()
			require.NoError(err)
			assetID := ids.GenerateTestID()
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetFT{
					To:    auth.NewED25519Address(other.PublicKey()),
					Asset: assetID,
					Value: 10,
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
			require.Contains(string(result.Error), "asset missing")

			exists, _, _, _, _, _, _, _, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), assetID.String(), false)
			require.NoError(err)
			require.False(exists)
		})

		ginkgo.By("mint a non-fungible asset that doesn't exist", func() {
			other, err := ed25519.GeneratePrivateKey()
			require.NoError(err)
			assetID := ids.GenerateTestID()
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetNFT{
					To:       auth.NewED25519Address(other.PublicKey()),
					Asset:    assetID,
					UniqueID: 0,
					URI:      []byte("uri"),
					Metadata: []byte("metadata"),
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
			require.Contains(string(result.Error), "asset missing")

			exists, _, _, _, _, _, _, _, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), assetID.String(), false)
			require.NoError(err)
			require.False(exists)
		})
	})

	ginkgo.It("create a new asset", func() {
		ginkgo.By("create a new asset (no metadata)", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1001,
				},
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         []byte("n00"),
					Symbol:                       []byte("s00"),
					Decimals:                     0,
					Metadata:                     nil,
					URI:                          []byte("uri"),
					MaxSupply:                    uint64(0),
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
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

		ginkgo.By("create a new asset (no symbol)", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1001,
				},
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         []byte("n00"),
					Symbol:                       nil,
					Decimals:                     0,
					Metadata:                     []byte("m00"),
					URI:                          []byte("u00"),
					MaxSupply:                    uint64(0),
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
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

		ginkgo.By("create asset with too long of metadata", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         []byte("n00"),
					Symbol:                       []byte("s00"),
					Decimals:                     0,
					Metadata:                     make([]byte, actions.MaxMetadataSize*2),
					URI:                          []byte("u00"),
					MaxSupply:                    uint64(0),
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
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

		ginkgo.By("create a new non-fungible asset (decimals is greater than 0)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetNonFungibleTokenID,
					Name:                         []byte("n00"),
					Symbol:                       []byte("s00"),
					Decimals:                     1,
					Metadata:                     []byte("m00"),
					URI:                          []byte("u00"),
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
			result := results[0]
			require.False(result.Success)
			require.Contains(string(result.Error), "decimal is invalid")
		})

		ginkgo.By("create a new fungible asset (simple metadata)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         asset1,
					Symbol:                       asset1Symbol,
					Decimals:                     asset1Decimals,
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
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)

			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1Symbol)
			require.Equal(decimals, asset1Decimals)
			require.Equal([]byte(metadata), asset1)
			require.Equal([]byte(uri), asset1)
			require.Zero(totalSupply)
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("create a new non-fungible asset (simple metadata)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetNonFungibleTokenID,
					Name:                         asset2,
					Symbol:                       asset2Symbol,
					Decimals:                     0,
					Metadata:                     asset2,
					URI:                          asset2,
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

			asset2ID = chain.CreateActionID(tx.ID(), 0)
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)

			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
			require.Equal([]byte(name), asset2)
			require.Equal([]byte(symbol), asset2Symbol)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset2)
			require.Equal([]byte(uri), asset2)
			require.Zero(totalSupply)
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("create a new dataset asset (simple metadata)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetDatasetTokenID,
					Name:                         asset3,
					Symbol:                       asset3Symbol,
					Decimals:                     0,
					Metadata:                     asset3,
					URI:                          asset3,
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

			asset3ID = chain.CreateActionID(tx.ID(), 0)
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset3ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))
			nftID := nchain.GenerateIDWithIndex(asset3ID, 0)
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender2, asset3ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset3ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
			require.Equal([]byte(name), asset3)
			require.Equal([]byte(symbol), asset3Symbol)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset3)
			require.Equal([]byte(uri), asset3)
			require.Equal(totalSupply, uint64(1))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)

			exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(collectionID, asset3ID.String())
			require.Equal(uniqueID, uint64(0))
			require.Equal([]byte(uri), asset3)
			require.Equal([]byte(metadata), asset3)
			require.Equal(owner, sender)
		})
	})

	ginkgo.It("update an asset", func() {
		ginkgo.By("update an asset that doesn't exist", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.UpdateAsset{
					Asset:    ids.GenerateTestID(),
					Name:     asset4,
					Symbol:   asset4Symbol,
					Metadata: asset4,
					URI:      asset4,
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
			require.Contains(string(result.Error), "asset not found")
		})

		ginkgo.By("update an asset(no field is updated)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.UpdateAsset{
					Asset: asset1ID,
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

		ginkgo.By("update an existing asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.UpdateAsset{
					Asset:     asset3ID,
					MaxSupply: uint64(100000),
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset3ID.String(), false)

			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
			require.Equal([]byte(name), asset3)
			require.Equal([]byte(symbol), asset3Symbol)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset3)
			require.Equal([]byte(uri), asset3)
			require.Equal(totalSupply, uint64(1))
			require.Equal(maxSupply, uint64(100000))
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})
	})

	ginkgo.It("mint an asset", func() {
		ginkgo.By("mint a new fungible asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetFT{
					To:    rsender2,
					Asset: asset1ID,
					Value: 15,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(15))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1Symbol)
			require.Equal(decimals, asset1Decimals)
			require.Equal([]byte(metadata), asset1)
			require.Equal([]byte(uri), asset1)
			require.Equal(totalSupply, uint64(15))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("mint a new non-fungible asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetNFT{
					To:       rsender2,
					Asset:    asset2ID,
					UniqueID: 0,
					URI:      []byte("uri"),
					Metadata: []byte("metadata"),
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset2ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))
			nftID := nchain.GenerateIDWithIndex(asset2ID, 0)
			balance, err = instances[0].ncli.Balance(context.TODO(), sender2, nftID.String())
			require.NoError(err)
			require.Equal(balance, uint64(1))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
			require.Equal([]byte(name), asset2)
			require.Equal([]byte(symbol), asset2Symbol)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset2)
			require.Equal([]byte(uri), asset2)
			require.Equal(totalSupply, uint64(1))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)

			exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(collectionID, asset2ID.String())
			require.Equal(uniqueID, uint64(0))
			require.Equal([]byte(uri), []byte("uri"))
			require.Equal([]byte(metadata), []byte("metadata"))
			require.Equal(owner, sender2)
		})

		ginkgo.By("mint fungible asset from wrong owner", func() {
			other, err := ed25519.GeneratePrivateKey()
			require.NoError(err)
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetFT{
					To:    auth.NewED25519Address(other.PublicKey()),
					Asset: asset1ID,
					Value: 10,
				}},
				factory2,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			result := results[0]

			require.False(result.Success)
			require.Contains(string(result.Error), "wrong mint actor")

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1Symbol)
			require.Equal(decimals, asset1Decimals)
			require.Equal([]byte(metadata), asset1)
			require.Equal([]byte(uri), asset1)
			require.Equal(totalSupply, uint64(15))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("mint non-fungible asset from wrong owner", func() {
			other, err := ed25519.GeneratePrivateKey()
			require.NoError(err)
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetNFT{
					To:       auth.NewED25519Address(other.PublicKey()),
					Asset:    asset2ID,
					UniqueID: 1,
					URI:      []byte("uri"),
					Metadata: []byte("metadata"),
				}},
				factory2,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			result := results[0]

			require.False(result.Success)
			require.Contains(string(result.Error), "wrong mint actor")

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
			require.Equal([]byte(name), asset2)
			require.Equal([]byte(symbol), asset2Symbol)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset2)
			require.Equal([]byte(uri), asset2)
			require.Equal(totalSupply, uint64(1))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("rejects empty mint", func() {
			other, err := ed25519.GeneratePrivateKey()
			require.NoError(err)
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				[]chain.Action{&actions.MintAssetFT{
					To:    auth.NewED25519Address(other.PublicKey()),
					Asset: asset1ID,
				}},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// bad codec)
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
			require.ErrorContains(err, "Uint64 field is not populated")
		})

		ginkgo.By("reject max mint", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetFT{
					To:    rsender2,
					Asset: asset1ID,
					Value: consts.MaxUint64,
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
			require.Contains(string(result.Error), "overflow")
		})

		ginkgo.By("rejects mint of native token", func() {
			other, err := ed25519.GeneratePrivateKey()
			require.NoError(err)
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				[]chain.Action{&actions.MintAssetFT{
					To:    auth.NewED25519Address(other.PublicKey()),
					Value: 10,
				}},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// bad codec)
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
			require.ErrorContains(err, "ID field is not populated")
		})

		ginkgo.By("mints another new asset (to self) on another account", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         asset3,
					Symbol:                       asset3Symbol,
					Decimals:                     asset3Decimals,
					Metadata:                     asset3,
					URI:                          asset3,
					MaxSupply:                    0,
					MintActor:                    rsender2,
					PauseUnpauseActor:            rsender2,
					FreezeUnfreezeActor:          rsender2,
					EnableDisableKYCAccountActor: rsender2,
				}},
				factory2,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)
			asset3ID = chain.CreateActionID(tx.ID(), 0)

			submit, _, _, err = instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.MintAssetFT{
					To:    rsender2,
					Asset: asset3ID,
					Value: 10,
				}},
				factory2,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept = expectBlk(instances[0])
			results = accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset3ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(10))
		})
	})

	ginkgo.It("burn an asset", func() {
		ginkgo.By("burn new fungible asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.BurnAssetFT{
					Asset: asset1ID,
					Value: 5,
				}},
				factory2,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID.String())
			require.NoError(err)
			require.Equal(balance, uint64(10))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1Symbol)
			require.Equal(decimals, asset1Decimals)
			require.Equal([]byte(metadata), asset1)
			require.Equal([]byte(uri), asset1)
			require.Equal(totalSupply, uint64(10))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("burn new non-fungible asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			nftID := nchain.GenerateIDWithIndex(asset2ID, 0)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.BurnAssetNFT{
					Asset: asset2ID,
					NftID: nftID,
				}},
				factory2,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[0])
			results := accept(false)
			require.Len(results, 1)
			require.True(results[0].Success)

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset2ID.String())
			require.NoError(err)
			require.Zero(balance)
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
			require.NoError(err)
			require.Zero(balance)

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
			require.Equal([]byte(name), asset2)
			require.Equal([]byte(symbol), asset2Symbol)
			require.Equal(decimals, uint8(0))
			require.Equal([]byte(metadata), asset2)
			require.Equal([]byte(uri), asset2)
			require.Equal(totalSupply, uint64(0))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})

		ginkgo.By("burn missing asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.BurnAssetFT{
					Asset: asset1ID,
					Value: 10,
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
			require.Contains(string(result.Error), "invalid balance")

			exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
			require.NoError(err)
			require.True(exists)
			require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
			require.Equal([]byte(name), asset1)
			require.Equal([]byte(symbol), asset1Symbol)
			require.Equal(decimals, asset1Decimals)
			require.Equal([]byte(metadata), asset1)
			require.Equal([]byte(uri), asset1)
			require.Equal(totalSupply, uint64(10))
			require.Zero(maxSupply)
			require.Equal(admin, sender)
			require.Equal(mintActor, sender)
			require.Equal(pauseUnpauseActor, sender)
			require.Equal(freezeUnfreezeActor, sender)
			require.Equal(enableDisableKYCAccountActor, sender)
		})
	})

	ginkgo.It("create and mint multiple assets", func() {
		ginkgo.By("create and mint multiple of fungible assets in a single tx", func() {
			// Create asset
			parser, err := instances[3].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[3].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{
					&actions.CreateAsset{
						AssetType:                    nconsts.AssetFungibleTokenID,
						Name:                         asset1,
						Symbol:                       asset1Symbol,
						Decimals:                     asset1Decimals,
						Metadata:                     asset1,
						URI:                          asset1,
						MaxSupply:                    0,
						MintActor:                    rsender,
						PauseUnpauseActor:            rsender,
						FreezeUnfreezeActor:          rsender,
						EnableDisableKYCAccountActor: rsender,
					},
					&actions.CreateAsset{
						AssetType:                    nconsts.AssetFungibleTokenID,
						Name:                         asset2,
						Symbol:                       asset2Symbol,
						Decimals:                     asset2Decimals,
						Metadata:                     asset2,
						URI:                          asset2,
						MaxSupply:                    0,
						MintActor:                    rsender,
						PauseUnpauseActor:            rsender,
						FreezeUnfreezeActor:          rsender,
						EnableDisableKYCAccountActor: rsender,
					},
				},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[3])
			results := accept(true)
			require.Len(results, 1)
			require.True(results[0].Success)

			asset1ID = chain.CreateActionID(tx.ID(), 0)
			asset2ID = chain.CreateActionID(tx.ID(), 1)

			// Mint multiple
			submit, _, _, err = instances[3].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{
					&actions.MintAssetFT{
						To:    rsender2,
						Asset: asset1ID,
						Value: 10,
					},
					&actions.MintAssetFT{
						To:    rsender2,
						Asset: asset2ID,
						Value: 10,
					},
				},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept = expectBlk(instances[3])
			results = accept(true)
			require.Len(results, 1)
			require.True(results[0].Success)

			// check sender2 assets
			balance1, err := instances[3].ncli.Balance(context.TODO(), sender2, asset1ID.String())
			require.NoError(err)
			require.Equal(balance1, uint64(10))

			balance2, err := instances[3].ncli.Balance(context.TODO(), sender2, asset2ID.String())
			require.NoError(err)
			require.Equal(balance2, uint64(10))
		})

		ginkgo.By("create and mint multiple of non-fungible assets in a single tx", func() {
			// Create asset
			parser, err := instances[3].ncli.Parser(context.Background())
			require.NoError(err)
			submit, tx, _, err := instances[3].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{
					&actions.CreateAsset{
						AssetType:                    nconsts.AssetNonFungibleTokenID,
						Name:                         asset1,
						Symbol:                       asset1Symbol,
						Decimals:                     0,
						Metadata:                     asset1,
						URI:                          asset1,
						MaxSupply:                    0,
						MintActor:                    rsender,
						PauseUnpauseActor:            rsender,
						FreezeUnfreezeActor:          rsender,
						EnableDisableKYCAccountActor: rsender,
					},
					&actions.CreateAsset{
						AssetType:                    nconsts.AssetNonFungibleTokenID,
						Name:                         asset2,
						Symbol:                       asset2Symbol,
						Decimals:                     0,
						Metadata:                     asset2,
						URI:                          asset2,
						MaxSupply:                    0,
						MintActor:                    rsender,
						PauseUnpauseActor:            rsender,
						FreezeUnfreezeActor:          rsender,
						EnableDisableKYCAccountActor: rsender,
					},
				},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept := expectBlk(instances[3])
			results := accept(true)
			require.Len(results, 1)
			require.True(results[0].Success)

			asset1ID = chain.CreateActionID(tx.ID(), 0)
			asset2ID = chain.CreateActionID(tx.ID(), 1)

			// Mint multiple
			submit, _, _, err = instances[3].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{
					&actions.MintAssetNFT{
						To:       rsender2,
						Asset:    asset1ID,
						UniqueID: 0,
						URI:      []byte("uri1"),
						Metadata: []byte("metadata1"),
					},
					&actions.MintAssetNFT{
						To:       rsender2,
						Asset:    asset2ID,
						UniqueID: 1,
						URI:      []byte("uri2"),
						Metadata: []byte("metadata2"),
					},
				},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			accept = expectBlk(instances[3])
			results = accept(true)
			require.Len(results, 1)
			require.True(results[0].Success)

			// check sender2 assets
			balance1, err := instances[3].ncli.Balance(context.TODO(), sender2, asset1ID.String())
			require.NoError(err)
			require.Equal(balance1, uint64(1))
			/* 		nftID := nchain.GenerateID(asset2ID, 0)
			balance1, err = instances[3].ncli.Balance(context.TODO(), sender2, nftID.String())
			require.NoError(err)
			require.Equal(balance1, uint64(1)) */

			balance2, err := instances[3].ncli.Balance(context.TODO(), sender2, asset2ID.String())
			require.NoError(err)
			require.Equal(balance2, uint64(1))
			/* 	nftID = nchain.GenerateID(asset2ID, 1)
			balance2, err = instances[0].ncli.Balance(context.TODO(), sender2, nftID.String())
			require.NoError(err)
			require.Equal(balance2, uint64(1)) */
		})
	})
})
