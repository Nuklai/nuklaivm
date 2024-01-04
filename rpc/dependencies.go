// Copyright (C) 2023, AllianceBlock. All rights reserved.
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
	GetBalanceFromState(context.Context, codec.Address) (uint64, error)
	GetEmissionInfo(context.Context) (uint64, uint64, uint64, error)
	GetAllValidators(ctx context.Context) ([]*emission.Validator, error)
	GetValidator(ctx context.Context, nodeID string) (*emission.Validator, error)
	GetUserStake(ctx context.Context, nodeID string, owner string) (*emission.UserStake, error)
	GetValidatorFromState(ctx context.Context, stakeID ids.ID) (
		bool, // exists
		ids.NodeID, // NodeID
		uint64, // StakedAmount
		uint64, // EndLockUp
		codec.Address, // Owner
		error,
	)
}
