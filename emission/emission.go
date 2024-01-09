// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
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
}

type UserStake struct {
	Owner        string                `json:"owner"`     // we always send address over RPC
	StakeInfo    map[string]*StakeInfo `json:"stakeInfo"` // the key is txID
	StakedAmount uint64                `json:"amount"`

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
		for nodeID, validator := range currentValidators {
			nodeIDString := nodeID.String()
			newValidator := &Validator{
				NodeID:        nodeIDString,
				NodePublicKey: base64.StdEncoding.EncodeToString(validator.PublicKey.Compress()),
				UserStake:     make(map[string]*UserStake),
				StakedAmount:  0,
				StakedReward:  0,
			}
			validators[nodeIDString] = newValidator
		}

		emission = &Emission{
			c:               c,
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
//
//nolint:interfacer
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

	validators := make([]*Validator, 0, len(e.validators))
	for _, value := range e.validators {
		validators = append(validators, value)
	}
	return validators
}

// MintNewNAI mints new tokens and distributes them to validators
func (e *Emission) MintNewNAI(ctx context.Context, mu *state.SimpleMutable, emissionAddr codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	mintNewNAI := e.rewardsPerBlock
	if e.totalSupply+mintNewNAI > e.maxSupply {
		mintNewNAI = e.maxSupply - e.totalSupply // Adjust to not exceed max supply
	}
	if mintNewNAI == 0 {
		return 0, nil // Nothing to mint
	}

	totalStaked := e.totalStaked()
	// No validators to distribute rewards to if totalStaked is 0
	if totalStaked == 0 {
		if err := storage.AddBalance(ctx, mu, emissionAddr, mintNewNAI, true); err != nil {
			return 0, err
		}
		if err := mu.Commit(ctx); err != nil {
			return 0, err
		}
		return mintNewNAI, nil
	}

	// Distribute rewards based on stake proportion
	for _, v := range e.validators {
		if v.StakedAmount >= e.minStakeAmount {
			v.StakedReward += mintNewNAI * v.StakedAmount / totalStaked
		}
	}
	return mintNewNAI, nil
}

func (e *Emission) DistributeFees(ctx context.Context, mu *state.SimpleMutable, fee uint64, emissionAddr codec.Address) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if fee == 0 {
		return nil // Nothing to distribute
	}

	feesForEmission := fee / 2
	feesForValidators := fee - feesForEmission

	// Give 50% fees to Emission
	if err := storage.AddBalance(ctx, mu, emissionAddr, feesForEmission, true); err != nil {
		return err
	}
	if err := mu.Commit(ctx); err != nil {
		return err
	}

	// Give remaining to Validators
	totalStaked := e.totalStaked()
	if totalStaked > 0 {
		// Distribute rewards based on stake proportion
		for _, v := range e.validators {
			if v.StakedAmount >= e.minStakeAmount {
				v.StakedReward += feesForValidators * v.StakedAmount / totalStaked
			}
		}
	}

	return nil
}

// totalStaked calculates the total amount staked by all validators
func (e *Emission) totalStaked() uint64 {
	var total uint64
	for _, v := range e.validators {
		if v.StakedAmount >= e.minStakeAmount {
			total += v.StakedAmount
		}
	}
	return total
}

// ClaimRewards allows validators to claim their accumulated rewards
// TODO: Make it so that we track staking rewards automatically rather than validators having to claim them and distributing it to their stakers
func (e *Emission) ClaimRewards(ctx context.Context, mu *state.SimpleMutable, emissionAddr codec.Address, validatorNodeID string, sig *bls.Signature, msg []byte, toAddress codec.Address) (uint64, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	validator, ok := e.validators[validatorNodeID]
	if !ok {
		return 0, nil // Validator not found
	}

	if validator.StakedReward == 0 {
		return 0, nil // Nothing to claim
	}

	pubKey, err := bls.PublicKeyFromBytes([]byte(validator.NodePublicKey))
	if err != nil {
		return 0, fmt.Errorf("invalid public key") // Invalid public key
	}

	if !bls.Verify(pubKey, sig, msg) {
		return 0, fmt.Errorf("invalid signature") // Invalid signature
	}

	claimedRewards := validator.StakedReward
	validator.StakedReward = 0

	if err := storage.SubBalance(ctx, mu, emissionAddr, claimedRewards); err != nil {
		return 0, err
	}
	if err := storage.AddBalance(ctx, mu, toAddress, claimedRewards, true); err != nil {
		return 0, err
	}
	if err := mu.Commit(ctx); err != nil {
		return 0, err
	}

	return claimedRewards, nil
}
