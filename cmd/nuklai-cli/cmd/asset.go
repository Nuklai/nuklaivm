// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"

	nutils "github.com/nuklai/nuklaivm/utils"
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
		assetType, err := prompt.Choice("assetType(0 for fungible, 1 for non-fungible and 2 for dataset)", 3)
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
		name, err := prompt.String("name", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Add symbol to token
		symbol, err := prompt.String("symbol", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}

		// Add decimal to token
		decimals, err := prompt.Choice("decimals", actions.MaxDecimals+1)
		if err != nil {
			return err
		}

		// Add metadata to token
		metadata, err := prompt.String("metadata", 1, actions.MaxMetadataSize)
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

		// Generate transaction
		assetID, err := nutils.GenerateRandomID()
		if err != nil {
			return err
		}
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.CreateAsset{
			AssetID:                      assetID,
			AssetType:                    uint8(assetType),
			Name:                         []byte(name),
			Symbol:                       []byte(symbol),
			Decimals:                     uint8(decimals), // already constrain above to prevent overflow
			Metadata:                     []byte(metadata),
			URI:                          []byte("https://nukl.ai"),
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
		assetID, err := prompt.ID("assetID")
		if err != nil {
			return err
		}

		// Add name to token
		name, err := prompt.String("name", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Add symbol to token
		symbol, err := prompt.String("symbol", 1, actions.MaxTextSize)
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
			AssetID: assetID,
			Name:    []byte(name),
			Symbol:  []byte(symbol),
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
		assetID, err := prompt.ID("assetID")
		if err != nil {
			return err
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := ncli.Asset(ctx, assetID.String(), false)
		if err != nil {
			return err
		}
		if !exists {
			utils.Outf("{{red}}assetID:%s does not exist{{/}}\n", assetID)
			utils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		if mintAdmin != priv.Address.String() {
			utils.Outf("{{red}}%s has permission to mint asset '%s' with assetID '%s', you are not{{/}}\n", mintAdmin, name, assetID)
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

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.MintAssetFT{
			AssetID: assetID,
			To:      recipient,
			Value:   amount,
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
		assetID, err := prompt.ID("assetID")
		if err != nil {
			return err
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := ncli.Asset(ctx, assetID.String(), false)
		if err != nil {
			return err
		}
		if !exists {
			utils.Outf("{{red}}name: %s with assetID:%s does not exist{{/}}\n", name, assetID)
			utils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		if mintAdmin != priv.Address.String() {
			utils.Outf("{{red}}%s has permission to mint asset '%s' with assetID '%s', you are not{{/}}\n", mintAdmin, name, assetID)
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

		// Choose unique id for the NFT
		uniqueID, err := prompt.Int("unique nft #", consts.MaxInt)
		if err != nil {
			return err
		}

		// Add metadata for the NFT
		metadataNFT, err := prompt.String("metadata", 1, actions.MaxMetadataSize)
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
			AssetID:  assetID,
			UniqueID: uint64(uniqueID),
			To:       recipient,
			URI:      []byte(metadataNFT),
			Metadata: []byte(metadataNFT),
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
		assetID, err := prompt.ID("assetID")
		if err != nil {
			return err
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := ncli.Asset(ctx, assetID.String(), false)
		if err != nil {
			return err
		}
		if !exists {
			utils.Outf("{{red}}assetID:%s does not exist{{/}}\n", assetID)
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
			AssetID: assetID,
			Value:   amount,
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
		assetID, err := prompt.ID("assetID")
		if err != nil {
			return err
		}

		// Select nft ID to burn
		nftID, err := prompt.ID("nftID")
		if err != nil {
			return err
		}

		if _, _, _, _, _, _, err = handler.GetAssetNFTInfo(context.TODO(), ncli, priv.Address, nftID, true); err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.BurnAssetNFT{
			AssetID: assetID,
			NftID:   nftID,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
