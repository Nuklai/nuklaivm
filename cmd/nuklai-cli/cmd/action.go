// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"sort"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
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
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to send
		assetID, err := handler.Root().PromptAsset("assetID", true)
		if err != nil {
			return err
		}
		_, decimals, balance, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetID, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, balance, nil)
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
			Asset: assetID,
			Value: amount,
		}, cli, ws, ncli, factory, true)
		return err
	},
}

var stakeValidatorCmd = &cobra.Command{
	Use: "stake-validator",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.Validators(ctx)
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
		nodeID := validatorChosen.NodeID

		// Get balance info
		balance, err := handler.GetBalance(ctx, ncli, priv.Address, ids.Empty)
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
		}, cli, ws, ncli, factory, true)
		return err
	},
}

var unstakeValidatorCmd = &cobra.Command{
	Use: "unstake-validator",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.Validators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
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
		keyIndex, err := handler.Root().PromptChoice("validator to unstake from", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		owner, err := codec.AddressBech32(nconsts.HRP, priv.Address)
		if err != nil {
			return err
		}
		stake, err := ncli.UserStakeInfo(ctx, nodeID, owner)
		if err != nil {
			return err
		}

		if len(stake.StakeInfo) == 0 {
			utils.Outf("{{red}}user is not staked to this validator{{/}}\n")
			return nil
		}
		// Get current height
		_, currentHeight, _, err := cli.Accepted(ctx)
		if err != nil {
			return err
		}
		// Make sure to iterate over the stake info map in the same order every time
		keys := make([]ids.ID, 0, len(stake.StakeInfo))
		for k := range stake.StakeInfo {
			keys = append(keys, k)
		}
		// Sorting based on string representation
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		// Show stake info to the user
		utils.Outf("{{cyan}}stake info:{{/}}\n")
		for index, txID := range keys {
			stakeInfo := stake.StakeInfo[txID]
			utils.Outf(
				"{{yellow}}%d:{{/}} TxID=%s StakedAmount=%d StartLockUpHeight=%d CurrentHeight=%d\n",
				index,
				txID.String(),
				stakeInfo.Amount,
				stakeInfo.StartLockUp,
				currentHeight,
			)
		}

		// Select the stake Id to unstake
		stakeIndex, err := handler.Root().PromptChoice("stake ID to unstake", len(stake.StakeInfo))
		if err != nil {
			return err
		}
		stakeChosen := stake.StakeInfo[keys[stakeIndex]]
		stakeID := stakeChosen.TxID

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.UnstakeValidator{
			Stake:  stakeID,
			NodeID: nodeID.Bytes(),
		}, cli, ws, ncli, factory, true)
		return err
	},
}
