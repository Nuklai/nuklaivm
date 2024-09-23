// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package version

import (
	"fmt"

	"github.com/nuklai/nuklaivm/consts"
	"github.com/spf13/cobra"
)

func init() {
	cobra.EnablePrefixMatching = true
}

// NewCommand implements "nuklaivm version" command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints out the verson",
		RunE:  versionFunc,
	}
	return cmd
}

func versionFunc(*cobra.Command, []string) error {
	fmt.Printf("%s@%s (%s)\n", consts.Name, consts.Version, consts.ID)
	return nil
}
