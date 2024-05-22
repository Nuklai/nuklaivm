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
	IsActive        bool   `json:"isActive"`     // Indicates if the delegator is currently active
	StakedAmount    uint64 `json:"stakedAmount"` // Total amount staked by the delegator
	stakeStartBlock uint64 // Start block of the stake
	stakeEndBlock   uint64 // End block of the stake
	alreadyClaimed  bool   // Indicates if the delegator has already claimed rewards
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
	Delegators               map[codec.Address]*Delegator

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

type Emission struct {
	c        Controller
	nuklaivm NuklaiVM

	TotalSupply     uint64          `json:"totalSupply"`     // Total supply of NAI
	MaxSupply       uint64          `json:"maxSupply"`       // Max supply of NAI
	EmissionAccount EmissionAccount `json:"emissionAccount"` // Emission Account Info

	validators  map[ids.NodeID]*Validator
	TotalStaked uint64 `json:"totalStaked"` // Total staked NAI

	EpochTracker EpochTracker `json:"epochTracker"` // Epoch Tracker Info

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

// GetNumDelegators returns the total number of delegators across all validators.
func (e *Emission) GetNumDelegators(nodeID ids.NodeID) int {
	e.c.Logger().Info("fetching total number of delegators")

	numDelegators := 0
	// Get delegators for all validators
	if nodeID == ids.EmptyNodeID {
		for _, validator := range e.validators {
			numDelegators += len(validator.Delegators)
		}
	} else {
		// Get delegators for a specific validator
		if validator, exists := e.validators[nodeID]; exists {
			numDelegators = len(validator.Delegators)
		}
	}

	return numDelegators
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
	if _, exists := validator.Delegators[actor]; !exists {
		return 0, ErrDelegatorNotFound
	}
	if validator.Delegators[actor].alreadyClaimed {
		return 0, nil // Rewards already claimed
	}

	// Iterate over each epoch within the stake period
	startEpoch := validator.Delegators[actor].stakeStartBlock / e.EpochTracker.EpochLength
	endEpoch := validator.Delegators[actor].stakeEndBlock / e.EpochTracker.EpochLength
	totalReward := uint64(0)

	for epoch := startEpoch; epoch < endEpoch; epoch++ {
		if reward, ok := validator.epochRewards[epoch]; ok {
			// Calculate reward for this epoch
			delegatorShare := float64(validator.Delegators[actor].StakedAmount) / float64(validator.DelegatedAmount)
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
		// If validator exists, it's a re-registration, update necessary fields
		validator.PublicKey = bls.PublicKeyToBytes(nodePublicKey)        // Update public key if needed
		validator.StakedAmount += stakedAmount                           // Adjust the staked amount
		validator.DelegationFeeRate = float64(delegationFeeRate) / 100.0 // Update delegation fee rate if needed
		validator.stakeStartBlock = stakeStartBlock
		validator.stakeEndBlock = stakeEndBlock
		// Note: We might want to keep some attributes unchanged, such as epochRewards, etc.
	} else {
		// If validator does not exist, create a new entry
		e.validators[nodeID] = &Validator{
			NodeID:            nodeID,
			PublicKey:         bls.PublicKeyToBytes(nodePublicKey),
			StakedAmount:      stakedAmount,
			DelegationFeeRate: float64(delegationFeeRate) / 100.0, // Convert to decimal
			Delegators:        make(map[codec.Address]*Delegator),
			epochRewards:      make(map[uint64]uint64),
			stakeStartBlock:   stakeStartBlock,
			stakeEndBlock:     stakeEndBlock,
		}
	}

	return nil
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

	if validator.IsActive {
		e.TotalStaked -= validator.StakedAmount
	}

	// Mark the validator as inactive
	validator.IsActive = false

	// If there are no more delegators, get the rewards and remove the validator
	if len(validator.Delegators) == 0 {
		rewardAmount += validator.UnclaimedDelegatedReward
		validator.UnclaimedDelegatedReward = 0
		e.TotalStaked -= validator.DelegatedAmount
		delete(e.validators, nodeID)
	}

	return rewardAmount, nil
}

// DelegateUserStake increases the delegated stake for a validator and rebalances the heap.
func (e *Emission) DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeStartBlock, stakeEndBlock, stakeAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("delegating user stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}

	// Check if the delegator was already staked
	if _, exists := validator.Delegators[delegatorAddress]; exists {
		return ErrDelegatorAlreadyStaked
	}

	// Add the delegator to the list
	validator.Delegators[delegatorAddress] = &Delegator{
		IsActive:        false, // this will be set to true automatically when stakeStartBlock is reached
		StakedAmount:    stakeAmount,
		stakeStartBlock: stakeStartBlock,
		stakeEndBlock:   stakeEndBlock,
		alreadyClaimed:  false,
	}

	return nil
}

// UndelegateUserStake decreases the delegated stake for a validator and rebalances the heap.
func (e *Emission) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address, stakeAmount uint64) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("undelegating user stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists
	if _, exists := validator.Delegators[actor]; !exists {
		return 0, ErrDelegatorNotFound
	}

	// Claim rewards while undelegating
	rewardAmount, err := e.CalculateUserDelegationRewards(nodeID, actor)
	if err != nil {
		return 0, err
	}
	validator.UnclaimedDelegatedReward -= rewardAmount // Reset unclaimed rewards

	delete(validator.Delegators, actor) // Remove the delegator from the list

	// If the validator is inactive and has no more delegators, remove the validator
	if !validator.IsActive && len(validator.Delegators) == 0 {
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
		if len(validator.Delegators) == 0 {
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
		validator.Delegators[actor].alreadyClaimed = true
		rewardAmount = reward
	}

	return rewardAmount, nil
}

func (e *Emission) MintNewNAI() uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	currentBlockHeight := e.GetLastAcceptedBlockHeight()

	// Check if the current block is the end of an epoch
	if currentBlockHeight%e.EpochTracker.EpochLength == 0 {
		e.c.Logger().Info("minting new NAI tokens at the end of the epoch")

		// Calculate total rewards for the epoch based on APR and staked amount
		totalEpochRewards := e.GetRewardsPerEpoch()

		// Calculate rewards per unit staked to minimize iterations
		rewardsPerStakeUnit := float64(0)
		if e.TotalStaked > 0 {
			rewardsPerStakeUnit = float64(totalEpochRewards) / float64(e.TotalStaked)
		}

		actualRewards := uint64(0)

		// Distribute rewards based on stake proportion
		for _, validator := range e.validators {
			lastBlockHeight := currentBlockHeight
			// Mark validator active based on if stakeStartBlock has started
			if lastBlockHeight > validator.stakeStartBlock && !validator.IsActive {
				validator.IsActive = true
				e.TotalStaked += (validator.StakedAmount + validator.DelegatedAmount)
			}
			if !validator.IsActive {
				continue
			}
			// Mark validator inactive based on if stakeEndBlock has ended
			if lastBlockHeight > validator.stakeEndBlock {
				validator.IsActive = false
				e.TotalStaked -= (validator.StakedAmount + validator.DelegatedAmount)
				continue
			}

			// Update the number of delegators if the current block height is greater than the delegator's stakeStartBlock
			for _, delegatorInfo := range validator.Delegators {
				if lastBlockHeight > delegatorInfo.stakeStartBlock && !delegatorInfo.IsActive {
					delegatorInfo.IsActive = true
					validator.DelegatedAmount += delegatorInfo.StakedAmount
					e.TotalStaked += delegatorInfo.StakedAmount
				}
				if !delegatorInfo.IsActive {
					continue
				}
				// Mark delegator inactive based on if stakeEndBlock has ended
				if lastBlockHeight > delegatorInfo.stakeEndBlock {
					delegatorInfo.IsActive = false
					validator.DelegatedAmount -= delegatorInfo.StakedAmount
					e.TotalStaked -= delegatorInfo.StakedAmount
				}
			}

			validatorStake := validator.StakedAmount + validator.DelegatedAmount
			totalValidatorReward := uint64(float64(validatorStake) * rewardsPerStakeUnit)

			// Calculate the rewards for the validator and for delegation
			validatorReward, delegationReward := distributeValidatorRewards(totalValidatorReward, validator.DelegationFeeRate, validator.DelegatedAmount)

			actualRewards += validatorReward + delegationReward

			// Update validator's and delegators' rewards
			validator.UnclaimedStakedReward += validatorReward
			validator.UnclaimedDelegatedReward += delegationReward

			// Track rewards per epoch for delegation
			epochNumber := currentBlockHeight / e.EpochTracker.EpochLength
			validator.epochRewards[epochNumber] = delegationReward
		}

		// Update the total supply with the new minted rewards
		e.TotalSupply += actualRewards

		// Return the total rewards distributed in this epoch
		return actualRewards
	}

	// No rewards are distributed until the end of the epoch
	return 0
}

// DistributeFees allocates transaction fees between the emission account and validators,
// based on the total staked amount.
func (e *Emission) DistributeFees(fee uint64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("distributing transaction fees")

	if e.TotalSupply+fee > e.MaxSupply {
		fee = e.MaxSupply - e.TotalSupply // Adjust to not exceed max supply
	}

	// Give 50% fees to Emission Account
	feesForEmission := fee / 2
	e.EmissionAccount.UnclaimedBalance += feesForEmission

	// Give remaining to Validators
	feesForValidators := fee - feesForEmission
	if e.TotalStaked == 0 || feesForValidators == 0 {
		return // No validators or no fees to distribute
	}

	// Calculate fees per unit staked to minimize iterations
	feesPerStakeUnit := float64(feesForValidators) / float64(e.TotalStaked)

	// Distribute fees based on stake proportion
	for _, validator := range e.validators {
		lastBlockHeight := e.GetLastAcceptedBlockHeight()
		// Mark validator active based on if stakeStartBlock has started
		if lastBlockHeight > validator.stakeStartBlock && !validator.IsActive {
			validator.IsActive = true
			e.TotalStaked += (validator.StakedAmount + validator.DelegatedAmount)
		}
		if !validator.IsActive {
			continue
		}
		// Mark validator inactive based on if stakeEndBlock has ended
		if lastBlockHeight > validator.stakeEndBlock {
			validator.IsActive = false
			e.TotalStaked -= (validator.StakedAmount + validator.DelegatedAmount)
			continue
		}

		// Update the number of delegators if the current block height is greater than the delegator's stakeStartBlock
		for _, delegatorInfo := range validator.Delegators {
			if lastBlockHeight > delegatorInfo.stakeStartBlock && !delegatorInfo.IsActive {
				delegatorInfo.IsActive = true
				validator.DelegatedAmount += delegatorInfo.StakedAmount
				e.TotalStaked += delegatorInfo.StakedAmount
			}
			if !delegatorInfo.IsActive {
				continue
			}
			// Mark delegator inactive based on if stakeEndBlock has ended
			if lastBlockHeight > delegatorInfo.stakeEndBlock {
				delegatorInfo.IsActive = false
				validator.DelegatedAmount -= delegatorInfo.StakedAmount
				e.TotalStaked -= delegatorInfo.StakedAmount
			}
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
			v.epochRewards = stakedValidator[0].epochRewards
		}
		validators = append(validators, &v)
	}
	return validators
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
