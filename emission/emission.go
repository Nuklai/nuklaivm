// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"context"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
)

var (
	emission *Emission
	once     sync.Once
)

type Delegator struct {
	IsActive        bool   `json:"isActive"`        // Indicates if the delegator is currently active
	StakedAmount    uint64 `json:"stakedAmount"`    // Total amount staked by the delegator
	StakeStartBlock uint64 `json:"stakeStartBlock"` // Start block of the stake
	StakeEndBlock   uint64 `json:"stakeEndBlock"`   // End block of the stake
	AlreadyClaimed  bool   `json:"alreadyClaimed"`  // Indicates if the delegator has already claimed rewards
}

type Validator struct {
	IsActive                 bool       `json:"isActive"`                 // Indicates if the validator is currently active
	NodeID                   ids.NodeID `json:"nodeID"`                   // Node ID of the validator
	PublicKey                []byte     `json:"publicKey"`                // Public key of the validator
	StakedAmount             uint64     `json:"stakedAmount"`             // Total amount staked by the validator
	UnclaimedStakedReward    uint64     `json:"stakedReward"`             // Total rewards accumulated by the validator
	DelegationFeeRate        float64    `json:"delegationFeeRate"`        // Fee rate for delegations
	DelegatedAmount          uint64     `json:"delegatedAmount"`          // Total amount delegated to the validator
	UnclaimedDelegatedReward uint64     `json:"unclaimedDelegatedReward"` // Total rewards accumulated by the delegators

	delegators      map[codec.Address]*Delegator
	epochRewards    map[uint64]uint64 // Rewards per epoch
	stakeStartBlock uint64            // Start block of the stake
	stakeEndBlock   uint64            // End block of the stake
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

type DelegatorEvent struct {
	ValidatorNodeID ids.NodeID
	Delegator       codec.Address
}

type Emission struct {
	c        Controller
	nuklaivm NuklaiVM

	TotalSupply     uint64          `json:"totalSupply"`     // Total supply of NAI
	MaxSupply       uint64          `json:"maxSupply"`       // Max supply of NAI
	EmissionAccount EmissionAccount `json:"emissionAccount"` // Emission Account Info

	validators  map[ids.NodeID]*Validator
	TotalStaked uint64 `json:"totalStaked"` // Total staked NAI

	EpochTracker EpochTracker `json:"epochTracker"` // Epoch Tracker Info

	activationEvents   map[uint64][]*Validator
	deactivationEvents map[uint64][]*Validator
	delegatorEvents    map[uint64][]*DelegatorEvent

	lock sync.RWMutex
}

// New initializes the Emission struct with initial parameters and sets up the validators heap
// and indices map.
func New(c Controller, vm NuklaiVM, totalSupply, maxSupply uint64, emissionAddress codec.Address) *Emission {
	once.Do(func() {
		c.Logger().Info("Initializing emission with max supply and rewards per block settings")

		if maxSupply == 0 {
			maxSupply = GetStakingConfig().RewardConfig.SupplyCap // Use the staking config's supply cap if maxSupply is not specified
		}

		emission = &Emission{ // Create the Emission instance with initialized values
			c:           c,
			nuklaivm:    vm,
			TotalSupply: totalSupply,
			MaxSupply:   maxSupply,
			EmissionAccount: EmissionAccount{ // Setup the emission account with the provided address
				Address: emissionAddress,
			},
			validators: make(map[ids.NodeID]*Validator),
			EpochTracker: EpochTracker{
				BaseAPR:        0.25, // 25% APR
				BaseValidators: 100,
				EpochLength:    10,
				// TODO: Enable this in production
				// EpochLength:    1200, // roughly 1 hour with 3 sec block time
			},
			activationEvents:   make(map[uint64][]*Validator),
			deactivationEvents: make(map[uint64][]*Validator),
			delegatorEvents:    make(map[uint64][]*DelegatorEvent),
		}
	})
	return emission
}

// GetEmission returns the singleton instance of Emission
func GetEmission() *Emission {
	return emission
}

// AddToTotalSupply increases the total supply of NAI by a specified amount, ensuring it
// does not exceed the max supply.
func (e *Emission) AddToTotalSupply(amount uint64) uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("adding to the total supply of NAI")
	if e.TotalSupply+amount > e.MaxSupply {
		amount = e.MaxSupply - e.TotalSupply // Adjust to not exceed max supply
	}
	e.TotalSupply += amount
	return e.TotalSupply
}

