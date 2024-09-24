// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

// State
// / (height) => store in root
//   -> [heightPrefix] => height
// 0x0/ (balance)
//   -> [owner] => balance
// 0x1/ (hypersdk-height)
// 0x2/ (hypersdk-timestamp)
// 0x3/ (hypersdk-fee)
// 0x4/ (account-storage)
// 0x4/address/0x1 (address associated contract)
// 0x4/address/0x1 (address associated state)
// 0x5/ (contracts-storage)

const (
	// Active state
	balancePrefix   = 0x0
	heightPrefix    = 0x1
	timestampPrefix = 0x2
	feePrefix       = 0x3
	accountsPrefix  = 0x4
	contractsPrefix = 0x5

	accountContractPrefix = 0x0
	accountStatePrefix    = 0x1

	assetPrefix    = 0x0
	assetNFTPrefix = 0x1
	datasetPrefix  = 0x2

	registerValidatorStakePrefix = 0x0
	delegateUserStakePrefix      = 0x0
)

var (
	failureByte = byte(0x0)
	successByte = byte(0x1)
)
