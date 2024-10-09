// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
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
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Add assettype to token
		assetType, err := prompt.Choice("assetType(0 for fungible, 1 for non-fungible and 2 for fractional)", 3)
		if err != nil {
			return err
		}
		if assetType < 0 || assetType > 2 {
			utils.Outf("{{red}}assetType:%s does not exist{{/}}\n", assetType)
			utils.Outf("{{red}}fungible=0 non-fungible=1 dataset=2{{/}}\n")
			utils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}

		// Add name to token
		name, err := prompt.String("name", 1, storage.MaxAssetMetadataSize)
		if err != nil {
			return err
		}

		// Add symbol to token
		symbol, err := prompt.String("symbol", 1, storage.MaxTextSize)
		if err != nil {
			return err
		}

		// Add decimal to token
		decimals, err := prompt.Choice("decimals", storage.MaxAssetDecimals+1)
		if err != nil {
			return err
		}

		// Add metadata to token
		metadata, err := prompt.String("metadata", 1, storage.MaxAssetMetadataSize)
		if err != nil {
			return err
		}

		// Add owner
		owner := priv.Address

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		result, _, err := sendAndWait(ctx, []chain.Action{&actions.CreateAsset{
			AssetType:                    uint8(assetType),
			Name:                         name,
			Symbol:                       symbol,
			Decimals:                     uint8(decimals), // already constrain above to prevent overflow
			Metadata:                     metadata,
			MaxSupply:                    uint64(0),
			MintAdmin:                    owner,
			PauseUnpauseAdmin:            owner,
			FreezeUnfreezeAdmin:          owner,
			EnableDisableKYCAccountAdmin: owner,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var updateAssetCmd = &cobra.Command{
	Use: "update",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select asset ID to update
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}

		// Add name to token
		name, err := prompt.String("name", 1, storage.MaxNameSize)
		if err != nil {
			return err
		}

		// Add symbol to token
		symbol, err := prompt.String("symbol", 1, storage.MaxSymbolSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.UpdateAsset{
			AssetAddress: assetAddress,
			Name:         name,
			Symbol:       symbol,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var mintAssetFTCmd = &cobra.Command{
	Use: "mint-ft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to mint
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}
		assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := ncli.Asset(ctx, assetAddress.String(), false)
		if err != nil {
			return err
		}
		if mintAdmin != priv.Address.String() {
			utils.Outf("{{red}}%s has permission to mint asset '%s' with assetID '%s', you are not{{/}}\n", mintAdmin, name, assetAddress)
			utils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		utils.Outf(
			"{{blue}}assetType:{{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			owner,
			mintAdmin,
			pauseUnpauseAdmin,
			freezeUnfreezeAdmin,
			enableDisableKYCAccountAdmin,
		)

		// Select recipient
		recipient, err := prompt.Address("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := parseAmount("amount", decimals, consts.MaxUint64)
		if err != nil {
			return err
		}
		fmt.Println("amount: ", amount, decimals)

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.MintAssetFT{
			AssetAddress: assetAddress,
			To:           recipient,
			Value:        amount,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var mintAssetNFTCmd = &cobra.Command{
	Use: "mint-nft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select nft collection id to mint to
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}
		assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := ncli.Asset(ctx, assetAddress.String(), false)
		if err != nil {
			return err
		}
		if mintAdmin != priv.Address.String() {
			utils.Outf("{{red}}%s has permission to mint asset '%s' with assetID '%s', you are not{{/}}\n", mintAdmin, name, assetAddress)
			utils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		utils.Outf(
			"{{blue}}assetType:{{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			owner,
			mintAdmin,
			pauseUnpauseAdmin,
			freezeUnfreezeAdmin,
			enableDisableKYCAccountAdmin,
		)

		// Select recipient
		recipient, err := prompt.Address("recipient")
		if err != nil {
			return err
		}

		// Add metadata for the NFT
		metadataNFT, err := prompt.String("metadata", 1, storage.MaxAssetMetadataSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.MintAssetNFT{
			AssetAddress: assetAddress,
			To:           recipient,
			Metadata:     metadataNFT,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var burnAssetFTCmd = &cobra.Command{
	Use: "burn-ft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to burn
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}
		assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := ncli.Asset(ctx, assetAddress.String(), false)
		if err != nil {
			return err
		}
		utils.Outf(
			"{{blue}}assetType:{{/}} %s name:{{/}} %s {{blue}}symbol:{{/}} %s {{blue}}decimals:{{/}} %d {{blue}}metadata:{{/}} %s {{blue}}uri:{{/}} %s {{blue}}totalSupply:{{/}} %d {{blue}}maxSupply:{{/}} %d {{blue}}admin:{{/}} %s {{blue}}mintActor:{{/}} %s {{blue}}pauseUnpauseActor:{{/}} %s {{blue}}freezeUnfreezeActor:{{/}} %s {{blue}}enableDisableKYCAccountActor:{{/}} %s\n",
			assetType,
			name,
			symbol,
			decimals,
			metadata,
			uri,
			totalSupply,
			maxSupply,
			owner,
			mintAdmin,
			pauseUnpauseAdmin,
			freezeUnfreezeAdmin,
			enableDisableKYCAccountAdmin,
		)

		// Select amount
		amount, err := parseAmount("amount", decimals, consts.MaxUint64)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.BurnAssetFT{
			AssetAddress: assetAddress,
			Value:        amount,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var burnAssetNFTCmd = &cobra.Command{
	Use: "burn-nft",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select asset ID to burn
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}

		// Select nft ID to burn
		nftAddress, err := prompt.Address("nftAddress")
		if err != nil {
			return err
		}

		if _, _, _, _, _, _, _, err = handler.GetAssetNFTInfo(context.TODO(), ncli, priv.Address, nftAddress, true); err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.BurnAssetNFT{
			AssetAddress:    assetAddress,
			AssetNftAddress: nftAddress,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