// CalculateUserDelegationRewards computes the rewards for a user's delegated stake to a
// validator, factoring in the delegation duration and amount.
func (e *Emission) CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.c.Logger().Info("calculating rewards for user delegation")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists and rewards haven't been claimed yet
	if _, exists := validator.delegators[actor]; !exists {
		return 0, ErrDelegatorNotFound
	}
	if validator.delegators[actor].AlreadyClaimed {
		return 0, nil // Rewards already claimed
	}

	// Iterate over each epoch within the stake period
	startEpoch := validator.delegators[actor].StakeStartBlock / e.EpochTracker.EpochLength
	endEpoch := validator.delegators[actor].StakeEndBlock / e.EpochTracker.EpochLength
	totalReward := uint64(0)

	for epoch := startEpoch; epoch < endEpoch; epoch++ {
		if reward, ok := validator.epochRewards[epoch]; ok {
			// Calculate reward for this epoch
			delegatorShare := float64(validator.delegators[actor].StakedAmount) / float64(validator.DelegatedAmount)
			epochReward := delegatorShare * float64(reward)
			totalReward += uint64(epochReward)
		}
	}

	return totalReward, nil
}

// RegisterValidatorStake adds a new validator to the heap with the specified staked amount
// and updates the total staked amount.
func (e *Emission) RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("registering validator stake")

	// Check if the validator was already registered and is active
	validator, exists := e.validators[nodeID]
	if exists && validator.IsActive {
		return ErrValidatorAlreadyRegistered
	}

	if exists {
		validator.PublicKey = bls.PublicKeyToBytes(nodePublicKey)
		validator.StakedAmount += stakedAmount
		validator.DelegationFeeRate = float64(delegationFeeRate) / 100.0
		validator.stakeStartBlock = stakeStartBlock
		validator.stakeEndBlock = stakeEndBlock
	} else {
		validator = &Validator{
			NodeID:            nodeID,
			PublicKey:         bls.PublicKeyToBytes(nodePublicKey),
			StakedAmount:      stakedAmount,
			DelegationFeeRate: float64(delegationFeeRate) / 100.0,
			delegators:        make(map[codec.Address]*Delegator),
			epochRewards:      make(map[uint64]uint64),
			stakeStartBlock:   stakeStartBlock,
			stakeEndBlock:     stakeEndBlock,
		}
		e.validators[nodeID] = validator
	}

	e.addActivationEvent(stakeStartBlock, validator)
	e.addDeactivationEvent(stakeEndBlock, validator)

	return nil
}

func (e *Emission) addActivationEvent(blockHeight uint64, validator *Validator) {
	if _, exists := e.activationEvents[blockHeight]; !exists {
		e.activationEvents[blockHeight] = []*Validator{}
	}
	e.activationEvents[blockHeight] = append(e.activationEvents[blockHeight], validator)
}

func (e *Emission) addDeactivationEvent(blockHeight uint64, validator *Validator) {
	if _, exists := e.deactivationEvents[blockHeight]; !exists {
		e.deactivationEvents[blockHeight] = []*Validator{}
	}
	e.deactivationEvents[blockHeight] = append(e.deactivationEvents[blockHeight], validator)
}

// WithdrawValidatorStake removes a validator from the heap and updates the total
// staked amount accordingly.
func (e *Emission) WithdrawValidatorStake(nodeID ids.NodeID) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("unregistering validator stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Validator claiming their rewards and resetting unclaimed rewards
	rewardAmount := validator.UnclaimedStakedReward
	validator.UnclaimedStakedReward = 0
	validator.StakedAmount = 0

	// Mark the validator as inactive
	validator.IsActive = false

	// If there are no more delegators, get the rewards and remove the validator
	if len(validator.delegators) == 0 {
		rewardAmount += validator.UnclaimedDelegatedReward
		validator.UnclaimedDelegatedReward = 0
		delete(e.validators, nodeID)
	}

	return rewardAmount, nil
}

