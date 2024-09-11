// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

type Hub interface{}

// GetMarketplace returns the singleton instance of Marketplace
func GetMarketplace() Hub {
	return marketplace
}
