// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/fees"
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
) (bool, int64, bool, fees.Dimensions, uint64, error) {
	return storage.GetTransaction(ctx, c.metaDB, txID)
}

func (c *Controller) GetAssetFromState(
	ctx context.Context,
	asset ids.ID,
) (bool, uint8, []byte, []byte, uint8, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	return storage.GetAssetFromState(ctx, c.inner.ReadState, asset)
}

func (c *Controller) GetAssetNFTFromState(
	ctx context.Context,
	nft ids.ID,
) (bool, ids.ID, uint64, []byte, codec.Address, error) {
	return storage.GetAssetNFTFromState(ctx, c.inner.ReadState, nft)
}

func (c *Controller) GetBalanceFromState(
	ctx context.Context,
	addr codec.Address,
	asset ids.ID,
) (uint64, error) {
	return storage.GetBalanceFromState(ctx, c.inner.ReadState, addr, asset)
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
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	return storage.GetDelegateUserStakeFromState(ctx, c.inner.ReadState, owner, nodeID)
}

func (c *Controller) GetDatasetFromState(
	ctx context.Context,
	datasetID ids.ID,
) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	return storage.GetAssetDatasetFromState(ctx, c.inner.ReadState, datasetID)
}
