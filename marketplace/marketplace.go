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
	for _, contrib := range m.tempDataContributions[datasetID] {
		if contrib.Contributor == contributor {
			return ErrAlreadyContributedToThisDataset
		}
	}

	// Add the data contribution to the list of contributions
	m.tempDataContributions[datasetID] = append(m.tempDataContributions[datasetID], &DataContribution{
		DataLocation:   dataLocation,
		DataIdentifier: dataIdentifier,
		Contributor:    contributor,
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
	for i, contrib := range m.tempDataContributions[datasetID] {
		if contrib.Contributor == contributor {
			found = true
			// Get the data contribution
			data = *m.tempDataContributions[datasetID][i]

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

// GetDataContributionByOwner retrieves the data contribution for a given dataset ID and owner (contributor).
func (m *Marketplace) GetDataContributionByOwner(_ context.Context, datasetID ids.ID, owner codec.Address) (DataContribution, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	// Initialize an empty DataContribution to return in case of an error
	var contribution DataContribution

	// Check if the dataset exists
	contributions, exists := m.tempDataContributions[datasetID]
	if !exists {
		return contribution, ErrDatasetNotFound
	}

	// Search for the contribution by the specified owner
	for _, contrib := range contributions {
		if contrib.Contributor == owner {
			return *contrib, nil
		}
	}

	// Return an error if no contribution is found for the specified owner
	return contribution, ErrContributionNotFound
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
