// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

// Metadata
// 0x0/ (tx)
//   -> [txID] => timestamp
//
// State
// / (height) => store in root
//   -> [heightPrefix] => height
//
// 0x0/ (balance)
//   -> [owner|asset] => balance

// 0x1/ (hypersdk-height)
// 0x2/ (hypersdk-timestamp)
// 0x3/ (hypersdk-fee)

// 0x4/ (hypersdk-incoming warp)
// 0x5/ (hypersdk-outgoing warp)

// 0x6/ (assets)
//   -> [asset] => metadataLen|metadata|supply|owner|warp

// 0x7/ (stake)
//   -> [nodeID] => stakeStartBlock|stakeEndBlock|stakedAmount|delegationFeeRate|rewardAddress|ownerAddress
// 0x8/ (delegate)
//   -> [owner|nodeID] => stakeStartBlock|stakedAmount|rewardAddress|ownerAddress

const (
	// metaDB
	txPrefix = 0x0

	// stateDB
	balancePrefix = 0x0

	heightPrefix    = 0x1
	timestampPrefix = 0x2
	feePrefix       = 0x3

	incomingWarpPrefix = 0x4
	outgoingWarpPrefix = 0x5

	assetPrefix                    = 0x6
	assetNFTPrefix                 = 0x7
	assetCollectionPagePrefix      = 0x8
	assetCollectionPageCountPrefix = 0x9

	registerValidatorStakePrefix = 0xA
	delegateUserStakePrefix      = 0xB

	datasetPrefix = 0xC
)

const (
	BalanceChunks                  uint16 = 1
	AssetChunks                    uint16 = 16
	AssetNFTChunks                 uint16 = 10
	AssetCollectionPageChunks      uint16 = 11
	AssetCollectionPageCountChunks uint16 = 1
	AssetDatasetChunks             uint16 = 3
	RegisterValidatorStakeChunks   uint16 = 4
	DelegateUserStakeChunks        uint16 = 2
	DatasetChunks                  uint16 = 103
)

var (
	failureByte  = byte(0x0)
	successByte  = byte(0x1)
	heightKey    = []byte{heightPrefix}
	timestampKey = []byte{timestampPrefix}
	feeKey       = []byte{feePrefix}

	balanceKeyPool = sync.Pool{
		New: func() any {
			return make([]byte, 1+codec.AddressLen+ids.IDLen+consts.Uint16Len)
		},
	}
)
