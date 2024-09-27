// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

/* import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/stretchr/testify/assert"
)

// Assume that MockController and MockNuklaiVM are already implemented
// with the necessary mocked methods.

func TestRegisterAndUnregisterValidator(t *testing.T) {
	c := &MockController{}
	vm := &MockNuklaiVM{}
	emission := New(c, vm, 10000, 20000, codec.Address{})

	nodeID := ids.GenerateTestNodeID()
	publicKey := []byte("publicKey")
	stakedAmount := uint64(100)
	delegationFeeRate := uint64(10)

	// Test new registration
	err := emission.RegisterValidatorStake(nodeID, publicKey, stakedAmount, delegationFeeRate)
	assert.NoError(t, err, "Registering new validator should not produce an error")

	// Ensure validator is marked as active
	validator, exists := emission.validators[nodeID]
	assert.True(t, exists, "Validator should exist after registration")
	assert.True(t, validator.IsActive, "Validator should be marked as active after registration")

	// Test unregistration
	err = emission.UnregisterValidatorStake(nodeID)
	assert.NoError(t, err, "Unregistering validator should not produce an error")

	// Ensure validator is marked as inactive
	validator, exists = emission.validators[nodeID]
	assert.True(t, exists, "Validator should still exist after unregistration")
	assert.False(t, validator.IsActive, "Validator should be marked as inactive after unregistration")

	// Test re-registration
	err = emission.RegisterValidatorStake(nodeID, publicKey, stakedAmount, delegationFeeRate)
	assert.NoError(t, err, "Re-registering validator should not produce an error")

	// Ensure validator is marked as active again
	validator, exists = emission.validators[nodeID]
	assert.True(t, exists, "Validator should exist after re-registration")
	assert.True(t, validator.IsActive, "Validator should be marked as active after re-registration")
}

func TestDelegateAndUndelegateStake(t *testing.T) {
	c := &MockController{}
	vm := &MockNuklaiVM{}
	emission := New(c, vm, 10000, 20000, codec.Address{})

	nodeID := ids.GenerateTestNodeID()
	publicKey := []byte("publicKey")
	stakedAmount := uint64(100)
	delegationFeeRate := uint64(10)

	// Register a validator
	err := emission.RegisterValidatorStake(nodeID, publicKey, stakedAmount, delegationFeeRate)
	assert.NoError(t, err)

	delegatorAddr := codec.Address{1} // Example delegator address
	delegationAmount := uint64(50)

	// Delegate stake
	err = emission.DelegateUserStake(nodeID, delegatorAddr, delegationAmount)
	assert.NoError(t, err, "Delegating stake should not produce an error")

	// Check updated delegated amount
	validator, _ := emission.validators[nodeID]
	assert.Equal(t, delegationAmount, validator.DelegatedAmount, "Delegated amount should be updated after delegation")

	// Undelegate stake
	_, err = emission.UndelegateUserStake(nodeID, delegatorAddr, delegationAmount)
	assert.NoError(t, err, "Undelegating stake should not produce an error")

	// Check updated delegated amount
	validator, _ = emission.validators[nodeID]
	assert.Equal(t, uint64(0), validator.DelegatedAmount, "Delegated amount should be zero after undelegation")
}

func TestClaimRewards(t *testing.T) {
	// This test would depend heavily on the implementation details of your
	// reward calculation and distribution logic. You'll need to simulate an
	// environment where rewards are generated (perhaps by calling MintNewNAI
	// multiple times) and then ensure that validators and delegators can
	// claim their expected rewards.
}

// Additional tests should be written to cover more scenarios such as:
// - Validators re-registering after all delegators have left
// - Delegators staking and unstaking across multiple epochs
// - Edge cases where actions are taken at the boundary of an epoch

// Remember to implement cleanup logic in your tests if necessary, especially
// for global or persistent state that might affect subsequent tests.
*/
