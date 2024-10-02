// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
)

var _ Hub = (*MockMarketplace)(nil)

type MockMarketplace struct {
	DataContribution DataContribution
}

func MockNewMarketplace(mockMarketplace *MockMarketplace) *MockMarketplace {
	marketplace = mockMarketplace
	return mockMarketplace
}

func (m *MockMarketplace) InitiateContributeDataset(datasetID ids.ID, dataLocation, dataIdentifier []byte, contributor codec.Address) error {
	return nil
}

func (m *MockMarketplace) CompleteContributeDataset(datasetID ids.ID, contributor codec.Address) (DataContribution, error) {
	return m.DataContribution, nil
}

func (m *MockMarketplace) GetDataContribution(datasetID ids.ID, owner codec.Address) ([]DataContribution, error) {
	return []DataContribution{m.DataContribution}, nil
}
