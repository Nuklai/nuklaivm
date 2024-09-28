// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/consts"

	hutils "github.com/ava-labs/hypersdk/utils"
)

var datasetCmd = &cobra.Command{
	Use: "dataset",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var createDatasetCmd = &cobra.Command{
	Use: "create",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Add name to dataset
		name, err := prompt.String("name", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Add description to dataset
		description, err := prompt.String("description", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Prompt for isCommunityDataset
		isCommunityDataset, err := prompt.Bool("isCommunityDataset")
		if err != nil {
			return err
		}

		// Add metadata to dataset
		metadata, err := prompt.String("metadata", 1, actions.MaxDatasetMetadataSize)
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
			AssetID:            ids.Empty,
			Name:               []byte(name),
			Description:        []byte(description),
			Categories:         []byte(name),
			LicenseName:        []byte("MIT"),
			LicenseSymbol:      []byte("MIT"),
			LicenseURL:         []byte("https://opensource.org/licenses/MIT"),
			Metadata:           []byte(metadata),
			IsCommunityDataset: isCommunityDataset,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var createDatasetFromExistingAssetCmd = &cobra.Command{
	Use: "create-from-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select asset ID
		assetID, err := prompt.ID("assetID")
		if err != nil {
			return err
		}

		// Add name to dataset
		name, err := prompt.String("name", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Add description to dataset
		description, err := prompt.String("description", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Prompt for isCommunityDataset
		isCommunityDataset, err := prompt.Bool("isCommunityDataset")
		if err != nil {
			return err
		}

		// Add metadata to dataset
		metadata, err := prompt.String("metadata", 1, actions.MaxDatasetMetadataSize)
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
			AssetID:            assetID,
			Name:               []byte(name),
			Description:        []byte(description),
			Categories:         []byte(name),
			LicenseName:        []byte("MIT"),
			LicenseSymbol:      []byte("MIT"),
			LicenseURL:         []byte("https://opensource.org/licenses/MIT"),
			Metadata:           []byte(metadata),
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
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Update name to dataset
		name, err := prompt.String("name", 1, actions.MaxMetadataSize)
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
			DatasetID:          datasetID,
			Name:               []byte(name),
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
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Get dataset info
		hutils.Outf("Retrieving dataset info for datasetID: %s\n", datasetID)
		_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetDatasetInfo(ctx, ncli, datasetID)
		if err != nil {
			return err
		}

		// Get asset info
		hutils.Outf("Retrieving asset info for assetID: %s\n", datasetID)
		_, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetAssetInfo(ctx, ncli, priv.Address, datasetID, true)
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

		// Select dataset ID to contribute to
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Add data identifier to dataset
		dataIdentifier, err := prompt.String("dataIdentifier", 1, actions.MaxMetadataSize-actions.MaxTextSize)
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
			DatasetID:      datasetID,
			DataLocation:   []byte("default"),
			DataIdentifier: []byte(dataIdentifier),
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
		datasetIDStr, err := prompt.String("datasetID", 1, consts.MaxInt)
		if err != nil {
			return err
		}
		datasetID, err := ids.FromString(datasetIDStr)
		if err != nil {
			return err
		}

		// Get pending data contributions info
		hutils.Outf("Retrieving pending data contributions info for datasetID: %s\n", datasetID)
		contributions, err := handler.GetDataContributionPendingInfo(ctx, ncli, datasetID)
		if err != nil {
			return err
		}
		if len(contributions) == 0 {
			hutils.Outf("{{red}}This contribution does not exist{{/}}\n")
			hutils.Outf("{{red}}exiting...{{/}}\n")
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
		datasetID, err := prompt.ID("datasetID")
		if err != nil {
			return err
		}

		// Select the contributor
		contributor, err := prompt.Address("contributor")
		if err != nil {
			return err
		}

		// Choose unique id for the NFT to be minted
		uniqueIDStr, err := prompt.String("unique nft #", 1, actions.MaxTextSize)
		if err != nil {
			return err
		}
		uniqueID, err := strconv.ParseUint(uniqueIDStr, 10, 64)
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
			DatasetID:                 datasetID,
			Contributor:               contributor,
			UniqueNFTIDForContributor: uniqueID,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
