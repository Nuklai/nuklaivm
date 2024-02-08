// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
)

type Controller interface {
	Genesis() *genesis.Genesis
	Tracer() trace.Tracer
	GetTransaction(context.Context, ids.ID) (bool, int64, bool, chain.Dimensions, uint64, error)
	GetAssetFromState(context.Context, ids.ID) (bool, []byte, uint8, []byte, uint64, codec.Address, bool, error)
	GetBalanceFromState(context.Context, codec.Address, ids.ID) (uint64, error)
	GetLoanFromState(context.Context, ids.ID, ids.ID) (uint64, error)

	GetEmissionInfo() (uint64, uint64, uint64, *emission.EmissionAccount, error)
	GetAllValidators() ([]*emission.Validator, error)
	GetValidator(nodeID ids.NodeID) (*emission.Validator, error)
	GetUserStake(nodeID ids.NodeID, owner string) (*emission.UserStake, error)
	GetValidatorFromState(ctx context.Context, stakeID ids.ID) (
		bool, // exists
		ids.NodeID, // NodeID
		uint64, // StakedAmount
		uint64, // EndLockUp
		codec.Address, // Owner
		error,
	)
}
