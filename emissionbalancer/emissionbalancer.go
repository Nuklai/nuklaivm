package emissionbalancer

import (
	"sync"
)

type Validator struct {
	NodeID        string
	NodePublicKey string
	NodeWeight    uint64
	WalletAddress string
	StakedAmount  uint64
	StakedReward  uint64
}

type EmissionBalancer struct {
	c Controller

	TotalSupply     uint64
	MaxSupply       uint64
	RewardsPerBlock uint64
	Validators      map[string]*Validator

	lock sync.RWMutex
}

var (
	balancer *EmissionBalancer
	once     sync.Once
)

// New initializes the Balancer with the maximum supply
func New(c Controller, maxSupply, rewardsPerBlock uint64) *EmissionBalancer {
	once.Do(func() {
		c.Logger().Info("setting maxSupply and rewardsPerBlock for emission balancer")
		balancer = &EmissionBalancer{
			c:               c,
			MaxSupply:       maxSupply,
			RewardsPerBlock: rewardsPerBlock,
			Validators:      make(map[string]*Validator),
		}
	})
	return balancer
}

// GetEmissionBalancer returns the singleton instance of EmissionBalancer
func GetEmissionBalancer() *EmissionBalancer {
	return balancer
}

// TODO: Make it so that the total supply is stored on the database
func (e *EmissionBalancer) AddToTotalSupply(amount uint64) uint64 {
	e.c.Logger().Info("adding to total supply of NAI")
	e.TotalSupply += amount
	return e.TotalSupply
}

// StakeValidator adds a new validator to the EmissionBalancer
// TODO: Make it so that the staked validators are stored on the database
// TODO: Make it so that we check whether the validator is part of the current validator set
func (e *EmissionBalancer) StakeValidator(signature, nodeID, nodePublicKey, walletAddress string, amountToStake uint64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Validators[nodeID] = &Validator{
		NodeID:        nodeID,
		NodePublicKey: nodePublicKey,
		NodeWeight:    0,
		WalletAddress: walletAddress,
		StakedAmount:  amountToStake,
	}
}

// MintAndDistribute mints new tokens and distributes them to validators
// TODO: Make it so that we check whether the validator is part of the current validator set
func (e *EmissionBalancer) MintAndDistribute() {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.TotalSupply >= e.MaxSupply {
		return // Cap reached, no more minting
	}

	totalStaked := e.totalStaked()
	if totalStaked == 0 {
		return // No validators to distribute rewards to
	}

	mintAmount := e.RewardsPerBlock
	if e.TotalSupply+mintAmount > e.MaxSupply {
		mintAmount = e.MaxSupply - e.TotalSupply // Adjust to not exceed max supply
	}
	e.TotalSupply += mintAmount

	// Distribute rewards based on stake proportion
	for _, v := range e.Validators {
		v.StakedReward += mintAmount * v.StakedAmount / totalStaked
	}
}

// totalStaked calculates the total amount staked by all validators
func (e *EmissionBalancer) totalStaked() uint64 {
	var total uint64
	for _, v := range e.Validators {
		total += v.StakedAmount
	}
	return total
}

// ClaimRewards allows validators to claim their accumulated rewards
func (e *EmissionBalancer) ClaimRewards(validatorNodeID string) uint64 {
	e.lock.Lock()
	defer e.lock.Unlock()

	v, ok := e.Validators[validatorNodeID]
	if !ok {
		return 0 // Validator not found
	}

	claimedRewards := v.StakedReward
	v.StakedReward = 0
	return claimedRewards
}
