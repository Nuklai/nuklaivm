// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

var assetCmd = &cobra.Command{
	Use: "asset",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var createAssetCmd = &cobra.Command{
	Use: "create",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Add assettype to token
		assetType, err := handler.Root().PromptChoice("assetType(0 for fungible, 1 for non-fungible and 2 for dataset)", 3)
		if err != nil {
			return err
		}
		if assetType < 0 || assetType > 2 {
			hutils.Outf("{{red}}assetType:%s does not exist{{/}}\n", assetType)
			hutils.Outf("{{red}}fungible=0 non-fungible=1 dataset=2{{/}}\n")
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}

		// Add name to token
		name, err := handler.Root().PromptString("name", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Add symbol to token
		symbol, err := handler.Root().PromptString("symbol", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}

		// Add decimal to token
		decimals, err := handler.Root().PromptChoice("decimals", actions.MaxDecimals+1)
		if err != nil {
			return err
		}

		// Add metadata to token
		metadata, err := handler.Root().PromptString("metadata", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Add owner
		owner := priv.Address

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		txID, err := sendAndWait(ctx, []chain.Action{&actions.CreateAsset{
			AssetType:                    uint8(assetType),
			Name:                         []byte(name),
			Symbol:                       []byte(symbol),
			Decimals:                     uint8(decimals), // already constrain above to prevent overflow
			Metadata:                     []byte(metadata),
			URI:                          []byte("https://nukl.ai"),
			MaxSupply:                    uint64(0),
			MintActor:                    owner,
			PauseUnpauseActor:            owner,
			FreezeUnfreezeActor:          owner,
			EnableDisableKYCAccountActor: owner,
		}}, cli, scli, tcli, factory, true)
		if err != nil {
			return err
		}

		// Print assetID
		assetID := chain.CreateActionID(txID, 0)
		hutils.Outf("{{green}}assetID:{{/}} %s\n", assetID)

		// Print nftID if it's a dataset
		if uint8(assetType) == nconsts.AssetDatasetTokenID {
			nftID := nchain.GenerateID(assetID, 0)
			hutils.Outf("{{green}}nftID:{{/}} %s\n", nftID)
		}
		return nil
	},
}

var mintAssetFTCmd = &cobra.Command{
	Use: "mint-ft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to mint
		assetID, err := handler.Root().PromptAsset("assetID", false)
		if err != nil {
			return err
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := tcli.Asset(ctx, assetID.String(), false)
		if err != nil {
			return err
		}
		if !exists {
			hutils.Outf("{{red}}assetID:%s does not exist{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		if mintActor != codec.MustAddressBech32(nconsts.HRP, priv.Address) {
			hutils.Outf("{{red}}%s has permission to mint asset '%s' with assetID '%s', you are not{{/}}\n", mintActor, name, assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		hutils.Outf(
			"{{blue}}assetType:{{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			admin,
			mintActor,
			pauseUnpauseActor,
			freezeUnfreezeActor,
			enableDisableKYCAccountActor,
		)

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, consts.MaxUint64, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.MintAssetFT{
			Asset: assetID,
			To:    recipient,
			Value: amount,
		}}, cli, scli, tcli, factory, true)
		return err
	},
}

var mintAssetNFTCmd = &cobra.Command{
	Use: "mint-nft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select nft collection id to mint to
		assetID, err := handler.Root().PromptAsset("assetID", false)
		if err != nil {
			return err
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := tcli.Asset(ctx, assetID.String(), false)
		if err != nil {
			return err
		}
		if !exists {
			hutils.Outf("{{red}}name: %s with assetID:%s does not exist{{/}}\n", name, assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		if mintActor != codec.MustAddressBech32(nconsts.HRP, priv.Address) {
			hutils.Outf("{{red}}%s has permission to mint asset '%s' with assetID '%s', you are not{{/}}\n", mintActor, name, assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		hutils.Outf(
			"{{blue}}assetType:{{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			admin,
			mintActor,
			pauseUnpauseActor,
			freezeUnfreezeActor,
			enableDisableKYCAccountActor,
		)

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Choose unique id for the NFT
		uniqueIDStr, err := handler.Root().PromptString("unique nft #", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}
		uniqueID, err := strconv.ParseUint(uniqueIDStr, 10, 64)
		if err != nil {
			return err
		}

		// Add URI for the NFT
		uriNFT, err := handler.Root().PromptString("uri", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.MintAssetNFT{
			Asset:    assetID,
			To:       recipient,
			UniqueID: uniqueID,
			URI:      []byte(uriNFT),
		}}, cli, scli, tcli, factory, true)
		if err != nil {
			return err
		}
		// Print nftID
		nftID := nchain.GenerateID(assetID, uniqueID)
		hutils.Outf("{{green}}NFT ID:{{/}} %s\n", nftID)
		return nil
	},
}

var burnAssetFTCmd = &cobra.Command{
	Use: "burn-ft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to burn
		assetID, err := handler.Root().PromptAsset("assetID", false)
		if err != nil {
			return err
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := tcli.Asset(ctx, assetID.String(), false)
		if err != nil {
			return err
		}
		if !exists {
			hutils.Outf("{{red}}assetID:%s does not exist{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		hutils.Outf(
			"{{blue}}assetType:{{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			admin,
			mintActor,
			pauseUnpauseActor,
			freezeUnfreezeActor,
			enableDisableKYCAccountActor,
		)

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, consts.MaxUint64, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.BurnAssetFT{
			Asset: assetID,
			Value: amount,
		}}, cli, scli, tcli, factory, true)
		return err
	},
}

var burnAssetNFTCmd = &cobra.Command{
	Use: "burn-nft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select asset ID to burn
		assetID, err := handler.Root().PromptAsset("assetID", false)
		if err != nil {
			return err
		}

		// Select nft ID to burn
		nftID, err := handler.Root().PromptAsset("nftID", false)
		if err != nil {
			return err
		}

		if _, _, _, _, _, err = handler.GetAssetNFTInfo(context.TODO(), tcli, priv.Address, nftID, true); err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.BurnAssetNFT{
			Asset: assetID,
			NftID: nftID,
		}}, cli, scli, tcli, factory, true)
		if err != nil {
			return err
		}
		return nil
	},
}
