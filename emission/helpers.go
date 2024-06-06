package emission

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
)

var (
	once     sync.Once
	emission Tracker
)

type Delegator struct {
	IsActive        bool   `json:"isActive"`        // Indicates if the delegator is currently active
	StakedAmount    uint64 `json:"stakedAmount"`    // Total amount staked by the delegator
	StakeStartBlock uint64 `json:"stakeStartBlock"` // Start block of the stake
	StakeEndBlock   uint64 `json:"stakeEndBlock"`   // End block of the stake
}

type Validator struct {
	IsActive                   bool       `json:"isActive"`                   // Indicates if the validator is currently active
	NodeID                     ids.NodeID `json:"nodeID"`                     // Node ID of the validator
	PublicKey                  []byte     `json:"publicKey"`                  // Public key of the validator
	StakedAmount               uint64     `json:"stakedAmount"`               // Total amount staked by the validator
	AccumulatedStakedReward    uint64     `json:"accumulatedStakedReward"`    // Total rewards accumulated by the validator
	DelegationFeeRate          float64    `json:"delegationFeeRate"`          // Fee rate for delegations
	DelegatedAmount            uint64     `json:"delegatedAmount"`            // Total amount delegated to the validator
	AccumulatedDelegatedReward uint64     `json:"accumulatedDelegatedReward"` // Total rewards accumulated by the delegators of the validator

	delegators              map[codec.Address]*Delegator
	epochRewards            map[uint64]uint64 // Rewards per epoch
	delegatedAmountPerEpoch map[uint64]uint64 // Delegated amounts per epoch
	stakeStartBlock         uint64            // Start block of the stake
	stakeEndBlock           uint64            // End block of the stake
}

type EmissionAccount struct {
	Address           codec.Address `json:"address"`
	AccumulatedReward uint64        `json:"accumulatedReward"`
}

type EpochTracker struct {
	BaseAPR        float64 `json:"baseAPR"`        // Base APR to use
	BaseValidators uint64  `json:"baseValidators"` // Base number of validators to use
	EpochLength    uint64  `json:"epochLength"`    // Number of blocks per reward epoch
}

type DelegatorEvent struct {
	ValidatorNodeID ids.NodeID
	Delegator       codec.Address
}

func distributeValidatorRewards(totalValidatorReward uint64, delegationFeeRate float64, delegatedAmount uint64) (uint64, uint64) {
	delegationRewards := uint64(0)
	if delegatedAmount > 0 {
		delegationRewards = uint64(float64(totalValidatorReward) * delegationFeeRate)
	}
	validatorRewards := totalValidatorReward - delegationRewards
	return validatorRewards, delegationRewards
}

func addValidatorEvent(events map[uint64][]*Validator, blockHeight uint64, validator *Validator) {
	if _, exists := events[blockHeight]; !exists {
		events[blockHeight] = []*Validator{}
	}
	events[blockHeight] = append(events[blockHeight], validator)
}

func addDelegatorEvent(events map[uint64][]*DelegatorEvent, blockHeight uint64, nodeID ids.NodeID, delegatorAddress codec.Address) {
	event := &DelegatorEvent{
		ValidatorNodeID: nodeID,
		Delegator:       delegatorAddress,
	}
	if _, exists := events[blockHeight]; !exists {
		events[blockHeight] = []*DelegatorEvent{}
	}
	events[blockHeight] = append(events[blockHeight], event)
}
