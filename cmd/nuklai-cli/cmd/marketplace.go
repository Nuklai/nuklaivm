// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/consts"

	hutils "github.com/ava-labs/hypersdk/utils"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nutils "github.com/nuklai/nuklaivm/utils"
)

var marketplaceCmd = &cobra.Command{
	Use: "marketplace",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var publishDatasetMarketplaceCmd = &cobra.Command{
	Use: "publish",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Select assetForPayment ID
		assetForPayment, err := prompt.Asset("assetForPayment", nconsts.Symbol, true)
		if err != nil {
			return err
		}

		balance, _, _, _, decimals, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetForPayment, true)
		if balance == 0 || err != nil {
			return err
		}

		// Get priceAmountPerBlock
		priceAmountPerBlock, err := parseAmount("priceAmountPerBlock", decimals, balance)
		if err != nil {
			return err
		}

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
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.PublishDatasetMarketplace{
			MarketplaceAssetID: assetID,
			DatasetID:          datasetID,
			BaseAssetID:        assetForPayment,
			BasePrice:          priceAmountPerBlock,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var subscribeDatasetMarketplaceCmd = &cobra.Command{
	Use: "subscribe",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Get dataset info
		hutils.Outf("Retrieving dataset info for datasetID: %s\n", datasetID)
		_, _, _, _, _, _, _, _, saleID, baseAsset, basePrice, _, _, _, _, _, err := handler.GetDatasetInfo(ctx, ncli, datasetID)
		if err != nil {
			return err
		}
		marketplaceID, err := ids.FromString(saleID)
		if err != nil {
			return err
		}

		// Select assetForPayment ID
		assetForPayment, err := prompt.Asset("assetForPayment", nconsts.Symbol, true)
		if err != nil {
			return err
		}
		if !strings.EqualFold(assetForPayment.String(), baseAsset) {
			return fmt.Errorf("assetForPayment must be the same as the dataset's baseAsset. BaseAsset: %s", baseAsset)
		}

		// Get numBlocksToSubscribe
		numBlocksToSubscribe, err := prompt.Int("numBlocksToSubscribe", consts.MaxInt)
		if err != nil {
			return err
		}

		// Ensure user has enough balance
		balance, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetForPayment, true)
		if err != nil {
			return err
		}
		if balance < basePrice*uint64(numBlocksToSubscribe) {
			return fmt.Errorf("insufficient balance. Required: %d", basePrice*uint64(numBlocksToSubscribe))
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.SubscribeDatasetMarketplace{
			DatasetID:            datasetID,
			MarketplaceAssetID:   marketplaceID,
			AssetForPayment:      assetForPayment,
			NumBlocksToSubscribe: uint64(numBlocksToSubscribe),
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var infoDatasetMarketplaceCmd = &cobra.Command{
	Use: "info",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Select dataset ID
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}
		// Get dataset info from the marketplace
		hutils.Outf("Retrieving dataset info from the marketplace: %s\n", datasetID)
		_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetDatasetInfoFromMarketplace(ctx, ncli, datasetID)
		if err != nil {
			return err
		}
		return nil
	},
}

var claimPaymentMarketplaceCmd = &cobra.Command{
	Use: "claim-payment",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Get dataset info
		// Get dataset info from the marketplace
		hutils.Outf("Retrieving dataset info from the marketplace: %s\n", datasetID)
		_, _, _, saleID, baseAsset, _, _, _, _, _, _, _, _, _, _, err := handler.GetDatasetInfoFromMarketplace(ctx, ncli, datasetID)
		if err != nil {
			return err
		}
		marketplaceID, err := ids.FromString(saleID)
		if err != nil {
			return err
		}

		// Select assetForPayment ID
		assetForPayment, err := ids.FromString(baseAsset)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ClaimMarketplacePayment{
			DatasetID:          datasetID,
			MarketplaceAssetID: marketplaceID,
			AssetForPayment:    assetForPayment,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
