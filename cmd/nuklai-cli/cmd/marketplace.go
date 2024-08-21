// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
)

var marketPlaceCmd = &cobra.Command{
	Use: "mp",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var createDatasetCmd = &cobra.Command{
	Use: "create-dataset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Add name to dataset
		name, err := handler.Root().PromptString("name", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}

		// Add description to dataset
		description, err := handler.Root().PromptString("description", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Prompt for isCommunityDataset
		isCommunityDataset, err := handler.Root().PromptBool("isCommunityDataset")
		if err != nil {
			return err
		}

		// Add metadata to dataset
		metadata, err := handler.Root().PromptString("metadata", 1, actions.MaxDatasetMetadataSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		txID, err := sendAndWait(ctx, []chain.Action{&actions.CreateDataset{
			Name:               []byte(name),
			Description:        []byte(description),
			Categories:         []byte(name),
			LicenseName:        []byte("MIT"),
			LicenseSymbol:      []byte("MIT"),
			LicenseURL:         []byte("https://opensource.org/licenses/MIT"),
			Metadata:           []byte(metadata),
			IsCommunityDataset: isCommunityDataset,
		}}, cli, scli, tcli, factory, true)
		if err != nil {
			return err
		}

		// Print datasetID/assetID
		datasetID := chain.CreateActionID(txID, 0)
		hutils.Outf("{{green}}datasetID:{{/}} %s\n", datasetID)
		hutils.Outf("{{green}}assetID:{{/}} %s\n", datasetID)

		return nil
	},
}

var getDatasetCmd = &cobra.Command{
	Use: "get-dataset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Select dataset to look up
		datasetID, err := handler.Root().PromptAsset("datasetID", false)
		if err != nil {
			return err
		}

		// Get dataset info
		_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetDatasetInfo(ctx, ncli, datasetID)
		if err != nil {
			return err
		}

		return nil
	},
}
