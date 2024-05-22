// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"encoding/json"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/genesis"
)

var genesisCmd = &cobra.Command{
	Use: "genesis",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var genGenesisCmd = &cobra.Command{
	Use:   "generate [custom allocates file] [emission balancer file] [options]",
	Short: "Creates a new genesis in the default location",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return ErrInvalidArgs
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		g := genesis.Default()
		if len(minUnitPrice) > 0 {
			d, err := chain.ParseDimensions(minUnitPrice)
			if err != nil {
				return err
			}
			g.MinUnitPrice = d
		}
		if len(maxBlockUnits) > 0 {
			d, err := chain.ParseDimensions(maxBlockUnits)
			if err != nil {
				return err
			}
			g.MaxBlockUnits = d
		}
		if len(windowTargetUnits) > 0 {
			d, err := chain.ParseDimensions(windowTargetUnits)
			if err != nil {
				return err
			}
			g.WindowTargetUnits = d
		}
		if minBlockGap >= 0 {
			g.MinBlockGap = minBlockGap
		}

		// Read custom allocations file
		a, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		allocs := []*genesis.CustomAllocation{}
		if err := json.Unmarshal(a, &allocs); err != nil {
			return err
		}
		g.CustomAllocation = allocs

		// Read emission balancer file
		eb, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		emissionBalancer := genesis.EmissionBalancer{}
		if err := json.Unmarshal(eb, &emissionBalancer); err != nil {
			return err
		}
		totalSupply := uint64(0)
		for _, alloc := range allocs {
			if _, err := codec.ParseAddressBech32(nconsts.HRP, alloc.Address); err != nil {
				return err
			}
			totalSupply += alloc.Balance
		}
		if _, err := codec.ParseAddressBech32(nconsts.HRP, emissionBalancer.EmissionAddress); err != nil {
			return err
		}
		g.EmissionBalancer = emissionBalancer

		b, err := json.Marshal(g)
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
