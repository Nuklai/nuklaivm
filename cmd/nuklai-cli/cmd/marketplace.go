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

	hutils "github.com/ava-labs/hypersdk/utils"
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

		// Select dataset
		datasetAddress, err := prompt.Address("datasetAddress")
		if err != nil {
			return err
		}

		// Select paymentAssetAddress
		paymentAssetAddress, err := parseAsset("paymentAssetAddress")
		if err != nil {
			return err
		}

		balance, _, _, _, decimals, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, paymentAssetAddress, true, true, -1)
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

		result, _, err := sendAndWait(ctx, []chain.Action{&actions.PublishDatasetMarketplace{
			DatasetAddress:       datasetAddress,
			PaymentAssetAddress:  paymentAssetAddress,
			DatasetPricePerBlock: priceAmountPerBlock,
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
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select marketplaceAddress
		marketplaceAddress, err := prompt.Address("marketplaceAddress")
		if err != nil {
			return err
		}

		// Select paymentAssetAddress
		paymentAssetAddress, err := parseAsset("paymentAssetAddress")
		if err != nil {
			return err
		}

		// Get numBlocksToSubscribe
		numBlocksToSubscribe, err := prompt.Int("numBlocksToSubscribe", consts.MaxInt)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.SubscribeDatasetMarketplace{
			MarketplaceAssetAddress: marketplaceAddress,
			PaymentAssetAddress:     paymentAssetAddress,
			NumBlocksToSubscribe:    uint64(numBlocksToSubscribe),
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

		// Select datasetAddress
		datasetAddress, err := prompt.Address("datasetAddress")
		if err != nil {
			return err
		}
		// Get dataset info from the marketplace
		hutils.Outf("Retrieving dataset info from the marketplace: %s\n", datasetAddress)
		_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetDatasetInfoFromMarketplace(ctx, ncli, datasetAddress)
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

		marketplaceAddress, err := prompt.Address("marketplaceAddress")
		if err != nil {
			return err
		}

		// Select paymentAssetAddress
		paymentAssetAddress, err := parseAsset("paymentAssetAddress")
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
			MarketplaceAssetAddress: marketplaceAddress,
			PaymentAssetAddress:     paymentAssetAddress,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
