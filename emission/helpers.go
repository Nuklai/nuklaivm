package emission

import (
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
)

var (
	once sync.Once
)

type Validator struct {
	IsActive                 bool       `json:"isActive"`          // Indicates if the validator is currently active
	NodeID                   ids.NodeID `json:"nodeID"`            // Node ID of the validator
	PublicKey                []byte     `json:"publicKey"`         // Public key of the validator
	StakedAmount             uint64     `json:"stakedAmount"`      // Total amount staked by the validator
	UnclaimedStakedReward    uint64     `json:"stakedReward"`      // Total rewards accumulated by the validator
	DelegationFeeRate        float64    `json:"delegationFeeRate"` // Fee rate for delegations
	DelegatedAmount          uint64     `json:"delegatedAmount"`   // Total amount delegated to the validator
	UnclaimedDelegatedReward uint64     `json:"delegatedReward"`   // Total rewards accumulated by the delegators

	DelegatorsLastClaim map[codec.Address]uint64 // Map of delegator addresses to their last claim block height
	epochRewards        map[uint64]uint64        // Rewards per epoch
	stakeStartTime      time.Time                // Start time of the stake
	stakeEndTime        time.Time                // End time of the stake
}

type EmissionAccount struct {
	Address          codec.Address `json:"address"`
	UnclaimedBalance uint64        `json:"unclaimedBalance"`
}

type EpochTracker struct {
	BaseAPR        float64 `json:"baseAPR"`        // Base APR to use
	BaseValidators uint64  `json:"baseValidators"` // Base number of validators to use
	EpochLength    uint64  `json:"epochLength"`    // Number of blocks per reward epoch
}

func distributeValidatorRewards(totalValidatorReward uint64, delegationFeeRate float64, delegatedAmount uint64) (uint64, uint64) {
	delegationRewards := uint64(0)
	if delegatedAmount > 0 {
		delegationRewards = uint64(float64(totalValidatorReward) * delegationFeeRate)
	}
	validatorRewards := totalValidatorReward - delegationRewards
	return validatorRewards, delegationRewards
}
