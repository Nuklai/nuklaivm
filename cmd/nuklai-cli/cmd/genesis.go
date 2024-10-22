// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"encoding/json"
	"os"

	"github.com/fatih/color"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/fees"
	hgenesis "github.com/ava-labs/hypersdk/genesis"
)

var genesisCmd = &cobra.Command{
	Use: "genesis",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

type GenesisCustomAllocation struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

var genGenesisCmd = &cobra.Command{
	Use:   "generate [custom allocations file] [emission balancer file] [options]",
	Short: "Creates a new genesis in the default location",
	PreRunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 2 {
			return ErrInvalidArgs
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		a, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		// Read custom allocations file
		var allocs []*GenesisCustomAllocation
		if err := json.Unmarshal(a, &allocs); err != nil {
			return err
		}
		// Convert []*GenesisCustomAllocation to []*hgenesis.CustomAllocation
		convertedAllocs := make([]*hgenesis.CustomAllocation, len(allocs))
		for i, alloc := range allocs {
			convertedAddress, err := codec.StringToAddress(alloc.Address)
			if err != nil {
				return err
			}
			convertedAllocs[i] = &hgenesis.CustomAllocation{
				Address: convertedAddress,
				Balance: alloc.Balance,
			}
		}

		// Read emission balancer file
		eb, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		var emissionBalancer genesis.EmissionBalancer
		if err := json.Unmarshal(eb, &emissionBalancer); err != nil {
			return err
		}

		genesis := genesis.NewGenesis(convertedAllocs, emissionBalancer)
		if len(minUnitPrice) > 0 {
			d, err := fees.ParseDimensions(minUnitPrice)
			if err != nil {
				return err
			}
			genesis.Rules.MinUnitPrice = d
		}
		if len(maxBlockUnits) > 0 {
			d, err := fees.ParseDimensions(maxBlockUnits)
			if err != nil {
				return err
			}
			genesis.Rules.MaxBlockUnits = d
		}
		if len(windowTargetUnits) > 0 {
			d, err := fees.ParseDimensions(windowTargetUnits)
			if err != nil {
				return err
			}
			genesis.Rules.WindowTargetUnits = d
		}
		if minBlockGap >= 0 {
			genesis.Rules.MinBlockGap = minBlockGap
		}

		b, err := json.Marshal(genesis)
		if err != nil {
			return err
		}
		if err := os.WriteFile(genesisFile, b, fsModeWrite); err != nil {
			return err
		}

		color.Green("created genesis and saved to %s", genesisFile)
		return nil
	},
}
