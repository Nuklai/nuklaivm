// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

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
		_, _, _, _, bcli, _, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get emission info
		_, _, _, err = handler.GetEmissionInfo(ctx, bcli)
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
		_, _, _, _, bcli, _, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get emission info
		_, err = handler.GetAllValidators(ctx, bcli)
		if err != nil {
			return err
		}

		return nil
	},
}