// DelegateUserStake increases the delegated stake for a validator and rebalances the heap.
func (e *Emission) DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeStartBlock, stakeEndBlock, stakeAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("delegating user stake")

	validator, exists := e.validators[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}

	if _, exists := validator.delegators[delegatorAddress]; exists {
		return ErrDelegatorAlreadyStaked
	}

	delegator := &Delegator{
		IsActive:        false,
		StakedAmount:    stakeAmount,
		StakeStartBlock: stakeStartBlock,
		StakeEndBlock:   stakeEndBlock,
		AlreadyClaimed:  false,
	}
	validator.delegators[delegatorAddress] = delegator

	e.addDelegatorEvent(stakeStartBlock, nodeID, delegatorAddress)
	e.addDelegatorEvent(stakeEndBlock, nodeID, delegatorAddress)

	return nil
}

func (e *Emission) addDelegatorEvent(blockHeight uint64, nodeID ids.NodeID, delegatorAddress codec.Address) {
	event := &DelegatorEvent{
		ValidatorNodeID: nodeID,
		Delegator:       delegatorAddress,
	}
	if _, exists := e.delegatorEvents[blockHeight]; !exists {
		e.delegatorEvents[blockHeight] = []*DelegatorEvent{}
	}
	e.delegatorEvents[blockHeight] = append(e.delegatorEvents[blockHeight], event)
}

// UndelegateUserStake decreases the delegated stake for a validator and rebalances the heap.
func (e *Emission) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("undelegating user stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists
	if _, exists := validator.delegators[actor]; !exists {
		return 0, ErrDelegatorNotFound
	}

	// Claim rewards while undelegating
	rewardAmount, err := e.CalculateUserDelegationRewards(nodeID, actor)
	if err != nil {
		return 0, err
	}
	validator.UnclaimedDelegatedReward -= rewardAmount // Reset unclaimed rewards

	delete(validator.delegators, actor) // Remove the delegator from the list

	// If the validator is inactive and has no more delegators, remove the validator
	if !validator.IsActive && len(validator.delegators) == 0 {
		delete(e.validators, nodeID)
	}

	return rewardAmount, nil
}

// ClaimStakingRewards lets validators and delegators claim their rewards
func (e *Emission) ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("claiming staking rewards")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	rewardAmount := uint64(0)
	if actor == codec.EmptyAddress {
		// Validator claiming their rewards
		rewardAmount = validator.UnclaimedStakedReward
		validator.UnclaimedStakedReward = 0 // Reset unclaimed rewards

		// If there are no more delegators, get the rewards
		if len(validator.delegators) == 0 {
			rewardAmount += validator.UnclaimedDelegatedReward
			validator.UnclaimedDelegatedReward = 0
		}
	} else {
		// Delegator claiming their rewards
		reward, err := e.CalculateUserDelegationRewards(nodeID, actor)
		if err != nil {
			return 0, err
		}
		validator.UnclaimedDelegatedReward -= reward // Reset unclaimed rewards
		validator.delegators[actor].AlreadyClaimed = true
		rewardAmount = reward
	}

	return rewardAmount, nil
}

func (e *Emission) processEvents(blockHeight uint64) {
	if validators, exists := e.activationEvents[blockHeight]; exists {
		for _, validator := range validators {
			if !validator.IsActive {
				validator.IsActive = true
				e.TotalStaked += validator.StakedAmount + validator.DelegatedAmount
			}
		}
		delete(e.activationEvents, blockHeight)
	}

	if validators, exists := e.deactivationEvents[blockHeight]; exists {
		for _, validator := range validators {
			if validator.IsActive {
				validator.IsActive = false
				e.TotalStaked -= validator.StakedAmount + validator.DelegatedAmount
			}
		}
		delete(e.deactivationEvents, blockHeight)
	}

	if events, exists := e.delegatorEvents[blockHeight]; exists {
		for _, event := range events {
			validator, exists := e.validators[event.ValidatorNodeID]
			if !exists {
				continue
			}
			delegator, exists := validator.delegators[event.Delegator]
			if !exists {
				continue
			}
			if blockHeight >= delegator.StakeStartBlock && blockHeight < delegator.StakeEndBlock && !delegator.IsActive {
				delegator.IsActive = true
				validator.DelegatedAmount += delegator.StakedAmount
				e.TotalStaked += delegator.StakedAmount
			}
			if blockHeight >= delegator.StakeEndBlock && delegator.IsActive {
				delegator.IsActive = false
				validator.DelegatedAmount -= delegator.StakedAmount
				e.TotalStaked -= delegator.StakedAmount
			}
		}
		delete(e.delegatorEvents, blockHeight)
	}
}

