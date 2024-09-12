// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"context"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"go.uber.org/zap"
)

var _ Hub = (*Marketplace)(nil)

type Marketplace struct {
	c        Controller
	nuklaivm NuklaiVM

	tempDataContributions map[ids.ID][]*DataContribution
	tempDataContribors    map[ids.ID][]codec.Address

	lock sync.RWMutex
}

// NewMarketplace initializes the Marketplace struct with initial parameters
func NewMarketplace(c Controller, vm NuklaiVM) *Marketplace {
	once.Do(func() {
		c.Logger().Info("Initializing marketplace")

		marketplace = &Marketplace{ // Create the Marketplace instance with initialized values
			c:                     c,
			nuklaivm:              vm,
			tempDataContributions: make(map[ids.ID][]*DataContribution),
			tempDataContribors:    make(map[ids.ID][]codec.Address),
		}
	})
	return marketplace.(*Marketplace)
}

// InitiateContributeDataset initiates the contribution of a dataset
func (m *Marketplace) InitiateContributeDataset(_ context.Context, datasetID ids.ID, dataLocation, dataIdentifier []byte, contributor codec.Address) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, exists := m.tempDataContributions[datasetID]; !exists {
		m.tempDataContributions[datasetID] = []*DataContribution{}
	}

	// Check if the contributor has already contributed to this dataset
	// Each contributor can only contribute once to a dataset
	for _, contrib := range m.tempDataContribors[datasetID] {
		if contrib == contributor {
			return ErrAlreadyContributedToThisDataset
		}
	}

	// Add the contributor to the list of contributors
	m.tempDataContribors[datasetID] = append(m.tempDataContribors[datasetID], contributor)
	// Add the data contribution to the list of contributions
	m.tempDataContributions[datasetID] = append(m.tempDataContributions[datasetID], &DataContribution{
		DataLocation:   dataLocation,
		DataIdentifier: dataIdentifier,
	})

	return nil
}

// CompleteContributeDataset completes the contribution of a dataset
func (m *Marketplace) CompleteContributeDataset(_ context.Context, datasetID ids.ID, contributor codec.Address) (DataContribution, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	data := DataContribution{}
	if _, exists := m.tempDataContributions[datasetID]; !exists {
		return data, ErrDatasetNotFound
	}

	// Check if the contributor has contributed to this dataset
	var found bool
	for i, contrib := range m.tempDataContribors[datasetID] {
		if contrib == contributor {
			found = true
			// Get the data contribution
			data = *m.tempDataContributions[datasetID][i]

			// Remove the contributor from the list of contributors
			m.tempDataContribors[datasetID] = append(m.tempDataContribors[datasetID][:i], m.tempDataContribors[datasetID][i+1:]...)

			// Remove the data contribution from the list of contributions
			m.tempDataContributions[datasetID] = append(m.tempDataContributions[datasetID][:i], m.tempDataContributions[datasetID][i+1:]...)
			break
		}
	}
	if !found {
		return data, ErrContributionNotFound
	}

	return data, nil
}

func (m *Marketplace) GetVMMutableState() (state.Mutable, error) {
	m.c.Logger().Info("fetching VM state")
	stateDB, err := m.nuklaivm.State()
	if err != nil {
		m.c.Logger().Error("error fetching VM state", zap.Error(err))
		return nil, err
	}
	return state.NewSimpleMutable(stateDB), nil
}
