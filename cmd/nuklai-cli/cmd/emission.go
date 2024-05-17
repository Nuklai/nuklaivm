// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/fatih/color"
	"github.com/nuklai/nuklaivm/actions"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/spf13/cobra"
)

var emissionCmd = &cobra.Command{
	Use: "emission",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var emissionInfoCmd = &cobra.Command{
	Use: "info",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Get emission info
		_, _, _, _, _, _, _, _, err = handler.GetEmissionInfo(ctx, ncli)
		if err != nil {
			return err
		}

		return nil
	},
}

var emissionAllValidatorsCmd = &cobra.Command{
	Use: "all-validators",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Get validators info
		_, err = handler.GetAllValidators(ctx, ncli)
		if err != nil {
			return err
		}

		return nil
	},
}

var emissionStakedValidatorsCmd = &cobra.Command{
	Use: "staked-validators",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Get validators info
		_, err = handler.GetStakedValidators(ctx, ncli)
		if err != nil {
			return err
		}

		return nil
	},
}

var emissionModifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Modify emission configuration parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		address, _ := cmd.Flags().GetString("address")
		supply, _ := cmd.Flags().GetUint64("maxsupply")
		apr, _ := cmd.Flags().GetUint64("base-apr")
		validators, _ := cmd.Flags().GetUint64("base-validators")
		epoch, _ := cmd.Flags().GetUint64("epoch-length")

		newAddress := codec.EmptyAddress
		if address != "" {
			if newAddress, err = codec.ParseAddressBech32(nconsts.HRP, address); err != nil {
				return err
			}
		}

		g := genesis.Default()
		ctx := context.Background()
		_, _, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Read emission balancer file
		eb, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		emissionBalancer := genesis.EmissionBalancer{}
		if err := json.Unmarshal(eb, &emissionBalancer); err != nil {
			return err
		}

		// Generate transaction
		if _, _, err = sendAndWait(ctx, nil, &actions.ModifyEmissionConfigParams{
			MaxSupply:             supply,
			TrackerBaseAPR:        apr,
			TrackerBaseValidators: validators,
			TrackerEpochLength:    epoch,
			AccountAddress:        newAddress,
		}, hcli, hws, ncli, factory, true); err != nil {
			return err
		}

		emissionBalancer.MaxSupply = supply
		emissionBalancer.BaseAPR = apr
		emissionBalancer.BaseValidators = validators
		emissionBalancer.EpochLength = epoch
		g.EmissionBalancer = emissionBalancer

		b, err := json.Marshal(g)
		if err != nil {
			return err
		}
		if err := os.WriteFile(genesisFile, b, fsModeWrite); err != nil {
			return err
		}

		// modify emission balancer json
		color.Green("modified genesis and saved to %s", genesisFile)
		return nil
	},
}
