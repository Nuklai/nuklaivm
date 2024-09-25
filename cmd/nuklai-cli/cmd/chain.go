// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/cobra"
)

var chainCmd = &cobra.Command{
	Use: "chain",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var importChainCmd = &cobra.Command{
	Use: "import",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) == 1 {
			return handler.ImportChain(args[0])
		}
		return handler.ImportChain("http://127.0.0.1:9650/ext/bc/nuklaivm")
	},
}

var importCliChainCmd = &cobra.Command{
	Use: "import-cli [path]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return ErrInvalidArgs
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		return handler.ImportCLI(args[0])
	},
}

var setChainCmd = &cobra.Command{
	Use: "set",
	RunE: func(*cobra.Command, []string) error {
		return handler.Root().SetDefaultChain()
	},
}

var chainInfoCmd = &cobra.Command{
	Use: "info",
	RunE: func(*cobra.Command, []string) error {
		return handler.Root().PrintChainInfo()
	},
}

var watchChainCmd = &cobra.Command{
	Use: "watch",
	RunE: func(*cobra.Command, []string) error {
		return handler.Root().WatchChain(hideTxs)
	},
}
