// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"
)

func TestDelegateUserStakeAction(t *testing.T) {
	// Mock emission instance with a starting block height
	_, err := emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 50, // Mock block height
	})
	require.NoError(t, err)

	addr := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "ValidatorNotYetRegistered",
			Actor: addr,
			Action: &DelegateUserStake{
				NodeID:          nodeID, // Non-existent validator node ID
				StakeStartBlock: 55,
				StakeEndBlock:   100,
				StakedAmount:    1000,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrValidatorNotYetRegistered,
		},
		{
			Name:  "UserAlreadyStaked",
			Actor: addr,
			Action: &DelegateUserStake{
				NodeID:          nodeID,
				StakeStartBlock: 55,
				StakeEndBlock:   100,
				StakedAmount:    1000,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, 5000, 100, addr, addr))
				// Set the user stake
				require.NoError(t, storage.SetDelegateUserStake(context.Background(), store, addr, nodeID, 55, 100, 1000, addr))
				return store
			}(),
			ExpectedErr: ErrUserAlreadyStaked,
		},
		{
			Name:  "InvalidStakedAmount",
			Actor: addr,
			Action: &DelegateUserStake{
				NodeID:          nodeID,
				StakeStartBlock: 55,
				StakeEndBlock:   100,
				StakedAmount:    50, // Invalid staked amount, less than min stake
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, 5000, 100, addr, addr))
				return store
			}(),
			ExpectedErr: ErrDelegateStakedAmountInvalid,
		},
		{
			Name:     "ValidStake",
			ActionID: ids.GenerateTestID(),
			Actor:    addr,
			Action: &DelegateUserStake{
				NodeID:          nodeID,
				StakeStartBlock: 55,
				StakeEndBlock:   100,
				StakedAmount:    emission.GetStakingConfig().MinDelegatorStake, // Valid staked amount
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, emission.GetStakingConfig().MinValidatorStake, 100, addr, addr))
				// Set the balance for the user
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, ids.Empty, emission.GetStakingConfig().MinDelegatorStake*2))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly deducted
				balance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, emission.GetStakingConfig().MinDelegatorStake, balance)

				// Check if the stake was created correctly
				exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, _, _ := storage.GetDelegateUserStake(ctx, store, addr, nodeID)
				require.True(t, exists)
				require.Equal(t, uint64(55), stakeStartBlock)
				require.Equal(t, uint64(100), stakeEndBlock)
				require.Equal(t, emission.GetStakingConfig().MinDelegatorStake, stakedAmount)
				require.Equal(t, addr, rewardAddress)
			},
			ExpectedOutputs: &DelegateUserStakeResult{
				StakedAmount:       emission.GetStakingConfig().MinDelegatorStake,
				BalanceBeforeStake: emission.GetStakingConfig().MinDelegatorStake * 2,
				BalanceAfterStake:  emission.GetStakingConfig().MinDelegatorStake,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkDelegateUserStake(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	// Mock emission instance with a starting block height
	_, err := emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 50,
	})
	require.NoError(err)

	delegateUserStakeBenchmark := &chaintest.ActionBenchmark{
		Name:  "DelegateUserStakeBenchmark",
		Actor: actor,
		Action: &DelegateUserStake{
			NodeID:          nodeID,
			StakeStartBlock: 55,
			StakeEndBlock:   100,
			StakedAmount:    emission.GetStakingConfig().MinDelegatorStake,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Register the validator
			require.NoError(storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, emission.GetStakingConfig().MinValidatorStake, 100, actor, actor))
			// Set the balance for the user
			require.NoError(storage.SetBalance(context.Background(), store, actor, ids.Empty, emission.GetStakingConfig().MinDelegatorStake*2))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly deducted
			balance, err := storage.GetBalance(ctx, store, actor, ids.Empty)
			require.NoError(err)
			require.Equal(b, emission.GetStakingConfig().MinDelegatorStake, balance)

			// Check if the stake was created correctly
			exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, _, _ := storage.GetDelegateUserStake(ctx, store, actor, nodeID)
			require.True(exists)
			require.Equal(b, uint64(55), stakeStartBlock)
			require.Equal(b, uint64(100), stakeEndBlock)
			require.Equal(b, emission.GetStakingConfig().MinDelegatorStake, stakedAmount)
			require.Equal(b, actor, rewardAddress)
		},
	}

	ctx := context.Background()
	delegateUserStakeBenchmark.Run(ctx, b)
}
