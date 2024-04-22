package emission

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
)

type Tracker interface {
	GetStakedValidator(nodeID ids.NodeID) []*Validator
	GetAllValidators(ctx context.Context) []*Validator
	GetLastAcceptedBlockTimestamp() time.Time
	GetLastAcceptedBlockHeight() uint64
	GetEmissionValidators() map[ids.NodeID]*Validator
	DistributeFees(fee uint64)
	MintNewNAI() uint64
	ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error)
	UndelegateUserStake(nodeID ids.NodeID, actor codec.Address, stakeAmount uint64) (uint64, error)
	DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeAmount uint64) error
	WithdrawValidatorStake(nodeID ids.NodeID) (uint64, error)
	RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakeStartTime, stakeEndTime, stakedAmount, delegationFeeRate uint64) error
	CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address, currentBlockHeight uint64) (uint64, error)
	GetRewardsPerEpoch() uint64
	GetAPRForValidators() float64
	GetNumDelegators(nodeID ids.NodeID) int
	AddToTotalSupply(amount uint64) uint64
	GetInfo() (emissionAccount EmissionAccount, totalSupply uint64, maxSupply uint64, totalStaked uint64, epochTracker EpochTracker)
}

// GetEmission returns the singleton instance of Emission
func GetEmission() Tracker {
	return emission
}
