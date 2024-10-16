// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"

	hutils "github.com/ava-labs/hypersdk/utils"
)

var datasetCmd = &cobra.Command{
	Use: "dataset",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var createDatasetFromExistingAssetCmd = &cobra.Command{
	Use: "create",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select asset
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}

		// Add name to dataset
		name, err := prompt.String("name", 1, storage.MaxNameSize)
		if err != nil {
			return err
		}

		// Add description to dataset
		description, err := prompt.String("description", 1, storage.MaxTextSize)
		if err != nil {
			return err
		}

		// Prompt for isCommunityDataset
		isCommunityDataset, err := prompt.Bool("isCommunityDataset")
		if err != nil {
			return err
		}

		// Add metadata to dataset
		metadata, err := prompt.String("metadata", 1, storage.MaxDatasetMetadataSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.CreateDataset{
			AssetAddress:       assetAddress,
			Name:               name,
			Description:        description,
			Categories:         name,
			LicenseName:        "MIT",
			LicenseSymbol:      "MIT",
			LicenseURL:         "https://opensource.org/licenses/MIT",
			Metadata:           metadata,
			IsCommunityDataset: isCommunityDataset,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var updateDatasetCmd = &cobra.Command{
	Use: "update",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID to update
		datasetAddress, err := prompt.Address("datasetAddress")
		if err != nil {
			return err
		}

		// Update name to dataset
		name, err := prompt.String("name", 1, storage.MaxNameSize)
		if err != nil {
			return err
		}

		// Prompt for isCommunityDataset
		isCommunityDataset, err := prompt.Bool("isCommunityDataset")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.UpdateDataset{
			DatasetAddress:     datasetAddress,
			Name:               name,
			IsCommunityDataset: isCommunityDataset,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var getDatasetCmd = &cobra.Command{
	Use: "info",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, _, _, _, _, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Select dataset to look up
		datasetAddress, err := prompt.Address("datasetAddress")
		if err != nil {
			return err
		}

		// Get dataset info
		hutils.Outf("Retrieving dataset info for datasetID: %s\n", datasetAddress)
		_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetDatasetInfo(ctx, ncli, datasetAddress)
		if err != nil {
			return err
		}

		// Get asset info
		hutils.Outf("Retrieving asset info for assetID: %s\n", datasetAddress)
		_, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetAssetInfo(ctx, ncli, priv.Address, datasetAddress, true, true, -1)
		return err
	},
}

var initiateContributeDatasetCmd = &cobra.Command{
	Use: "initiate-contribute",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset to contribute to
		datasetAddress, err := prompt.Address("datasetAddress")
		if err != nil {
			return err
		}

		// Add data identifier to dataset
		dataIdentifier, err := prompt.String("dataIdentifier", 1, storage.MaxAssetMetadataSize-storage.MaxDatasetDataLocationSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.InitiateContributeDataset{
			DatasetAddress: datasetAddress,
			DataLocation:   "default",
			DataIdentifier: dataIdentifier,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var getDataContributionPendingCmd = &cobra.Command{
	Use: "contribute-info",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Select dataset to look up
		contributionID, err := prompt.ID("contributionID")
		if err != nil {
			return err
		}

		// Get pending data contributions info
		hutils.Outf("Retrieving pending data contributions info for datasetID: %s\n", contributionID)
		_, _, _, _, _, err = handler.GetDataContributionInfo(ctx, ncli, contributionID)
		if err != nil {
			return err
		}
		return nil
	},
}

var completeContributeDatasetCmd = &cobra.Command{
	Use: "complete-contribute",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select dataset ID
		datasetAddress, err := prompt.Address("datasetAddress")
		if err != nil {
			return err
		}

		// Select contribution ID
		contributionID, err := prompt.ID("contributionID")
		if err != nil {
			return err
		}

		// Select the contributor
		contributor, err := prompt.Address("contributor")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.CompleteContributeDataset{
			DatasetContributionID: contributionID,
			DatasetAddress:        datasetAddress,
			DatasetContributor:    contributor,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
