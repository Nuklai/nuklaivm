// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"github.com/nuklai/nuklaivm/storage"
	"go.uber.org/zap"
)

const (
	baseAPR        = 0.25 // 25% APR
	baseValidators = 100  // Number of validators to get base APR
)

var (
	emission *Emission
	once     sync.Once
)

type Validator struct {
	NodeID                ids.NodeID `json:"nodeID"`          // Node ID of the validator
	PublicKey             string     `json:"publicKey"`       // Public key of the validator
	StakedAmount          uint64     `json:"stakedAmount"`    // Total amount staked by the validator
	DelegatedStake        uint64     `json:"delegatedStake"`  // Total number of user delegations to the validator
	DelegatedAmount       uint64     `json:"delegatedAmount"` // Total amount delegated to the validator
	UnclaimedStakedReward uint64     `json:"stakedReward"`    // Total rewards accumulated by the validator
}

type EmissionAccount struct {
	Address          codec.Address `json:"address"`
	UnclaimedBalance uint64        `json:"unclaimedBalance"`
}

type Emission struct {
	c        Controller
	nuklaivm NuklaiVM

	totalSupply     uint64
	maxSupply       uint64
	emissionAccount EmissionAccount

	validators       ValidatorHeap
	validatorIndices map[ids.NodeID]int // Map to keep track of each validator's index in the heap
	totalStaked      uint64

	lock sync.RWMutex
}

