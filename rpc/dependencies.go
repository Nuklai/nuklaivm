// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
)

type Controller interface {
	Genesis() *genesis.Genesis
	Tracer() trace.Tracer
	GetTransaction(context.Context, ids.ID) (bool, int64, bool, fees.Dimensions, uint64, error)
	GetAssetFromState(context.Context, ids.ID) (bool, []byte, uint8, []byte, uint64, codec.Address, error)
	GetBalanceFromState(context.Context, codec.Address, ids.ID) (uint64, error)

	GetEmissionInfo() (uint64, uint64, uint64, uint64, uint64, emission.EmissionAccount, emission.EpochTracker, error)
	GetValidators(ctx context.Context, staked bool) ([]*emission.Validator, error)
	GetStakedValidatorInfo(nodeID ids.NodeID) (*emission.Validator, error)
	GetValidatorStakeFromState(ctx context.Context, nodeID ids.NodeID) (
		bool, // exists
		uint64, // StakeStartBlock
		uint64, // StakeEndBlock
		uint64, // StakedAmount
		uint64, // DelegationFeeRate
		codec.Address, // RewardAddress
		codec.Address, // OwnerAddress
		error,
	)
	GetDelegatedUserStakeFromState(ctx context.Context, owner codec.Address, nodeID ids.NodeID) (
		bool, // exists
		uint64, // StakeStartBlock
		uint64, // StakeEndBlock
		uint64, // StakedAmount
		codec.Address, // RewardAddress
		codec.Address, // OwnerAddress
		error,
	)
}
