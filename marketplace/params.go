// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"github.com/ava-labs/avalanchego/ids"

	hutils "github.com/ava-labs/hypersdk/utils"
)

type DatasetConfig struct {
	// Collateral Asset ID for data contribution
	CollateralAssetIDForDataContribution ids.ID `json:"collateralAssetIDForDataContribution"`

	// Collateral needed to start the contribution process to the dataset
	CollateralAmountForDataContribution uint64 `json:"collateralAmountForDataContribution"`

	// Minimum amount of blocks to subscribe to
	MinBlocksToSubscribe uint64 `json:"minBlocksToSubscribe"`
}

func GetDatasetConfig() DatasetConfig {
	collateralAmountForDataContribution, _ := hutils.ParseBalance("1") // 1 NAI

	return DatasetConfig{
		CollateralAssetIDForDataContribution: ids.Empty, // Using NAI as collateral
		CollateralAmountForDataContribution:  collateralAmountForDataContribution,
		MinBlocksToSubscribe:                 5, // TODO: 720(1 hour) for production
	}
}
