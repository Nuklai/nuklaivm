package emission

import (
	"context"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/utils"
	"go.uber.org/zap"
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

type Emission struct {
	c        Controller
	nuklaivm NuklaiVM

	TotalSupply     uint64          `json:"totalSupply"`     // Total supply of NAI
	MaxSupply       uint64          `json:"maxSupply"`       // Max supply of NAI
	EmissionAccount EmissionAccount `json:"emissionAccount"` // Emission Account Info

	validators  map[ids.NodeID]*Validator
	TotalStaked uint64 `json:"totalStaked"` // Total staked NAI

	EpochTracker EpochTracker `json:"epochTracker"` // Epoch Tracker Info

	activationEvents            map[uint64][]*Validator
	deactivationEvents          map[uint64][]*Validator
	delegatorActivationEvents   map[uint64][]*DelegatorEvent
	delegatorDeactivationEvents map[uint64][]*DelegatorEvent

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
			activationEvents:            make(map[uint64][]*Validator),
			deactivationEvents:          make(map[uint64][]*Validator),
			delegatorActivationEvents:   make(map[uint64][]*DelegatorEvent),
			delegatorDeactivationEvents: make(map[uint64][]*DelegatorEvent),
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
	e.c.Logger().Info("calculating rewards for user delegation",
		zap.String("nodeID", nodeID.String()),
	)

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		e.c.Logger().Error("validator not found", zap.String("nodeID", nodeID.String()))
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists
	delegator, exists := validator.delegators[actor]
	if !exists {
		e.c.Logger().Error("delegator not found")
		return 0, ErrDelegatorNotFound
	}

	// Calculate the total rewards for the delegator proportionally
	startEpoch := delegator.StakeStartBlock / e.EpochTracker.EpochLength
	endEpoch := delegator.StakeEndBlock / e.EpochTracker.EpochLength
	totalReward := uint64(0)

	e.c.Logger().Info("delegator details",
		zap.Uint64("startEpoch", startEpoch),
		zap.Uint64("endEpoch", endEpoch),
	)

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		reward, rewardExists := validator.epochRewards[epoch]
		delegatedAmountForEpoch, amountExists := validator.delegatedAmountPerEpoch[epoch]

		e.c.Logger().Info("epoch details",
			zap.Uint64("epoch", epoch),
			zap.Bool("rewardExists", rewardExists),
			zap.Uint64("reward", reward),
			zap.Bool("amountExists", amountExists),
			zap.Uint64("delegatedAmountForEpoch", delegatedAmountForEpoch),
		)

		if rewardExists && amountExists && delegatedAmountForEpoch > 0 {
			// Calculate the reward proportion for this epoch
			delegatorShare := float64(delegator.StakedAmount) / float64(delegatedAmountForEpoch)
			epochReward := uint64(float64(reward) * delegatorShare)
			totalReward += epochReward

			e.c.Logger().Info("epoch reward",
				zap.Uint64("epoch", epoch),
				zap.Uint64("epochReward", epochReward),
				zap.Uint64("totalReward", totalReward),
			)
		}
	}

	e.c.Logger().Info("total delegation reward", zap.Uint64("totalReward", totalReward))
	return totalReward, nil
}

// RegisterValidatorStake adds a new validator to the heap with the specified staked amount
// and updates the total staked amount.
func (e *Emission) RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("registering validator stake")

	// Check if the validator was already registered
	validator, exists := e.validators[nodeID]
	if exists {
		// Preserve existing rewards and data
		validator.PublicKey = bls.PublicKeyToBytes(nodePublicKey)
		validator.StakedAmount = stakedAmount
		validator.DelegationFeeRate = float64(delegationFeeRate) / 100.0
		validator.stakeStartBlock = stakeStartBlock
		validator.stakeEndBlock = stakeEndBlock
	} else {
		// Create a new validator
		validator = &Validator{
			NodeID:                  nodeID,
			PublicKey:               bls.PublicKeyToBytes(nodePublicKey),
			StakedAmount:            stakedAmount,
			DelegationFeeRate:       float64(delegationFeeRate) / 100.0,
			delegators:              make(map[codec.Address]*Delegator),
			epochRewards:            make(map[uint64]uint64),
			delegatedAmountPerEpoch: make(map[uint64]uint64),
			stakeStartBlock:         stakeStartBlock,
			stakeEndBlock:           stakeEndBlock,
		}
		e.validators[nodeID] = validator
	}

	validator.IsActive = true
	e.addActivationEvent(stakeStartBlock, validator)
	e.addDeactivationEvent(stakeEndBlock, validator)

	e.c.Logger().Info("validator registered",
		zap.String("nodeID", nodeID.String()),
		zap.Uint64("stakedAmount", stakedAmount),
		zap.Uint64("stakeStartBlock", stakeStartBlock),
		zap.Uint64("stakeEndBlock", stakeEndBlock),
	)

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
	rewardAmount := validator.AccumulatedStakedReward
	validator.AccumulatedStakedReward = 0

	// Mark the validator as inactive
	validator.IsActive = false
	validator.StakedAmount = 0

	e.c.Logger().Info("validator stake withdrawn",
		zap.String("nodeID", nodeID.String()),
		zap.Uint64("rewardAmount", rewardAmount),
	)

	return rewardAmount, nil
}

