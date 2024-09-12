// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
)

type Hub interface {
	InitiateContributeDataset(ctx context.Context, datasetID ids.ID, dataLocation, dataIdentifier []byte, contributor codec.Address) error
	CompleteContributeDataset(ctx context.Context, datasetID ids.ID, contributor codec.Address) (DataContribution, error)
	GetVMMutableState() (state.Mutable, error)
}

// GetMarketplace returns the singleton instance of Marketplace
func GetMarketplace() Hub {
	return marketplace
}