func (e *Emission) MintNewNAI() uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	currentBlockHeight := e.GetLastAcceptedBlockHeight()
	e.processEvents(currentBlockHeight)

	if currentBlockHeight%e.EpochTracker.EpochLength == 0 {
		e.c.Logger().Info("minting new NAI tokens at the end of the epoch")

		totalEpochRewards := e.GetRewardsPerEpoch()
		rewardsPerStakeUnit := float64(0)
		if e.TotalStaked > 0 {
			rewardsPerStakeUnit = float64(totalEpochRewards) / float64(e.TotalStaked)
		}

		actualRewards := uint64(0)
		for _, validator := range e.validators {
			if !validator.IsActive {
				continue
			}

			validatorStake := validator.StakedAmount + validator.DelegatedAmount
			totalValidatorReward := uint64(float64(validatorStake) * rewardsPerStakeUnit)

			validatorReward, delegationReward := distributeValidatorRewards(totalValidatorReward, validator.DelegationFeeRate, validator.DelegatedAmount)

			actualRewards += validatorReward + delegationReward
			validator.UnclaimedStakedReward += validatorReward
			validator.UnclaimedDelegatedReward += delegationReward

			epochNumber := currentBlockHeight / e.EpochTracker.EpochLength
			validator.epochRewards[epochNumber] = delegationReward
		}

		e.TotalSupply += actualRewards
		return actualRewards
	}

	return 0
}

func (e *Emission) DistributeFees(fee uint64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	currentBlockHeight := e.GetLastAcceptedBlockHeight()
	e.processEvents(currentBlockHeight)

	e.c.Logger().Info("distributing transaction fees")

	if e.TotalSupply+fee > e.MaxSupply {
		fee = e.MaxSupply - e.TotalSupply
	}

	feesForEmission := fee / 2
	e.EmissionAccount.UnclaimedBalance += feesForEmission

	feesForValidators := fee - feesForEmission
	if e.TotalStaked == 0 || feesForValidators == 0 {
		return
	}

	feesPerStakeUnit := float64(feesForValidators) / float64(e.TotalStaked)

	for _, validator := range e.validators {
		if !validator.IsActive {
			continue
		}

		validatorStake := validator.StakedAmount + validator.DelegatedAmount
		totalValidatorFee := uint64(float64(validatorStake) * feesPerStakeUnit)

		validatorFee, delegationFee := distributeValidatorRewards(totalValidatorFee, validator.DelegationFeeRate, validator.DelegatedAmount)

		validator.UnclaimedStakedReward += validatorFee
		validator.UnclaimedDelegatedReward += delegationFee
	}
}

func distributeValidatorRewards(totalValidatorReward uint64, delegationFeeRate float64, delegatedAmount uint64) (uint64, uint64) {
	delegationRewards := uint64(0)
	if delegatedAmount > 0 {
		delegationRewards = uint64(float64(totalValidatorReward) * delegationFeeRate)
	}
	validatorRewards := totalValidatorReward - delegationRewards
	return validatorRewards, delegationRewards
}

// GetStakedValidator retrieves the details of a specific validator by their NodeID.
func (e *Emission) GetStakedValidator(nodeID ids.NodeID) []*Validator {
	e.c.Logger().Info("fetching staked validator")

	if nodeID == ids.EmptyNodeID {
		validators := make([]*Validator, 0, len(e.validators))
		for _, validator := range e.validators {
			validators = append(validators, validator)
		}
		return validators
	}

	// Find the validator
	if validator, exists := e.validators[nodeID]; exists {
		return []*Validator{validator}
	}
	return []*Validator{}
}

