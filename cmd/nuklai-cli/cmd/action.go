// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/actions"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/spf13/cobra"
)

var actionCmd = &cobra.Command{
	Use: "action",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var transferCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get balance info
		balance, err := handler.GetBalance(ctx, bcli, priv.Address)
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", nconsts.Decimals, balance, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.Transfer{
			To:    recipient,
			Value: amount,
		}, cli, bcli, ws, factory, true)
		return err
	},
}

var stakeValidatorCmd = &cobra.Command{
	Use: "stake-validator",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := bcli.Validators(ctx)
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
		keyIndex, err := handler.Root().PromptChoice("validator to stake to", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID, err := ids.NodeIDFromString(validatorChosen.NodeID)
		if err != nil {
			return err
		}

		// Get balance info
		balance, err := handler.GetBalance(ctx, bcli, priv.Address)
		if balance == 0 || err != nil {
			return err
		}

		// Select staked amount
		stakedAmount, err := handler.Root().PromptAmount("Staked amount", nconsts.Decimals, balance, nil)
		if err != nil {
			return err
		}

		// Select endLockUp block height
		endLockUp, err := handler.Root().PromptTime("End LockUp Height")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.StakeValidator{
			NodeID:       nodeID.Bytes(),
			StakedAmount: stakedAmount,
			EndLockUp:    uint64(endLockUp),
		}, cli, bcli, ws, factory, true)
		return err
	},
}
