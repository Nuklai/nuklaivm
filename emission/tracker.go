// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
)

type Tracker interface {
	AddToTotalSupply(amount uint64) uint64
	GetRewardsPerEpoch() uint64
	GetAPRForValidators() float64
	GetNumDelegators(nodeID ids.NodeID) int
	CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error)
	RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate uint64) error
	WithdrawValidatorStake(nodeID ids.NodeID) (uint64, error)
	DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeStartBlock, stakeEndBlock, stakedAmount uint64) error
	UndelegateUserStake(nodeID ids.NodeID, actor codec.Address) (uint64, error)
	ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error)
	MintNewNAI() uint64
	DistributeFees(fee uint64)
	GetStakedValidator(nodeID ids.NodeID) []*Validator
	GetAllValidators(ctx context.Context) []*Validator
	GetDelegatorsForValidator(nodeID ids.NodeID) ([]*Delegator, error)
	GetLastAcceptedBlockTimestamp() time.Time
	GetLastAcceptedBlockHeight() uint64
	GetEmissionValidators() map[ids.NodeID]*Validator
	GetInfo() (emissionAccount EmissionAccount, totalSupply uint64, maxSupply uint64, totalStaked uint64, epochTracker EpochTracker)
}

// GetEmission returns the singleton instance of Emission
func GetEmission() Tracker {
	return emission
}