// New initializes the Emission struct with initial parameters and sets up the validators heap
// and indices map.
func New(c Controller, vm NuklaiVM, totalSupply, maxSupply uint64, emissionAddress codec.Address) *Emission {
	once.Do(func() {
		c.Logger().Info("Initializing emission with max supply and rewards per block settings")

		validatorsHeap := make(ValidatorHeap, 0)
		heap.Init(&validatorsHeap) // Initialize an empty heap for validators

		validatorIndices := make(map[ids.NodeID]int) // Map to track validator indices within the heap

		emissionAccount := &EmissionAccount{ // Setup the emission account with the provided address
			Address:          emissionAddress,
			UnclaimedBalance: 0,
		}

		if maxSupply == 0 {
			maxSupply = GetStakingConfig().RewardConfig.SupplyCap // Use the staking config's supply cap if maxSupply is not specified
		}

		emission = &Emission{ // Create the Emission instance with initialized values
			c:                c,
			nuklaivm:         vm,
			totalSupply:      totalSupply,
			maxSupply:        maxSupply,
			emissionAccount:  *emissionAccount,
			validators:       validatorsHeap,
			validatorIndices: validatorIndices,
			totalStaked:      0,
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

	e.c.Logger().Info("Adding to the total supply of NAI")
	if e.totalSupply+amount > e.maxSupply {
		amount = e.maxSupply - e.totalSupply // Adjust to not exceed max supply
	}
	e.totalSupply += amount
	return e.totalSupply
}

// GetTotalSupply returns the current total supply of NAI.
func (e *Emission) GetTotalSupply() uint64 {
	e.c.Logger().Info("fetching total supply of NAI")
	return e.totalSupply
}

// GetMaxSupply returns the maximum supply limit for NAI.
func (e *Emission) GetMaxSupply() uint64 {
	e.c.Logger().Info("fetching max supply of NAI")
	return e.maxSupply
}

// GetEmissionAccount returns the emission account details including its address and
// unclaimed balance.
func (e *Emission) GetEmissionAccount() *EmissionAccount {
	e.c.Logger().Info("fetching Emission Account")
	return &e.emissionAccount
}

// GetTotalStaked returns the total amount of NAI staked across all validators.
func (e *Emission) GetTotalStaked() uint64 {
	e.c.Logger().Info("fetching total NAI staked")
	return e.totalStaked
}

// GetAPRForValidators calculates the Annual Percentage Rate (APR) for validators
// based on the number of validators.
func (e *Emission) GetAPRForValidators() float64 {
	e.lock.RLock()
	defer e.lock.RUnlock()

	apr := baseAPR // APR is expressed per year as a decimal, e.g., 0.25 for 25%
	// Beyond baseValidators, APR decreases proportionately
	if len(e.validators) > baseValidators {
		apr /= float64(len(e.validators)) / float64(baseValidators)
	}
	return apr
}

// CalculateAnnualRewards computes the annual rewards based on the total staked amount and the APR.
func (*Emission) CalculateAnnualRewards(totalStaked uint64, apr float64) uint64 {
	blocksPerYear := GetStakingConfig().RewardConfig.MintingPeriod.Seconds() / 5 // Block time is assumed to be 5 seconds
	totalAnnualRewards := float64(totalStaked) * apr
	rewardsPerBlock := totalAnnualRewards / blocksPerYear
	return uint64(rewardsPerBlock)
}

// GetRewardsPerBlock calculates the rewards distributed per block based on the current APR and
// total staked amount.
func (e *Emission) GetRewardsPerBlock() uint64 {
	e.c.Logger().Info("fetching amount of NAI rewards per block")
	apr := e.GetAPRForValidators()
	return e.CalculateAnnualRewards(e.totalStaked, apr)
}

// CalculateValidatorRewards computes the rewards for a specific validator and their
// delegators based on the staked amount.
func (e *Emission) CalculateValidatorRewards(nodeID ids.NodeID) (uint64, uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Find the validator in the heap
	index, exists := e.validatorIndices[nodeID]
	if !exists {
		return 0, 0, ErrValidatorNotFound
	}
	validator := e.validators[index].validator

	// Retrieve the vm state
	stateDB, err := e.nuklaivm.State()
	if err != nil {
		return 0, 0, err
	}
	mu := state.NewSimpleMutable(stateDB)
	exists, validatorStakeStartTime, _, _, delegationFeeRate, _, _, _ := storage.GetRegisterValidatorStake(context.TODO(), mu, nodeID)
	if !exists {
		return 0, 0, ErrValidatorNotFound
	}

	// Determine APR based on the number of validators
	apr := e.GetAPRForValidators()
	annualizedReward := e.CalculateAnnualRewards(validator.StakedAmount+validator.DelegatedAmount, apr)

	// Calculate the staking duration in seconds
	stakeStartTime, stakeEndTime := time.Unix(int64(validatorStakeStartTime), 0).UTC(), e.GetLastAcceptedBlockTimestamp()
	stakingDuration := stakeEndTime.Sub(stakeStartTime).Seconds()
	if stakingDuration < 0 {
		stakingDuration = 0 // Ensure staking duration is non-negative
	}

	// Calculate the pro-rata annualized staking rewards (scaled by 10^9 for NAI decimals)
	mintingPeriodSeconds := GetStakingConfig().RewardConfig.MintingPeriod.Seconds()
	stakingRewards := float64(validator.UnclaimedStakedReward) + ((float64(annualizedReward) / mintingPeriodSeconds) * stakingDuration)

	// Calculate validator's net rewards after delegation fee (scaled by 10^9)
	delegationRewards := stakingRewards * (float64(delegationFeeRate) / 100.0)
	validatorRewards := stakingRewards - delegationRewards

	// Convert net rewards to uint64 for return, scaling down from 10^9
	return uint64(validatorRewards), uint64(delegationRewards), nil
}

// CalculateUserDelegationRewards computes the rewards for a user's delegated stake to a
// validator, factoring in the delegation duration and amount.
func (e *Emission) CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Find the validator in the heap
	index, exists := e.validatorIndices[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}
	validator := e.validators[index].validator

	stateDB, err := e.nuklaivm.State()
	if err != nil {
		return 0, err
	}
	mu := state.NewSimpleMutable(stateDB)

	// Get user's delegation stake info
	exists, userStakeStartTime, userStakedAmount, _, _, _ := storage.GetDelegateUserStake(context.TODO(), mu, actor, nodeID)
	if !exists {
		return 0, ErrStakeNotFound
	}

	_, delegationRewards, err := e.CalculateValidatorRewards(nodeID)
	if err != nil {
		return 0, err
	}

	// Calculate staking duration in seconds
	now := e.GetLastAcceptedBlockTimestamp()
	startTime := time.Unix(int64(userStakeStartTime), 0).UTC()
	stakeDuration := now.Sub(startTime)
	if stakeDuration <= 0 {
		return 0, ErrInvalidStakeDuration
	}

	// Calculate user's share of delegation rewards based on their stake amount
	totalDelegatedStake := validator.DelegatedAmount
	userRewardShare := float64(userStakedAmount) / float64(totalDelegatedStake)

	// Longer staking duration means higher rewards
	// For simplicity, we are using a linear relationship: rewards increase linearly with staking //
	// duration
	durationFactor := float64(stakeDuration) / float64(86400) // Normalize by the number of seconds in a day for a daily reward increase
	adjustedUserRewardShare := userRewardShare * durationFactor

	userDelegationRewards := uint64(float64(delegationRewards) * adjustedUserRewardShare)

	return userDelegationRewards, nil
}

// RegisterValidatorStake adds a new validator to the heap with the specified staked amount
// and updates the total staked amount.
func (e *Emission) RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey string, stakedAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Check if the validator was already registered
	if _, exists := e.validatorIndices[nodeID]; exists {
		return ErrValidatorAlreadyRegistered
	}

	validator := &Validator{
		NodeID:                nodeID,
		PublicKey:             nodePublicKey,
		StakedAmount:          stakedAmount,
		DelegatedStake:        0,
		DelegatedAmount:       0,
		UnclaimedStakedReward: 0,
	}
	item := &ValidatorHeapItem{validator: validator}
	heap.Push(&e.validators, item)
	e.validatorIndices[nodeID] = item.index
	e.totalStaked += stakedAmount

	return nil
}

// UnregisterValidatorStake removes a validator from the heap and updates the total
// staked amount accordingly.
func (e *Emission) UnregisterValidatorStake(nodeID ids.NodeID) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Find the validator in the heap
	index, exists := e.validatorIndices[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}

	e.totalStaked -= e.validators[index].validator.StakedAmount
	heap.Remove(&e.validators, index)
	delete(e.validatorIndices, nodeID)

	return nil
}

