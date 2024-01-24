// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	hutils "github.com/ava-labs/hypersdk/utils"
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
		_, _, _, err = handler.GetEmissionInfo(ctx, ncli)
		if err != nil {
			return err
		}

		return nil
	},
}

var emissionValidatorsCmd = &cobra.Command{
	Use: "validators",
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

var emissionStakeCmd = &cobra.Command{
	Use: "user-stake-info",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Get current list of validators
		validators, err := ncli.Validators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s NodePublicKey=%s\n",
				i,
				validators[i].NodeID,
				validators[i].NodePublicKey,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("choose validator whom you have staked to", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]

		// Get the address to look up
		stakeOwner, err := handler.Root().PromptAddress("address to get staking info for")
		if err != nil {
			return err
		}

		// Get user stake info
		_, err = handler.GetUserStake(ctx, ncli, validatorChosen.NodeID, stakeOwner)
		if err != nil {
			return err
		}

		return nil
	},
}