// GetAllValidators fetches the current validators from the underlying VM
func (e *Emission) GetAllValidators(ctx context.Context) []*Validator {
	e.c.Logger().Info("fetching all staked and unstaked validators")

	currentValidators, _ := e.nuklaivm.CurrentValidators(ctx)
	validators := make([]*Validator, 0, len(currentValidators))
	for nodeID, validator := range currentValidators {
		v := Validator{
			NodeID:    nodeID,
			PublicKey: bls.PublicKeyToBytes(validator.PublicKey),
		}
		stakedValidator := e.GetStakedValidator(nodeID)
		if len(stakedValidator) > 0 {
			v.StakedAmount = stakedValidator[0].StakedAmount
			v.UnclaimedStakedReward = stakedValidator[0].UnclaimedStakedReward
			v.DelegationFeeRate = stakedValidator[0].DelegationFeeRate
			v.DelegatedAmount = stakedValidator[0].DelegatedAmount
			v.UnclaimedDelegatedReward = stakedValidator[0].UnclaimedDelegatedReward
			v.delegators = stakedValidator[0].delegators
			v.epochRewards = stakedValidator[0].epochRewards
		}
		validators = append(validators, &v)
	}
	return validators
}

// GetNumDelegators returns the total number of delegators across all validators.
func (e *Emission) GetNumDelegators(nodeID ids.NodeID) int {
	e.c.Logger().Info("fetching total number of delegators")

	numDelegators := 0
	// Get delegators for all validators
	if nodeID == ids.EmptyNodeID {
		for _, validator := range e.validators {
			numDelegators += len(validator.delegators)
		}
	} else {
		// Get delegators for a specific validator
		if validator, exists := e.validators[nodeID]; exists {
			numDelegators = len(validator.delegators)
		}
	}

	return numDelegators
}

// GetDelegatorsForValidator retrieves all delegators for a specific validator by their NodeID.
func (e *Emission) GetDelegatorsForValidator(nodeID ids.NodeID) ([]*Delegator, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	e.c.Logger().Info("fetching delegators for validator")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return nil, ErrValidatorNotFound
	}

	delegators := make([]*Delegator, 0, len(validator.delegators))
	for _, delegator := range validator.delegators {
		delegators = append(delegators, delegator)
	}

	return delegators, nil
}

// GetAPRForValidators calculates the Annual Percentage Rate (APR) for validators
// based on the number of validators.
func (e *Emission) GetAPRForValidators() float64 {
	e.c.Logger().Info("getting APR for validators")

	apr := e.EpochTracker.BaseAPR // APR is expressed per year as a decimal, e.g., 0.25 for 25%
	// Beyond baseValidators, APR decreases proportionately
	baseValidators := int(e.EpochTracker.BaseValidators)
	if len(e.validators) > baseValidators {
		apr /= float64(len(e.validators)) / float64(baseValidators)
	}
	return apr
}

// GetRewardsPerEpoch calculates the rewards per epoch based on the total staked amount
// and the APR for validators.
func (e *Emission) GetRewardsPerEpoch() uint64 {
	e.c.Logger().Info("getting rewards per epoch")

	// Calculate total rewards for the epoch based on APR and staked amount
	rewardsPerBlock := uint64((float64(e.TotalStaked) * e.GetAPRForValidators() / 365 / 24 / 60 / 60) * (float64(e.EpochTracker.EpochLength) * 3)) // 3 seconds per block

	if e.TotalSupply+rewardsPerBlock > e.MaxSupply {
		rewardsPerBlock = e.MaxSupply - e.TotalSupply // Adjust to not exceed max supply
	}
	return rewardsPerBlock
}

// GetLastAcceptedBlockTimestamp retrieves the timestamp of the last accepted block from the VM.
func (e *Emission) GetLastAcceptedBlockTimestamp() time.Time {
	e.c.Logger().Info("fetching last accepted block timestamp")
	return e.nuklaivm.LastAcceptedBlock().Timestamp().UTC()
}

// GetLastAcceptedBlockHeight retrieves the height of the last accepted block from the VM.
func (e *Emission) GetLastAcceptedBlockHeight() uint64 {
	e.c.Logger().Info("fetching last accepted block height")
	return e.nuklaivm.LastAcceptedBlock().Height()
}
