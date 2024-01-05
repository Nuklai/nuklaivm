package emission

import (
	"encoding/base64"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
)

var (
	emission *Emission
	once     sync.Once
)

type StakeInfo struct {
	TxID        string `json:"txID"`
	Amount      uint64 `json:"amount"`
	StartLockUp uint64 `json:"startLockUp"`
	EndLockUp   uint64 `json:"endLockUp"`
	Reward      uint64 `json:"reward"`
}

type UserStake struct {
	Owner        string                `json:"owner"`     // we always send address over RPC
	StakeInfo    map[string]*StakeInfo `json:"stakeInfo"` // the key is txID
	StakedAmount uint64                `json:"amount"`
	StakedReward uint64                `json:"reward"`

	owner codec.Address
}

type Validator struct {
	NodeID        string                `json:"nodeID"`
	NodePublicKey string                `json:"nodePublicKey"`
	UserStake     map[string]*UserStake `json:"userStake"` // the key is Owner
	StakedAmount  uint64                `json:"stakedAmount"`
	StakedReward  uint64                `json:"stakedReward"`
}

type Emission struct {
	c Controller

	totalSupply     uint64
	maxSupply       uint64
	rewardsPerBlock uint64

	validators     map[string]*Validator // the key is NodeID
	maxValidators  int
	minStakeAmount uint64

	lock sync.RWMutex
}

// New initializes the Emission with the maximum supply
func New(c Controller, maxSupply, rewardsPerBlock uint64, currentValidators map[ids.NodeID]*validators.GetValidatorOutput) *Emission {
	once.Do(func() {
		c.Logger().Info("setting maxSupply and rewardsPerBlock for emission")

		validators := make(map[string]*Validator)
		for nodeId, validator := range currentValidators {
			nodeIdString := nodeId.String()
			newValidator := &Validator{
				NodeID:        nodeIdString,
				NodePublicKey: base64.StdEncoding.EncodeToString(validator.PublicKey.Compress()),
				UserStake:     make(map[string]*UserStake),
				StakedAmount:  0,
				StakedReward:  0,
			}
			validators[nodeIdString] = newValidator
		}

		emission = &Emission{
			c:               c,
			totalSupply:     0,
			maxSupply:       maxSupply,
			rewardsPerBlock: rewardsPerBlock,
			validators:      validators,
			maxValidators:   7,
			minStakeAmount:  100,
		}
	})
	return emission
}

// GetEmission returns the singleton instance of Emission
func GetEmission() *Emission {
	return emission
}

func (e *Emission) AddToTotalSupply(amount uint64) uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.c.Logger().Info("adding to total supply of NAI")
	if e.totalSupply+amount > e.maxSupply {
		amount = e.maxSupply - e.totalSupply // Adjust to not exceed max supply
	}
	e.totalSupply += amount
	return e.totalSupply
}

func (e *Emission) GetTotalSupply() uint64 {
	e.c.Logger().Info("fetching total supply of NAI")
	return e.totalSupply
}

func (e *Emission) GetMaxSupply() uint64 {
	e.c.Logger().Info("fetching max supply of NAI")
	return e.maxSupply
}

func (e *Emission) GetRewardsPerBlock() uint64 {
	e.c.Logger().Info("fetching amount of NAI rewards per block")
	return e.rewardsPerBlock
}

// StakeValidator stakes the validator
func (e *Emission) StakeToValidator(txID ids.ID, actor codec.Address, currentValidators map[ids.NodeID]*validators.GetValidatorOutput, startLockUp uint64, action *actions.StakeValidator) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	nodeID, err := ids.ToNodeID(action.NodeID)
	if err != nil {
		return ErrInvalidNodeID // Invalid NodeID
	}

	currentValidator, ok := currentValidators[nodeID]
	if !ok {
		return ErrNotAValidator // Not a validator
	}

	stakeOwner := codec.MustAddressBech32(consts.HRP, actor)
	validator, ok := e.validators[nodeID.String()]
	if !ok {
		if len(e.validators) >= e.maxValidators {
			return ErrMaxValidatorsReached // Cap reached, no more validators
		}
		validator = &Validator{
			NodeID:        currentValidator.NodeID.String(),
			NodePublicKey: base64.StdEncoding.EncodeToString(currentValidator.PublicKey.Compress()),
			UserStake:     map[string]*UserStake{},
			StakedReward:  0,
		}
	}
	userStake, ok := validator.UserStake[stakeOwner]
	if !ok {
		userStake = &UserStake{
			Owner:        stakeOwner,
			StakeInfo:    map[string]*StakeInfo{},
			StakedAmount: action.StakedAmount,
			StakedReward: 0,
			owner:        actor,
		}
	}
	stakeInfo, ok := userStake.StakeInfo[txID.String()]
	if !ok {
		stakeInfo = &StakeInfo{
			TxID:        txID.String(),
			Amount:      action.StakedAmount,
			StartLockUp: startLockUp,
			EndLockUp:   action.EndLockUp,
			Reward:      0,
		}
	}
	userStake.StakeInfo[txID.String()] = stakeInfo
	validator.UserStake[stakeOwner] = userStake
	validator.StakedAmount += action.StakedAmount

	e.validators[nodeID.String()] = validator

	return nil
}

func (e *Emission) GetValidator(nodeID string) []*Validator {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if nodeID == "" {
		return e.getAllValidators()
	}

	validator, ok := e.validators[nodeID]
	if !ok {
		return []*Validator{}
	}
	return []*Validator{validator}
}

func (e *Emission) GetUserStake(nodeID, owner string) *UserStake {
	e.lock.RLock()
	defer e.lock.RUnlock()

	validator, ok := e.validators[nodeID]
	if !ok {
		return &UserStake{}
	}

	userStake, ok := validator.UserStake[owner]
	if !ok {
		return &UserStake{}
	}
	return userStake
}

func (e *Emission) getAllValidators() []*Validator {
	e.lock.RLock()
	defer e.lock.RUnlock()

	var validators []*Validator
	for _, value := range e.validators {
		validators = append(validators, value)
	}
	return validators
}

/* // MintAndDistribute mints new tokens and distributes them to validators
// TODO: Make it so that we check whether the validator is part of the current validator set
func (s *Stake) MintAndDistribute() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.TotalSupply >= s.MaxSupply {
		return // Cap reached, no more minting
	}

	totalStaked := s.totalStaked()
	if totalStaked == 0 {
		return // No validators to distribute rewards to
	}

	mintAmount := s.RewardsPerBlock
	if s.TotalSupply+mintAmount > s.MaxSupply {
		mintAmount = s.MaxSupply - s.TotalSupply // Adjust to not exceed max supply
	}
	s.TotalSupply += mintAmount

	// Distribute rewards based on stake proportion
	for _, v := range s.Validators {
		v.StakedReward += mintAmount * v.StakedAmount / totalStaked
	}
}

// totalStaked calculates the total amount staked by all validators
func (s *Stake) totalStaked() uint64 {
	var total uint64
	for _, v := range s.Validators {
		total += v.StakedAmount
	}
	return total
}

// ClaimRewards allows validators to claim their accumulated rewards
func (s *Stake) ClaimRewards(validatorNodeID string) uint64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, ok := s.Validators[validatorNodeID]
	if !ok {
		return 0 // Validator not found
	}

	claimedRewards := v.StakedReward
	v.StakedReward = 0
	return claimedRewards
} */