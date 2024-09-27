// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/near/borsh-go"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/spf13/cobra"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/utils"
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

		// Get assetID
		assetID, err := prompt.Asset("assetID", consts.Symbol, true)
		if err != nil {
			return err
		}

		// Get balance info
		balance, _, _, _, decimals, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetID, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := prompt.Address("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := prompt.Amount("amount", decimals, balance, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, txID, err := sendAndWait(ctx, []chain.Action{&actions.Transfer{
			AssetID: assetID,
			To:      recipient,
			Value:   amount,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		utils.Outf("{{green}}txID: {{/}}%s\n", txID)
		return processResult(result)
	},
}

var publishFileCmd = &cobra.Command{
	Use: "publishFile",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select contract bytes
		path, err := prompt.String("contract file", 1, 1000)
		if err != nil {
			return err
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ContractPublish{
			ContractBytes: bytes,
		}}, cli, bcli, ws, factory)

		if result != nil && result.Success {
			utils.Outf("{{green}}fee consumed:{{/}} %s\n", utils.FormatBalance(result.Fee, consts.Decimals))

			utils.Outf(hexutils.BytesToHex(result.Outputs[0]) + "\n")
		}
		return err
	},
}

var callCmd = &cobra.Command{
	Use: "call",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get balance info
		balance, err := handler.GetBalance(ctx, bcli, priv.Address, ids.Empty)
		if balance == 0 || err != nil {
			return err
		}

		// Select contract
		contractAddress, err := prompt.Address("contract address")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := prompt.Amount("amount", consts.Decimals, balance, nil)
		if err != nil {
			return err
		}

		// Select function
		function, err := prompt.String("function", 0, 100)
		if err != nil {
			return err
		}

		action := &actions.ContractCall{
			ContractAddress: contractAddress,
			Value:           amount,
			Function:        function,
		}

		specifiedStateKeysSet, fuel, err := bcli.Simulate(ctx, *action, priv.Address)
		if err != nil {
			return err
		}

		action.SpecifiedStateKeys = make([]actions.StateKeyPermission, 0, len(specifiedStateKeysSet))
		for key, value := range specifiedStateKeysSet {
			action.SpecifiedStateKeys = append(action.SpecifiedStateKeys, actions.StateKeyPermission{Key: key, Permission: value})
		}
		action.Fuel = fuel

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{action}, cli, bcli, ws, factory)
		if result != nil && result.Success {
			utils.Outf("{{green}}fee consumed:{{/}} %s\n", utils.FormatBalance(result.Fee, consts.Decimals))

			utils.Outf(hexutils.BytesToHex(result.Outputs[0]) + "\n")
			switch function {
			case "balance":
				{
					var intValue uint64
					err := borsh.Deserialize(&intValue, result.Outputs[0])
					if err != nil {
						return err
					}
					utils.Outf("%s\n", utils.FormatBalance(intValue, consts.Decimals))
				}
			case "get_value":
				{
					var intValue int64
					err := borsh.Deserialize(&intValue, result.Outputs[0])
					if err != nil {
						return err
					}
					utils.Outf("%d\n", intValue)
				}
			}
		}
		return err
	},
}

var deployCmd = &cobra.Command{
	Use: "deploy",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		contractID, err := prompt.Bytes("contract id")
		if err != nil {
			return err
		}

		creationInfo, err := prompt.Bytes("creation info")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ContractDeploy{
			ContractID:   contractID,
			CreationInfo: creationInfo,
		}}, cli, bcli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