// DelegateUserStake increases the delegated stake for a validator and rebalances the heap.
func (e *Emission) DelegateUserStake(nodeID ids.NodeID, stakeAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Find the validator in the heap
	index, exists := e.validatorIndices[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}
	validatorItem := e.validators[index]

	// Update the validator's stake
	validatorItem.validator.DelegatedStake++
	validatorItem.validator.DelegatedAmount += stakeAmount

	// Rebalance the heap with the updated validator stake
	e.validators.update(validatorItem, validatorItem.validator)

	return nil
}

// UndelegateUserStake decreases the delegated stake for a validator and rebalances the heap.
func (e *Emission) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address, stakeAmount uint64) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Find the validator in the heap
	index, ok := e.validatorIndices[nodeID]
	if !ok {
		return 0, ErrValidatorNotFound
	}
	validatorItem := e.validators[index]

	// Update the validator's stake
	validatorItem.validator.DelegatedStake--
	validatorItem.validator.DelegatedAmount -= stakeAmount

	// Rebalance the heap with the updated validator stake
	e.validators.update(validatorItem, validatorItem.validator)

	// Claim rewards while undelegating
	rewardAmount, err := e.ClaimStakingRewards(nodeID, actor)
	if err != nil {
		return 0, err
	}

	return rewardAmount, nil
}

// ClaimStakingRewards lets validators and delegators claim their rewards
func (e *Emission) ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Find the validator in the heap
	index, ok := e.validatorIndices[nodeID]
	if !ok {
		return 0, ErrValidatorNotFound
	}
	validatorItem := e.validators[index]

	rewardAmount := uint64(0)
	if actor != codec.EmptyAddress {
		// For a delegator claiming their rewards
		reward, err := e.CalculateUserDelegationRewards(nodeID, actor)
		if err != nil {
			return 0, err
		}
		rewardAmount = reward
	} else {
		// For a validator claiming their rewards
		reward, _, err := e.CalculateValidatorRewards(nodeID)
		if err != nil {
			return 0, err
		}
		rewardAmount = reward
	}

	if rewardAmount > validatorItem.validator.UnclaimedStakedReward {
		return 0, ErrInsufficientRewards
	}

	// Update the validator's unclaimed reward
	validatorItem.validator.UnclaimedStakedReward -= rewardAmount

	return rewardAmount, nil
}

