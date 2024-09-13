// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"sync"

	"github.com/ava-labs/hypersdk/codec"
)

var (
	once        sync.Once
	marketplace Hub
)

type DataContribution struct {
	DataLocation   []byte        `json:"dataLocation"`
	DataIdentifier []byte        `json:"dataIdentifier"`
	Contributor    codec.Address `json:"contributor"`
}