// DelegateUserStake increases the delegated stake for a validator and rebalances the heap.
func (e *Emission) DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeStartBlock, stakeEndBlock, stakedAmount uint64) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("delegating user stake")

	validator, exists := e.validators[nodeID]
	utils.Outf("=============DelegateUserStake - validator exists: ", exists)
	if !exists {
		return ErrValidatorNotFound
	}

	_, exists = validator.delegators[delegatorAddress]
	utils.Outf("=============DelegateUserStake - delegator exists: ", exists)
	if exists {
		return ErrDelegatorAlreadyStaked
	}

	delegator := &Delegator{
		IsActive:        false,
		StakedAmount:    stakedAmount,
		StakeStartBlock: stakeStartBlock,
		StakeEndBlock:   stakeEndBlock,
	}
	validator.delegators[delegatorAddress] = delegator

	e.addDelegatorActivationEvent(stakeStartBlock, nodeID, delegatorAddress)
	e.addDelegatorDeactivationEvent(stakeEndBlock, nodeID, delegatorAddress)

	e.c.Logger().Info("delegator registered",
		zap.String("nodeID", nodeID.String()),
		zap.Uint64("stakedAmount", stakedAmount),
		zap.Uint64("stakeStartBlock", stakeStartBlock),
		zap.Uint64("stakeEndBlock", stakeEndBlock),
	)

	return nil
}

func (e *Emission) addDelegatorActivationEvent(blockHeight uint64, nodeID ids.NodeID, delegatorAddress codec.Address) {
	event := &DelegatorEvent{
		ValidatorNodeID: nodeID,
		Delegator:       delegatorAddress,
	}
	if _, exists := e.delegatorActivationEvents[blockHeight]; !exists {
		e.delegatorActivationEvents[blockHeight] = []*DelegatorEvent{}
	}
	e.delegatorActivationEvents[blockHeight] = append(e.delegatorActivationEvents[blockHeight], event)
}

func (e *Emission) addDelegatorDeactivationEvent(blockHeight uint64, nodeID ids.NodeID, delegatorAddress codec.Address) {
	event := &DelegatorEvent{
		ValidatorNodeID: nodeID,
		Delegator:       delegatorAddress,
	}
	if _, exists := e.delegatorDeactivationEvents[blockHeight]; !exists {
		e.delegatorDeactivationEvents[blockHeight] = []*DelegatorEvent{}
	}
	e.delegatorDeactivationEvents[blockHeight] = append(e.delegatorDeactivationEvents[blockHeight], event)
}

// UndelegateUserStake decreases the delegated stake for a validator and rebalances the heap.
func (e *Emission) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("undelegating user stake",
		zap.String("nodeID", nodeID.String()))

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	// Check if the delegator exists
	_, exists = validator.delegators[actor]
	if !exists {
		e.c.Logger().Error("delegator not found")
		return 0, ErrDelegatorNotFound
	}

	// Calculate rewards while undelegating
	rewardAmount, err := e.CalculateUserDelegationRewards(nodeID, actor)
	if err != nil {
		e.c.Logger().Error("error calculating rewards", zap.Error(err))
		return 0, err
	}
	// Ensure AccumulatedDelegatedReward does not become negative
	if rewardAmount > validator.AccumulatedDelegatedReward {
		rewardAmount = validator.AccumulatedDelegatedReward
	}
	validator.AccumulatedDelegatedReward -= rewardAmount

	e.c.Logger().Info("removing delegator from validator",
		zap.String("nodeID", nodeID.String()))

	// Remove the delegator from the list
	delete(validator.delegators, actor)

	e.c.Logger().Info("checking validator status",
		zap.String("nodeID", nodeID.String()),
		zap.Int("remainingDelegators", len(validator.delegators)))

	// If the validator is inactive and has withdrawn and has no more delegators, remove the validator
	if !validator.IsActive && validator.StakedAmount == 0 && len(validator.delegators) == 0 {
		e.c.Logger().Info("removing validator",
			zap.String("nodeID", nodeID.String()))
		delete(e.validators, nodeID)
	}

	e.c.Logger().Info("undelegated user stake",
		zap.String("nodeID", nodeID.String()),
		zap.Uint64("rewardAmount", rewardAmount))

	return rewardAmount, nil
}

