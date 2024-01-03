// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var emissionbalancerCmd = &cobra.Command{
	Use: "emissionbalancer",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var emissionbalancerInfoCmd = &cobra.Command{
	Use: "info",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		_, _, _, _, bcli, _, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get emission balancer info
		_, _, _, _, err = handler.GetEmissionBalancerInfo(ctx, bcli)
		if err != nil {
			return err
		}

		return nil
	},
}
