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
	"github.com/ava-labs/hypersdk/state"
	"github.com/nuklai/nuklaivm/storage"
)

var _ Tracker = (*Manual)(nil)

type Manual struct {
	c        Controller
	nuklaivm NuklaiVM

	TotalSupply     uint64          `json:"totalSupply"`     // Total supply of NAI
	MaxSupply       uint64          `json:"maxSupply"`       // Max supply of NAI
	EmissionAccount EmissionAccount `json:"emissionAccount"` // Emission Account Info

	validators        map[ids.NodeID]*Validator
	CurrentValidators []*Validator
	TotalStaked       uint64 `json:"totalStaked"` // Total staked NAI

	EpochTracker EpochTracker `json:"epochTracker"` // Epoch Tracker Info

	lock sync.RWMutex
}

// New initializes the Emission struct with initial parameters and sets up the validators heap
// and indices map.
func NewManual(c Controller, vm NuklaiVM, totalSupply, maxSupply uint64, emissionAddress codec.Address) *Manual {
	once.Do(func() {
		c.Logger().Info("Initializing emission with max supply and rewards per block settings")

		if maxSupply == 0 {
			maxSupply = GetStakingConfig().RewardConfig.SupplyCap // Use the staking config's supply cap if maxSupply is not specified
		}

		emission = &Manual{ // Create the Emission instance with initialized values
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
	return emission.(*Manual)
}

// AddToTotalSupply increases the total supply of NAI by a specified amount, ensuring it
// does not exceed the max supply.
func (e *Manual) AddToTotalSupply(amount uint64) uint64 {
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
func (e *Manual) GetNumDelegators(nodeID ids.NodeID) int {
	e.c.Logger().Info("fetching total number of delegators")

	numDelegators := 0
	// Get delegators for all validators
	if nodeID == ids.EmptyNodeID {
		for _, validator := range e.validators {
			numDelegators += len(validator.delegatorsLastClaim)
		}
	} else {
		// Get delegators for a specific validator
		if validator, exists := e.validators[nodeID]; exists {
			numDelegators = len(validator.delegatorsLastClaim)
		}
	}

	return numDelegators
}

// GetAPRForValidators calculates the Annual Percentage Rate (APR) for validators
// based on the number of validators.
func (e *Manual) GetAPRForValidators() float64 {
	e.c.Logger().Info("getting APR for validators")

	apr := e.EpochTracker.BaseAPR // APR is expressed per year as a decimal, e.g., 0.25 for 25%
	// Beyond baseValidators, APR decreases proportionately
	baseValidators := int(e.EpochTracker.BaseValidators)
	if len(e.validators) > baseValidators {
		apr /= float64(len(e.validators)) / float64(baseValidators)
	}
	return apr
}

// GetRewardsPerEpoch calculates the rewards per epock based on the total staked amount
// and the APR for validators.
func (e *Manual) GetRewardsPerEpoch() uint64 {
	e.c.Logger().Info("getting rewards per epock")

	// Calculate total rewards for the epoch based on APR and staked amount
	rewardsPerBlock := uint64((float64(e.TotalStaked) * e.GetAPRForValidators() / 365 / 24 / 60 / 60) * (float64(e.EpochTracker.EpochLength) * 3)) // 3 seconds per block

	if e.TotalSupply+rewardsPerBlock > e.MaxSupply {
		rewardsPerBlock = e.MaxSupply - e.TotalSupply // Adjust to not exceed max supply
	}
	return rewardsPerBlock
}

// CalculateUserDelegationRewards computes the rewards for a user's delegated stake to a
// validator, factoring in the delegation duration and amount.
func (e *Manual) CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address, currentBlockHeight uint64) (uint64, error) {
	e.c.Logger().Info("calculating rewards for user delegation")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists
	lastClaimHeight, exists := validator.delegatorsLastClaim[actor]
	if !exists {
		return 0, ErrDelegatorNotFound
	}

	stateDB, err := e.nuklaivm.State()
	if err != nil {
		return 0, err
	}
	mu := state.NewSimpleMutable(stateDB)

	// Get user's delegation stake info
	exists, _, userStakedAmount, _, _, _ := storage.GetDelegateUserStake(context.TODO(), mu, actor, nodeID)
	if !exists {
		return 0, ErrStakeNotFound
	}

	// Iterate over each epoch since the last claim
	startEpoch := lastClaimHeight / e.EpochTracker.EpochLength
	endEpoch := currentBlockHeight / e.EpochTracker.EpochLength
	totalReward := uint64(0)

	for epoch := startEpoch; epoch < endEpoch; epoch++ {
		if reward, ok := validator.epochRewards[epoch]; ok {
			// Calculate reward for this epoch
			delegatorShare := float64(userStakedAmount) / float64(validator.DelegatedAmount)
			epochReward := delegatorShare * float64(reward)
			totalReward += uint64(epochReward)
		}
	}

	return totalReward, nil
}

// RegisterValidatorStake adds a new validator to the heap with the specified staked amount
// and updates the total staked amount.
func (e *Manual) RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate uint64) error {
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
		// Note: We might want to keep some attributes unchanged, such as delegatorsLastClaim, epochRewards, etc.
	} else {
		// If validator does not exist, create a new entry
		e.validators[nodeID] = &Validator{
			NodeID:              nodeID,
			PublicKey:           bls.PublicKeyToBytes(nodePublicKey),
			StakedAmount:        stakedAmount,
			DelegationFeeRate:   float64(delegationFeeRate) / 100.0, // Convert to decimal
			delegatorsLastClaim: make(map[codec.Address]uint64),
			epochRewards:        make(map[uint64]uint64),
			stakeStartBlock:     stakeStartBlock,
			stakeEndBlock:       stakeEndBlock,
		}
	}

	return nil
}

// WithdrawValidatorStake removes a validator from the heap and updates the total
// staked amount accordingly.
func (e *Manual) WithdrawValidatorStake(nodeID ids.NodeID) (uint64, error) {
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
	if len(validator.delegatorsLastClaim) == 0 {
		rewardAmount += validator.UnclaimedDelegatedReward
		validator.UnclaimedDelegatedReward = 0
		e.TotalStaked -= validator.DelegatedAmount
		delete(e.validators, nodeID)
	}

	return rewardAmount, nil
}

// DelegateUserStake increases the delegated stake for a validator and rebalances the heap.
func (e *Manual) DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeStartBlock, stakeAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("delegating user stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}
	// Check if the delegator was already staked   ERROR HERE ?
	if _, exists := validator.delegatorsLastClaim[delegatorAddress]; exists {
		return ErrDelegatorAlreadyStaked
	}
	// Update the validator's stake
	validator.DelegatedAmount += stakeAmount

	// We only add to total staked amount if the validator is active
	// If validator is inactive, we subtract from the total during distributeFees and mintNewNai functions
	// This will prevent us from adding to the total staked amount twice
	if validator.IsActive {
		e.TotalStaked += stakeAmount
	}

	// Update the delegator's stake
	validator.delegatorsLastClaim[delegatorAddress] = e.GetLastAcceptedBlockHeight()

	return nil
}

// UndelegateUserStake decreases the delegated stake for a validator and rebalances the heap.
func (e *Manual) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address, stakeAmount uint64) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("undelegating user stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists
	if _, exists := validator.delegatorsLastClaim[actor]; !exists {
		return 0, ErrDelegatorNotFound
	}

	// Claim rewards while undelegating
	currentBlockHeight := e.GetLastAcceptedBlockHeight()
	rewardAmount, err := e.CalculateUserDelegationRewards(nodeID, actor, currentBlockHeight)
	if err != nil {
		return 0, err
	}
	validator.delegatorsLastClaim[actor] = currentBlockHeight
	validator.UnclaimedDelegatedReward -= rewardAmount // Reset unclaimed rewards

	// Update the validator's stake
	validator.DelegatedAmount -= stakeAmount
	// We only subtract from total staked amount if the validator is active
	// If validator is inactive, we subtract from the total during distributeFees and mintNewNai functions
	// This will prevent us from adding to the total staked amount twice
	if validator.IsActive {
		e.TotalStaked -= stakeAmount
	}

	// Remove the delegator's entry
	delete(validator.delegatorsLastClaim, actor)

	// If the validator is inactive and has no more delegators, remove the validator
	if !validator.IsActive && len(validator.delegatorsLastClaim) == 0 {
		delete(e.validators, nodeID)
	}

	return rewardAmount, nil
}

