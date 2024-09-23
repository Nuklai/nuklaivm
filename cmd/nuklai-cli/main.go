// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// "nuklai-cli" implements nuklaivm client operation interface.
package main

import (
	"os"

	"github.com/nuklai/nuklaivm/cmd/nuklai-cli/cmd"

	"github.com/ava-labs/hypersdk/utils"
)

func main() {
	if err := cmd.Execute(); err != nil {
		utils.Outf("{{red}}nuklai-cli exited with error:{{/}} %+v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
