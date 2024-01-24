// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/storage"
)

func (c *Controller) Genesis() *genesis.Genesis {
	return c.genesis
}

func (c *Controller) Logger() logging.Logger {
	return c.inner.Logger()
}

func (c *Controller) Tracer() trace.Tracer {
	return c.inner.Tracer()
}

func (c *Controller) GetTransaction(
	ctx context.Context,
	txID ids.ID,
) (bool, int64, bool, chain.Dimensions, uint64, error) {
	return storage.GetTransaction(ctx, c.metaDB, txID)
}

func (c *Controller) GetAssetFromState(
	ctx context.Context,
	asset ids.ID,
) (bool, []byte, uint8, []byte, uint64, codec.Address, bool, error) {
	return storage.GetAssetFromState(ctx, c.inner.ReadState, asset)
}

func (c *Controller) GetBalanceFromState(
	ctx context.Context,
	addr codec.Address,
	asset ids.ID,
) (uint64, error) {
	return storage.GetBalanceFromState(ctx, c.inner.ReadState, addr, asset)
}

func (c *Controller) GetLoanFromState(
	ctx context.Context,
	asset ids.ID,
	destination ids.ID,
) (uint64, error) {
	return storage.GetLoanFromState(ctx, c.inner.ReadState, asset, destination)
}

func (c *Controller) GetEmissionInfo() (uint64, uint64, uint64, error) {
	return c.emission.GetTotalSupply(), c.emission.GetMaxSupply(), c.emission.GetRewardsPerBlock(), nil
}

func (c *Controller) GetAllValidators() ([]*emission.Validator, error) {
	return c.emission.GetValidator(ids.EmptyNodeID), nil
}

func (c *Controller) GetValidator(nodeID ids.NodeID) (*emission.Validator, error) {
	validators := c.emission.GetValidator(nodeID)
	return validators[0], nil
}

func (c *Controller) GetUserStake(nodeID ids.NodeID, owner string) (*emission.UserStake, error) {
	return c.emission.GetUserStake(nodeID, owner), nil
}

func (c *Controller) GetValidatorFromState(ctx context.Context, stakeID ids.ID) (
	bool, // exists
	ids.NodeID, // NodeID
	uint64, // StakedAmount
	uint64, // EndLockUp
	codec.Address, // Owner
	error,
) {
	return storage.GetStakeFromState(ctx, c.inner.ReadState, stakeID)
}
