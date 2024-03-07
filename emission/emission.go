// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/storage"
)

var (
	emission *Emission
	once     sync.Once
)

type Validator struct {
	NodeID                   ids.NodeID `json:"nodeID"`            // Node ID of the validator
	PublicKey                []byte     `json:"publicKey"`         // Public key of the validator
	StakedAmount             uint64     `json:"stakedAmount"`      // Total amount staked by the validator
	UnclaimedStakedReward    uint64     `json:"stakedReward"`      // Total rewards accumulated by the validator
	DelegationFeeRate        float64    `json:"delegationFeeRate"` // Fee rate for delegations
	DelegatedAmount          uint64     `json:"delegatedAmount"`   // Total amount delegated to the validator
	UnclaimedDelegatedReward uint64     `json:"delegatedReward"`   // Total rewards accumulated by the delegators

	delegatorsLastClaim map[codec.Address]uint64 // Map of delegator addresses to their last claim block height
}

type EmissionAccount struct {
	Address          codec.Address `json:"address"`
	UnclaimedBalance uint64        `json:"unclaimedBalance"`
}

type EpochTracker struct {
	BaseAPR        float64 `json:"baseAPR"`        // Base APR to use
	BaseValidators uint64  `json:"baseValidators"` // Base number of validators to use
	BlockCounter   uint64  `json:"blockCounter"`   // Tracks blocks since the last reward distribution
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

// GetRewardsPerEpoch calculates the rewards per epock based on the total staked amount
// and the APR for validators.
func (e *Emission) GetRewardsPerEpoch() uint64 {
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
func (e *Emission) CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address, currentBlockHeight uint64) (uint64, error) {
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

	// Calculate the number of blocks since the last claim
	blocksSinceLastClaim := currentBlockHeight - lastClaimHeight
	if blocksSinceLastClaim > e.EpochTracker.EpochLength {
		blocksSinceLastClaim = e.EpochTracker.EpochLength // Cap it at one epoch length to avoid overcalculation
	}

	// Calculate the delegator's share of the rewards
	delegatorShare := float64(userStakedAmount) / float64(validator.DelegatedAmount)

	// Define a durationFactor based on blocksSinceLastClaim relative to the
	// rewardEpochLength ensuring it does not exceed 1
	durationFactor := math.Min(1, float64(blocksSinceLastClaim)/float64(e.EpochTracker.EpochLength))

	// Calculate the maximum possible reward for the delegator
	maxPossibleReward := delegatorShare * float64(validator.UnclaimedDelegatedReward)

	// Adjust the total rewards for the delegator by the durationFactor
	// Ensuring the reward does not exceed maxPossibleReward
	delegationRewards := math.Min(maxPossibleReward, maxPossibleReward*durationFactor)

	return uint64(delegationRewards), nil
}

// RegisterValidatorStake adds a new validator to the heap with the specified staked amount
// and updates the total staked amount.
func (e *Emission) RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakedAmount, delegationFeeRate uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("registering validator stake")

	// Check if the validator was already registered
	if _, exists := e.validators[nodeID]; exists {
		return ErrValidatorAlreadyRegistered
	}

	validator := &Validator{
		NodeID:              nodeID,
		PublicKey:           bls.PublicKeyToBytes(nodePublicKey),
		StakedAmount:        stakedAmount,
		DelegationFeeRate:   float64(delegationFeeRate) / 100.0, // Convert to decimal
		delegatorsLastClaim: make(map[codec.Address]uint64),
	}
	e.validators[nodeID] = validator
	e.TotalStaked += stakedAmount

	return nil
}

// UnregisterValidatorStake removes a validator from the heap and updates the total
// staked amount accordingly.
func (e *Emission) UnregisterValidatorStake(nodeID ids.NodeID) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("unregistering validator stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}

	// TODO: maybe don't delete the validator because then, delegators may not be able to claim
	// their delegation rewards otherwise. Maybe market them as active/inactive
	e.TotalStaked -= (validator.StakedAmount + validator.DelegatedAmount)
	delete(e.validators, nodeID)

	return nil
}