// ClaimStakingRewards lets validators and delegators claim their rewards
func (e *Emission) ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("claiming staking rewards",
		zap.String("nodeID", nodeID.String()),
	)

	// Find the validator
	validator, exists := e.validators[nodeID]
	if !exists {
		return 0, ErrValidatorNotFound
	}

	rewardAmount := uint64(0)
	if actor == codec.EmptyAddress {
		// Validator claiming their rewards
		rewardAmount = validator.AccumulatedStakedReward
		validator.AccumulatedStakedReward = 0 // Reset unclaimed rewards
	} else {
		reward, err := e.CalculateUserDelegationRewards(nodeID, actor)
		if err != nil {
			return 0, err
		}
		rewardAmount = reward
		validator.AccumulatedDelegatedReward -= rewardAmount

	}

	e.c.Logger().Info("staking rewards claimed", zap.Uint64("rewardAmount", rewardAmount))
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
				e.TotalStaked -= (validator.StakedAmount + validator.DelegatedAmount)
			}
		}
		delete(e.deactivationEvents, blockHeight)
	}

	if events, exists := e.delegatorActivationEvents[blockHeight]; exists {
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
		}
		delete(e.delegatorActivationEvents, blockHeight)
	}

	if events, exists := e.delegatorDeactivationEvents[blockHeight]; exists {
		for _, event := range events {
			validator, exists := e.validators[event.ValidatorNodeID]
			if !exists {
				continue
			}
			delegator, exists := validator.delegators[event.Delegator]
			if !exists {
				continue
			}
			if blockHeight >= delegator.StakeEndBlock && delegator.IsActive {
				delegator.IsActive = false
				validator.DelegatedAmount -= delegator.StakedAmount
				e.TotalStaked -= delegator.StakedAmount
			}
		}
		delete(e.delegatorDeactivationEvents, blockHeight)
	}
}

func (e *Emission) distributeValidatorRewardsOrFees(totalAmount uint64, isReward bool) uint64 {
	amountPerStakeUnit := float64(0)
	if e.TotalStaked > 0 {
		amountPerStakeUnit = float64(totalAmount) / float64(e.TotalStaked)
	}

	actualDistributedAmount := uint64(0)
	for _, validator := range e.validators {
		if !validator.IsActive {
			continue
		}

		validatorStake := validator.StakedAmount + validator.DelegatedAmount
		totalValidatorAmount := uint64(float64(validatorStake) * amountPerStakeUnit)

		validatorRewardAmount, delegationRewardAmount := distributeValidatorRewards(totalValidatorAmount, validator.DelegationFeeRate, validator.DelegatedAmount)

		actualDistributedAmount += validatorRewardAmount + delegationRewardAmount

		validator.AccumulatedStakedReward += validatorRewardAmount

		if isReward {
			epochNumber := e.GetLastAcceptedBlockHeight() / e.EpochTracker.EpochLength
			validator.epochRewards[epochNumber] = delegationRewardAmount
			validator.delegatedAmountPerEpoch[epochNumber] = validator.DelegatedAmount
			validator.AccumulatedDelegatedReward += delegationRewardAmount
		} else {
			validator.AccumulatedStakedReward += actualDistributedAmount
		}
	}

	return actualDistributedAmount
}

func (e *Emission) MintNewNAI() uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	currentBlockHeight := e.GetLastAcceptedBlockHeight()
	e.processEvents(currentBlockHeight)

	if currentBlockHeight%e.EpochTracker.EpochLength == 0 {
		e.c.Logger().Info("minting new NAI tokens at the end of the epoch")

		totalEpochRewards := e.GetRewardsPerEpoch()
		actualDistributedRewards := e.distributeValidatorRewardsOrFees(totalEpochRewards, true)

		e.TotalSupply += actualDistributedRewards
		return actualDistributedRewards
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
	e.EmissionAccount.AccumulatedReward += feesForEmission

	feesForValidators := fee - feesForEmission
	if e.TotalStaked == 0 || feesForValidators == 0 {
		return
	}

	e.distributeValidatorRewardsOrFees(feesForValidators, false)
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
			v.AccumulatedStakedReward = stakedValidator[0].AccumulatedStakedReward
			v.DelegationFeeRate = stakedValidator[0].DelegationFeeRate
			v.DelegatedAmount = stakedValidator[0].DelegatedAmount
			v.AccumulatedDelegatedReward = stakedValidator[0].AccumulatedDelegatedReward
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
