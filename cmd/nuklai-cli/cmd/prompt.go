// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/cli/prompt"
)

func parseAmount(
	label string,
	decimals uint8,
	balance uint64,
) (uint64, error) {
	promptText := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if len(input) == 0 {
				return prompt.ErrInputEmpty
			}
			amount, err := utils.ParseBalance(input, decimals)
			if err != nil {
				return err
			}
			if amount > balance {
				return prompt.ErrInsufficientBalance
			}
			return nil
		},
	}
	rawAmount, err := promptText.Run()
	if err != nil {
		return 0, err
	}
	rawAmount = strings.TrimSpace(rawAmount)
	return utils.ParseBalance(rawAmount, decimals)
}
