// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"strings"

	"github.com/ava-labs/hypersdk/codec"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

func GetAssetAddressBySymbol(symbol string) (codec.Address, error) {
	if strings.TrimSpace(symbol) == "" || strings.EqualFold(symbol, nconsts.Symbol) {
		return storage.NAIAddress, nil
	}
	return codec.StringToAddress(symbol)
}