// GetStakedValidator retrieves the details of a specific validator by their NodeID.
func (e *Emission) GetStakedValidator(nodeID ids.NodeID) []*Validator {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if nodeID == ids.EmptyNodeID {
		return e.getAllValidators()
	}

	// Find the validator in the heap
	index, exists := e.validatorIndices[nodeID]
	if !exists {
		return []*Validator{}
	}

	return []*Validator{e.validators[index].validator}
}

// getAllValidators provides a list of all validators currently in the heap.
func (e *Emission) getAllValidators() []*Validator {
	e.lock.RLock()
	defer e.lock.RUnlock()

	validators := make([]*Validator, 0, len(e.validators))
	for _, item := range e.validators {
		validators = append(validators, item.validator)
	}
	return validators
}

// GetAllValidators fetches the current validators from the underlying VM
func (e *Emission) GetNuklaiVMValidators(ctx context.Context) (map[ids.NodeID]*validators.GetValidatorOutput, map[string]struct{}) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.nuklaivm.CurrentValidators(ctx)
}

// GetLastAcceptedBlockTimestamp retrieves the timestamp of the last accepted block from the VM.
func (e *Emission) GetLastAcceptedBlockTimestamp() time.Time {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.nuklaivm.LastAcceptedBlock().Timestamp().UTC()
}

// MintNewNAI mints new NAI tokens based on the rewards per block,
// distributing them among validators and the emission account as necessary.
func (e *Emission) MintNewNAI() uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	initialTotalRewards := e.GetRewardsPerBlock()
	totalRewards := initialTotalRewards
	if e.totalSupply+totalRewards > e.maxSupply {
		totalRewards = e.maxSupply - e.totalSupply // Adjust to not exceed max supply
	}
	if totalRewards == 0 {
		return 0 // Nothing to mint
	}

	// No validators to distribute rewards to if totalStaked is 0
	// So, give all the rewards to Emission Account
	if e.totalStaked == 0 {
		e.emissionAccount.UnclaimedBalance += totalRewards
		return totalRewards
	}

	// Distribute rewards based on stake proportion
	for _, validatorItem := range e.validators {
		validatorReward, _, err := e.CalculateValidatorRewards(validatorItem.validator.NodeID)
		if err != nil {
			e.c.Logger().Error("error calculating validator rewards: ", zap.Error(err))
			continue
		}

		validatorItem.validator.UnclaimedStakedReward += validatorReward
		totalRewards -= validatorReward
		if totalRewards <= 0 {
			break // Stop if we've allocated all the rewards
		}
	}

	// Any remaining rewards go to the emission account
	if totalRewards > 0 {
		e.emissionAccount.UnclaimedBalance += totalRewards
	}

	return initialTotalRewards
}

// DistributeFees allocates transaction fees between the emission account and validators,
// based on the total staked amount.
func (e *Emission) DistributeFees(fee uint64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.totalSupply+fee > e.maxSupply {
		fee = e.maxSupply - e.totalSupply // Adjust to not exceed max supply
	}

	// Give 50% fees to Emission Account
	feesForEmission := fee / 2
	e.emissionAccount.UnclaimedBalance += feesForEmission

	// Give remaining to Validators
	feesForValidators := fee - feesForEmission

	if e.totalStaked > 0 {
		// Distribute rewards based on stake proportion
		for _, validatorItem := range e.validators {
			validatorShare := float64(validatorItem.validator.StakedAmount) / float64(e.totalStaked)
			validatorFee := uint64(float64(feesForValidators) * validatorShare)
			validatorItem.validator.UnclaimedStakedReward += validatorFee
		}
	}
}
