// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/ava-labs/hypersdk/utils"
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
		clients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}

		// Get emission info
		_, _, _, err = handler.GetEmissionInfo(ctx, clients[0])
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
		clients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}

		// Get validators info
		_, err = handler.GetAllValidators(ctx, clients[0])
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
		clients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := clients[0].Validators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
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
		_, err = handler.GetUserStake(ctx, clients[0], validatorChosen.NodeID, stakeOwner)
		if err != nil {
			return err
		}

		return nil
	},
}
