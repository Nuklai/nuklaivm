// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"
	"fmt"

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

func (c *Controller) GetEmissionInfo() (uint64, uint64, uint64, uint64, uint64, emission.EmissionAccount, emission.EpochTracker, error) {
	emissionAccount, totalSupply, maxSupply, totalStaked, epochTracker := c.emission.GetInfo()
	return c.emission.GetLastAcceptedBlockHeight(), totalSupply, maxSupply, totalStaked, c.emission.GetRewardsPerEpoch(), emissionAccount, epochTracker, nil
}

func (c *Controller) GetValidators(ctx context.Context, staked bool) ([]*emission.Validator, error) {
	if staked {
		return c.emission.GetStakedValidator(ids.EmptyNodeID), nil
	} else {
		return c.emission.GetAllValidators(ctx), nil
	}
}

func (c *Controller) GetStakedValidatorInfo(nodeID ids.NodeID) (*emission.Validator, error) {
	validators := c.emission.GetStakedValidator(nodeID)
	return validators[0], nil
}

func (c *Controller) GetValidatorStakeFromState(ctx context.Context, nodeID ids.NodeID) (
	bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	uint64, // DelegationFeeRate
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	return storage.GetRegisterValidatorStakeFromState(ctx, c.inner.ReadState, nodeID)
}

func (c *Controller) GetDelegatedUserStakeFromState(ctx context.Context, owner codec.Address, nodeID ids.NodeID) (
	bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakedAmount
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	fmt.Println("GetDelegatedUserStakeFromState")
	fmt.Println(c.inner.ReadState)
	return storage.GetDelegateUserStakeFromState(ctx, c.inner.ReadState, owner, nodeID)
}