// DelegateUserStake increases the delegated stake for a validator and rebalances the heap.
func (e *Emission) DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("delegating user stake")

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return ErrValidatorNotFound
	}

	// Check if the delegator was already staked
	if _, exists := validator.delegatorsLastClaim[delegatorAddress]; exists {
		return ErrDelegatorAlreadyStaked
	}

	// Update the validator's stake
	validator.DelegatedAmount += stakeAmount
	e.TotalStaked += stakeAmount

	// Update the delegator's stake
	validator.delegatorsLastClaim[delegatorAddress] = e.GetLastAcceptedBlockHeight()

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
	if _, exists := validator.delegatorsLastClaim[actor]; !exists {
		return 0, ErrDelegatorNotFound
	}

	// Claim rewards while undelegating
	rewardAmount := uint64(0)
	if actor == codec.EmptyAddress {
		// Validator claiming their rewards and resetting unclaimed rewards
		rewardAmount, validator.UnclaimedStakedReward = validator.UnclaimedStakedReward, 0
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

	// Update the validator's stake
	validator.DelegatedAmount -= stakeAmount
	e.TotalStaked -= stakeAmount

	// Update the delegator's stake
	delete(validator.delegatorsLastClaim, actor)

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
	} else {
		// Delegator claiming their rewards
		currentBlockHeight := e.GetLastAcceptedBlockHeight()
		reward, err := e.CalculateUserDelegationRewards(nodeID, actor, currentBlockHeight)
		if err != nil {
			return 0, err
		}
		utils.Outf("Emission----------------------BEFORE: validator.UnclaimedDelegatedReward:%d\n", validator.UnclaimedDelegatedReward)
		utils.Outf("Emission----------------------currentBlockHeight:%d\n", currentBlockHeight)
		utils.Outf("Emission----------------------reward:%d\n", reward)
		validator.delegatorsLastClaim[actor] = currentBlockHeight
		validator.UnclaimedDelegatedReward -= reward // Reset unclaimed rewards
		rewardAmount = reward
		utils.Outf("Emission----------------------AFTER: validator.UnclaimedDelegatedReward:%d\n", validator.UnclaimedDelegatedReward)
	}

	return rewardAmount, nil
}

// MintNewNAI calculates and distributes rewards to all the staked validators at the end of each
// reward epoch
func (e *Emission) MintNewNAI() uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.EpochTracker.BlockCounter++ // Increment block counter for each new block

	// Check if the current block is the end of an epoch
	if e.EpochTracker.BlockCounter >= e.EpochTracker.EpochLength {
		e.c.Logger().Info("minting new NAI tokens")

		// Calculate total rewards for the epoch based on APR and staked amount
		totalEpochRewards := e.GetRewardsPerEpoch()

		// Calculate rewards per unit staked to minimize iterations
		rewardsPerStakeUnit := float64(totalEpochRewards) / float64(e.TotalStaked)

		// Distribute rewards based on stake proportion
		for _, validator := range e.validators {
			validatorStake := validator.StakedAmount + validator.DelegatedAmount
			totalValidatorReward := uint64(float64(validatorStake) * rewardsPerStakeUnit)

			validatorReward, delegationReward := distributeValidatorRewards(totalValidatorReward, validator.DelegationFeeRate, validator.DelegatedAmount)
			// TODO: Ensure that the validator's stakeEndTime has not passed before adding the rewards
			validator.UnclaimedStakedReward += validatorReward
			validator.UnclaimedDelegatedReward += delegationReward
		}

		e.TotalStaked += totalEpochRewards // Update the total supply with the new minted rewards

		e.EpochTracker.BlockCounter = 0 // Reset block counter for the next epoch
		return totalEpochRewards        // Return the total rewards distributed in this epoch
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
		validatorStake := validator.StakedAmount + validator.DelegatedAmount
		totalValidatorFee := uint64(float64(validatorStake) * feesPerStakeUnit)

		validatorFee, delegationFee := distributeValidatorRewards(totalValidatorFee, validator.DelegationFeeRate, validator.DelegatedAmount)
		// TODO: Ensure that the validator's stakeEndTime has not passed before adding the rewards
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
			v.delegatorsLastClaim = stakedValidator[0].delegatorsLastClaim
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