// ClaimStakingRewards lets validators and delegators claim their rewards
func (e *Manual) ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
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
		if len(validator.delegatorsLastClaim) == 0 {
			rewardAmount += validator.UnclaimedDelegatedReward
			validator.UnclaimedDelegatedReward = 0
		}
	} else {
		// Delegator claiming their rewards
		currentBlockHeight := e.GetLastAcceptedBlockHeight()
		reward, err := e.CalculateUserDelegationRewards(nodeID, actor, currentBlockHeight)
		if err != nil {
			return 0, err
		}
		validator.delegatorsLastClaim[actor] = currentBlockHeight
		validator.UnclaimedDelegatedReward -= reward // Reset unclaimed rewards
		rewardAmount = reward
	}

	return rewardAmount, nil
}

func (e *Manual) MintNewNAI() uint64 {
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
			lastBlockHeight := e.GetLastAcceptedBlockHeight()
			// Mark validator active based on if stakeStartTime has started
			if lastBlockHeight > validator.stakeStartBlock {
				validator.IsActive = true
				e.TotalStaked += (validator.StakedAmount + validator.DelegatedAmount)
			}
			if !validator.IsActive {
				continue
			}
			// Mark validator inactive based on if stakeEndTime has ended
			if lastBlockHeight > validator.stakeEndBlock {
				validator.IsActive = false
				e.TotalStaked -= (validator.StakedAmount + validator.DelegatedAmount)
				continue
			}

			validatorStake := validator.StakedAmount + validator.DelegatedAmount
			totalValidatorReward := uint64(float64(validatorStake) * rewardsPerStakeUnit)

			// Calculate the rewards for the validator and for delegation
			validatorReward, delegationReward := uint64(0), uint64(0)
			if len(validator.delegatorsLastClaim) > 0 {
				validatorReward, delegationReward = distributeValidatorRewards(totalValidatorReward, validator.DelegationFeeRate, validator.DelegatedAmount)
			}

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
func (e *Manual) DistributeFees(fee uint64) {
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
		// Mark validator active based on if stakeStartTime has started
		if lastBlockHeight > validator.stakeStartBlock {
			validator.IsActive = true
			e.TotalStaked += (validator.StakedAmount + validator.DelegatedAmount)
		}
		if !validator.IsActive {
			continue
		}
		// Mark validator inactive based on if stakeEndTime has ended
		if lastBlockHeight > validator.stakeEndBlock {
			validator.IsActive = false
			e.TotalStaked -= (validator.StakedAmount + validator.DelegatedAmount)
			continue
		}

		validatorStake := validator.StakedAmount + validator.DelegatedAmount
		totalValidatorFee := uint64(float64(validatorStake) * feesPerStakeUnit)

		validatorFee, delegationFee := uint64(0), uint64(0)
		if len(validator.delegatorsLastClaim) > 0 {
			validatorFee, delegationFee = distributeValidatorRewards(totalValidatorFee, validator.DelegationFeeRate, validator.DelegatedAmount)
		}
		validator.UnclaimedStakedReward += validatorFee
		validator.UnclaimedDelegatedReward += delegationFee
	}
}

