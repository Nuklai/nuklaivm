// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"sync"
)

var _ Hub = (*Marketplace)(nil)

type Marketplace struct {
	c Controller

	lock sync.RWMutex
}

// NewMarketplace initializes the Marketplace struct with initial parameters
func NewMarketplace(c Controller) *Marketplace {
	once.Do(func() {
		c.Logger().Info("Initializing marketplace")

		marketplace = &Marketplace{ // Create the Marketplace instance with initialized values
			c: c,
		}
	})
	return marketplace.(*Marketplace)
}
