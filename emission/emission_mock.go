// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
)

var _ Tracker = (*MockEmission)(nil)

type MockEmission struct {
	TotalSupplyVal          uint64
	RewardsPerEpoch         uint64
	APRForValidators        float64
	NumDelegators           int
	StakeRewards            uint64
	LastAcceptedBlockHeight uint64
}

func MockNewEmission(mockEmission *MockEmission) (*MockEmission, error) {
	emission = mockEmission
	return mockEmission, nil
}

func (m *MockEmission) AddToTotalSupply(amount uint64) uint64 {
	return m.TotalSupplyVal
}

func (m *MockEmission) GetRewardsPerEpoch() uint64 {
	return m.RewardsPerEpoch
}

func (m *MockEmission) GetAPRForValidators() float64 {
	return m.APRForValidators
}

func (m *MockEmission) GetNumDelegators(nodeID ids.NodeID) int {
	return m.NumDelegators
}

func (m *MockEmission) CalculateUserDelegationRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	return m.StakeRewards, nil
}

func (m *MockEmission) RegisterValidatorStake(nodeID ids.NodeID, nodePublicKey *bls.PublicKey, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate uint64) error {
	return nil
}

func (m *MockEmission) WithdrawValidatorStake(nodeID ids.NodeID) (uint64, error) {
	return m.StakeRewards, nil
}

func (m *MockEmission) DelegateUserStake(nodeID ids.NodeID, delegatorAddress codec.Address, stakeStartBlock, stakeEndBlock, stakedAmount uint64) error {
	return nil
}

func (m *MockEmission) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	return m.StakeRewards, nil
}

func (m *MockEmission) ClaimStakingRewards(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
	return m.StakeRewards, nil
}

func (m *MockEmission) MintNewNAI() uint64 {
	return m.StakeRewards
}

func (m *MockEmission) DistributeFees(fee uint64) {
}

func (m *MockEmission) GetStakedValidator(nodeID ids.NodeID) []*Validator {
	return nil
}

func (m *MockEmission) GetAllValidators(ctx context.Context) []*Validator {
	return nil
}

func (m *MockEmission) GetDelegatorsForValidator(nodeID ids.NodeID) ([]*Delegator, error) {
	return nil, nil
}

func (m *MockEmission) GetLastAcceptedBlockTimestamp() time.Time {
	return time.Now()
}

func (m *MockEmission) GetLastAcceptedBlockHeight() uint64 {
	return m.LastAcceptedBlockHeight
}

func (m *MockEmission) GetEmissionValidators() map[ids.NodeID]*Validator {
	return nil
}

func (m *MockEmission) GetInfo() (emissionAccount EmissionAccount, totalSupply uint64, maxSupply uint64, totalStaked uint64, epochTracker EpochTracker) {
	return EmissionAccount{}, 0, 0, 0, EpochTracker{}
}

// Helper functions for testing
func (m *MockEmission) SetStakeRewards(rewards uint64) {
	m.StakeRewards = rewards
}

func (m *MockEmission) SetLastAcceptedBlockHeight(height uint64) {
	m.LastAcceptedBlockHeight = height
}

func GetMockEmission() *MockEmission {
	return emission.(*MockEmission)
}
