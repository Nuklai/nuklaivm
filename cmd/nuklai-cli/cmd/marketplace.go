// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/consts"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	nchain "github.com/nuklai/nuklaivm/chain"
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
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetIDStr, err := handler.Root().PromptString("datasetID", 1, consts.MaxInt)
		if err != nil {
			return err
		}
		datasetID, err := ids.FromString(datasetIDStr)
		if err != nil {
			return err
		}

		// Select assetForPayment ID
		assetForPayment, err := handler.Root().PromptAsset("assetForPayment", true)
		if err != nil {
			return err
		}

		balance, _, _, _, decimals, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetForPayment, true)
		if balance == 0 || err != nil {
			return err
		}

		// Get priceAmountPerBlock
		priceAmountPerBlock, err := handler.Root().PromptAmount("priceAmountPerBlock", decimals, balance, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		txID, err := sendAndWait(ctx, []chain.Action{&actions.PublishDatasetMarketplace{
			Dataset:   datasetID,
			BaseAsset: assetForPayment,
			BasePrice: priceAmountPerBlock,
		}}, hcli, hws, ncli, factory, true)
		if err != nil {
			return err
		}

		// Print assetID for the marketplace token
		assetID := chain.CreateActionID(txID, 0)
		hutils.Outf("{{green}}assetID:{{/}} %s\n", assetID)

		return nil
	},
}

var subscribeDatasetMarketplaceCmd = &cobra.Command{
	Use: "subscribe",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetIDStr, err := handler.Root().PromptString("datasetID", 1, consts.MaxInt)
		if err != nil {
			return err
		}
		datasetID, err := ids.FromString(datasetIDStr)
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
		assetForPayment, err := handler.Root().PromptAsset("assetForPayment", true)
		if err != nil {
			return err
		}
		if !strings.EqualFold(assetForPayment.String(), baseAsset) {
			return fmt.Errorf("assetForPayment must be the same as the dataset's baseAsset. BaseAsset: %s", baseAsset)
		}

		// Get numBlocksToSubscribe
		numBlocksToSubscribeStr, err := handler.Root().PromptString("numBlocksToSubscribe", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}
		numBlocksToSubscribe, err := strconv.ParseUint(numBlocksToSubscribeStr, 10, 64)
		if err != nil {
			return err
		}

		// Ensure user has enough balance
		balance, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetForPayment, true)
		if err != nil {
			return err
		}
		if balance < basePrice*numBlocksToSubscribe {
			return fmt.Errorf("insufficient balance. Required: %d", basePrice*numBlocksToSubscribe)
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.SubscribeDatasetMarketplace{
			Dataset:              datasetID,
			MarketplaceID:        marketplaceID,
			AssetForPayment:      assetForPayment,
			NumBlocksToSubscribe: numBlocksToSubscribe,
		}}, hcli, hws, ncli, factory, true)
		if err != nil {
			return err
		}

		// Print nftID that the user received for the subscription
		nftID := nchain.GenerateIDWithAddress(marketplaceID, priv.Address)
		hutils.Outf("{{green}}nftID:{{/}} %s\n", nftID)

		return nil
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
		datasetIDStr, err := handler.Root().PromptString("datasetID", 1, consts.MaxInt)
		if err != nil {
			return err
		}
		datasetID, err := ids.FromString(datasetIDStr)
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
		_, _, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetIDStr, err := handler.Root().PromptString("datasetID", 1, consts.MaxInt)
		if err != nil {
			return err
		}
		datasetID, err := ids.FromString(datasetIDStr)
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
		assetForPayment, err := ids.FromString((baseAsset))
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.ClaimMarketplacePayment{
			Dataset:         datasetID,
			MarketplaceID:   marketplaceID,
			AssetForPayment: assetForPayment,
		}}, hcli, hws, ncli, factory, true)
		if err != nil {
			return err
		}

		return nil
	},
}
