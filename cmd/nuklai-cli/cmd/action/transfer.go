// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package action

import (
	"context"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/cmd/nuklai-cli/cmd/handler"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/spf13/cobra"
)

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
		amount, err := handler.Root().PromptAmount("amount", consts.Decimals, balance, nil)
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