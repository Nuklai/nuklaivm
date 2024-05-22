// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
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
	Use:   "modify-config",
	Short: "Modify emission balancer configuration parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ebFilePath, _ := cmd.Flags().GetString("update-emission")
		address, _ := cmd.Flags().GetString("address")
		supply, _ := cmd.Flags().GetUint64("maxsupply")
		apr, _ := cmd.Flags().GetUint64("base-apr")
		baseValidators, _ := cmd.Flags().GetUint64("base-validators")
		epoch, _ := cmd.Flags().GetUint64("epoch-length")

		newAddress := codec.EmptyAddress
		if address != "" {
			if newAddress, err = codec.ParseAddressBech32(nconsts.HRP, address); err != nil {
				return err
			}
		}

		ctx := context.Background()
		_, _, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Generate transaction
		res, _, err := sendAndWait(ctx, nil, &actions.ModifyEmissionConfigParams{
			MaxSupply:             supply,
			TrackerBaseAPR:        apr,
			TrackerBaseValidators: baseValidators,
			TrackerEpochLength:    epoch,
			AccountAddress:        newAddress,
		}, hcli, hws, ncli, factory, true)

		if err != nil {
			return err
		}

		if res && ebFilePath != "" {
			emissionBalancer := genesis.EmissionBalancer{}

			eb, err := os.ReadFile(ebFilePath)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(eb, &emissionBalancer); err != nil {
				return err
			}
			if supply > 0 && supply != emissionBalancer.MaxSupply {
				emissionBalancer.MaxSupply = supply
			}
			if apr > 0 && apr != emissionBalancer.BaseAPR {
				emissionBalancer.BaseAPR = apr
			}
			if baseValidators > 0 && baseValidators != emissionBalancer.BaseValidators {
				emissionBalancer.BaseValidators = baseValidators
			}
			if epoch > 0 && epoch != emissionBalancer.EpochLength {
				emissionBalancer.EpochLength = epoch
			}
			if address != "" && address != emissionBalancer.EmissionAddress {
				emissionBalancer.EmissionAddress = address
			}
			e, err := json.Marshal(emissionBalancer)
			if err != nil {
				return err
			}
			if err := os.WriteFile(ebFilePath, e, fsModeWrite); err != nil {
				return err
			}
			// modify emission balancer file
			color.Green("modified emission balancer file and saved to %s", ebFilePath)
		}
		fmt.Println(handler.GetEmissionInfo(ctx, ncli))
		return nil
	},
}
