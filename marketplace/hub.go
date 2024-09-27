// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
)

type Hub interface {
	InitiateContributeDataset(datasetID ids.ID, dataLocation, dataIdentifier []byte, contributor codec.Address) error
	CompleteContributeDataset(datasetID ids.ID, contributor codec.Address) (DataContribution, error)
	GetDataContribution(datasetID ids.ID, owner codec.Address) ([]DataContribution, error)
}

// GetMarketplace returns the singleton instance of Marketplace
func GetMarketplace() Hub {
	return marketplace
}
