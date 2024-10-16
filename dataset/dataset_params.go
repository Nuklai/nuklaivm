// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package dataset

import (
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/codec"

	hutils "github.com/ava-labs/hypersdk/utils"
)

type DatasetConfig struct {
	// Collateral Asset Address for data contribution
	CollateralAssetAddressForDataContribution codec.Address `json:"collateralAssetAddressForDataContribution"`

	// Collateral needed to start the contribution process to the dataset
	CollateralAmountForDataContribution uint64 `json:"collateralAmountForDataContribution"`

	// Minimum amount of blocks to subscribe to
	MinBlocksToSubscribe uint64 `json:"minBlocksToSubscribe"`
}

func GetDatasetConfig() DatasetConfig {
	collateralAmountForDataContribution, _ := hutils.ParseBalance("1") // 1 NAI

	return DatasetConfig{
		CollateralAssetAddressForDataContribution: storage.NAIAddress, // Using NAI as collateral
		CollateralAmountForDataContribution:       collateralAmountForDataContribution,
		MinBlocksToSubscribe:                      5, // TODO: 720(1 hour) for production
	}
}
