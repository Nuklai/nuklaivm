// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/x/merkledb"
	"github.com/ava-labs/hypersdk/chain"
)

type Controller interface {
	Logger() logging.Logger
}

type NuklaiVM interface {
	CurrentValidators(ctx context.Context) (map[ids.NodeID]*validators.GetValidatorOutput, map[string]struct{})
	LastAcceptedBlock() *chain.StatelessBlock
	State() (merkledb.MerkleDB, error)
}
