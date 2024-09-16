// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/x/merkledb"
)

type Controller interface {
	Logger() logging.Logger
}

type NuklaiVM interface {
	State() (merkledb.MerkleDB, error)
}
