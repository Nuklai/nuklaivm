// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	hutils "github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

type DatasetConfig struct {
	// Collateral needed to start the contribution process to the dataset
	CollateralForDataContribution uint64 `json:"collateralForDataContribution"`

	// Minumum amount of blocks to subscribe to
	MinBlocksToSubscribe uint64 `json:"minBlocksToSubscribe"`
}

func GetDatasetConfig() DatasetConfig {
	collateralForDataContribution, _ := hutils.ParseBalance("1", nconsts.Decimals) // 1 NAI

	return DatasetConfig{
		CollateralForDataContribution: collateralForDataContribution,
		MinBlocksToSubscribe:          100, // TODO: 720(1 hour) for production
	}
}