// GetStakedValidator retrieves the details of a specific validator by their NodeID.
func (e *Manual) GetStakedValidator(nodeID ids.NodeID) []*Validator {
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
func (e *Manual) GetAllValidators(ctx context.Context) []*Validator {
	e.c.Logger().Info("fetching all staked and unstaked validators")

	for _, v := range e.CurrentValidators {
		stakedValidator := e.GetStakedValidator(v.NodeID)
		if len(stakedValidator) > 0 {
			v.StakedAmount = stakedValidator[0].StakedAmount
			v.UnclaimedStakedReward = stakedValidator[0].UnclaimedStakedReward
			v.DelegationFeeRate = stakedValidator[0].DelegationFeeRate
			v.DelegatedAmount = stakedValidator[0].DelegatedAmount
			v.UnclaimedDelegatedReward = stakedValidator[0].UnclaimedDelegatedReward
			v.delegatorsLastClaim = stakedValidator[0].delegatorsLastClaim
		}
	}
	return e.CurrentValidators
}

// GetLastAcceptedBlockTimestamp retrieves the timestamp of the last accepted block from the VM.
func (e *Manual) GetLastAcceptedBlockTimestamp() time.Time {
	e.c.Logger().Info("fetching last accepted block timestamp")
	return e.nuklaivm.LastAcceptedBlock().Timestamp().UTC()
}

// GetLastAcceptedBlockHeight retrieves the height of the last accepted block from the VM.
func (e *Manual) GetLastAcceptedBlockHeight() uint64 {
	e.c.Logger().Info("fetching last accepted block height")
	return e.nuklaivm.LastAcceptedBlock().Height()
}

func (e *Manual) GetEmissionValidators() map[ids.NodeID]*Validator {
	e.c.Logger().Info("fetching emission validators")
	return e.validators
}

func (e *Manual) GetInfo() (emissionAccount EmissionAccount, totalSupply uint64, maxSupply uint64, totalStaked uint64, epochTracker EpochTracker) {
	return e.EmissionAccount, e.TotalSupply, e.MaxSupply, e.TotalStaked, e.EpochTracker
}
