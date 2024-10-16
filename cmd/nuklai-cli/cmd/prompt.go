// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/codec"
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

func parseAsset(label string) (codec.Address, error) {
	text := fmt.Sprintf("%s (use %s for native token)", label, consts.Symbol)
	promptText := promptui.Prompt{
		Label: text,
		Validate: func(input string) error {
			if len(input) == 0 {
				return prompt.ErrInputEmpty
			}
			if input == consts.Symbol {
				return nil
			}
			_, err := codec.StringToAddress(input)
			return err
		},
	}
	asset, err := promptText.Run()
	if err != nil {
		return codec.EmptyAddress, err
	}
	asset = strings.TrimSpace(asset)
	assetAddress := storage.NAIAddress
	if asset != consts.Symbol {
		assetAddress, err = codec.StringToAddress(asset)
		if err != nil {
			return codec.EmptyAddress, err
		}
	}
	if assetAddress == codec.EmptyAddress {
		return codec.EmptyAddress, prompt.ErrInvalidChoice
	}
	return assetAddress, nil
}
