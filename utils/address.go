// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"encoding/binary"
	"strings"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
)

func GetAssetAddressBySymbol(symbol string) (codec.Address, error) {
	if strings.TrimSpace(symbol) == "" || strings.EqualFold(symbol, nconsts.Symbol) {
		return storage.NAIAddress, nil
	}
	return codec.StringToAddress(symbol)
}

func GenerateAddressWithIndex(address codec.Address, i uint64) codec.Address {
	actionBytes := make([]byte, codec.AddressLen+consts.Uint64Len)
	copy(actionBytes, address[:])
	binary.BigEndian.PutUint64(actionBytes[codec.AddressLen:], i)
	id := utils.ToID(actionBytes)
	return codec.CreateAddress(nconsts.AssetNonFungibleTokenID, id)
}
