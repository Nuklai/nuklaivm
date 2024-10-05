// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/consts"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

const (
	// Active state
	assetAccountBalancePrefix byte = iota // 0x0
	heightPrefix                          // 0x1
	timestampPrefix                       // 0x2
	feePrefix                             // 0x3
	accountsPrefix                        // 0x4
	contractsPrefix                       // 0x5

	accountContractPrefix // 0x6
	accountStatePrefix    // 0x7

	validatorStakePrefix // 0x8
	delegatorStakePrefix // 0x9

	assetInfoPrefix               // 0xa
	assetNFTPrefix                // 0xb
	datasetInfoPrefix             // 0xc
	marketplaceContributionPrefix // 0xd
)

var (
	failureByte = byte(0x0)
	successByte = byte(0x1)
)

var (
	NAIAddress codec.Address
)

func init() {
	NAIAddress = AssetAddress(consts.AssetFungibleTokenID, []byte(consts.Name), []byte(consts.Symbol), consts.Decimals, []byte(consts.Metadata), []byte(consts.Metadata), codec.EmptyAddress)
}
