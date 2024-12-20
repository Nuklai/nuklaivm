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
	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 10, StakeRewards: 20})

	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "ValidatorNotYetRegistered",
			Actor: actor,
			Action: &DelegateUserStake{
				NodeID:          nodeID, // Non-existent validator node ID
				StakeStartBlock: 25,
				StakeEndBlock:   50,
				StakedAmount:    1000,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrValidatorNotYetRegistered,
		},
		{
			Name:  "UserAlreadyStaked",
			Actor: actor,
			Action: &DelegateUserStake{
				NodeID:          nodeID,
				StakeStartBlock: 25,
				StakeEndBlock:   50,
				StakedAmount:    1000,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetValidatorStake(context.Background(), store, nodeID, 25, 50, 5000, 10, actor, actor))
				// Set the user stake
				require.NoError(t, storage.SetDelegatorStake(context.Background(), store, actor, nodeID, 25, 50, 1000, actor))
				return store
			}(),
			ExpectedErr: ErrUserAlreadyStaked,
		},
		{
			Name:  "InvalidStakedAmount",
			Actor: actor,
			Action: &DelegateUserStake{
				NodeID:          nodeID,
				StakeStartBlock: 25,
				StakeEndBlock:   50,
				StakedAmount:    100, // Invalid staked amount, less than min stake
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetValidatorStake(context.Background(), store, nodeID, 25, 50, 5000, 10, actor, actor))
				return store
			}(),
			ExpectedErr: ErrDelegateStakedAmountInvalid,
		},
		{
			Name:  "ValidStake",
			Actor: actor,
			Action: &DelegateUserStake{
				NodeID:          nodeID,
				StakeStartBlock: 25,
				StakeEndBlock:   50,
				StakedAmount:    emission.GetStakingConfig().MinDelegatorStake, // Valid staked amount
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetValidatorStake(context.Background(), store, nodeID, 25, 50, emission.GetStakingConfig().MinValidatorStake, 10, actor, actor))
				// Set the balance for the user
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, storage.NAIAddress, actor, emission.GetStakingConfig().MinDelegatorStake*2))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly deducted
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
				require.NoError(t, err)
				require.Equal(t, emission.GetStakingConfig().MinDelegatorStake, balance)

				// Check if the stake was created correctly
				exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, _, _ := storage.GetDelegatorStakeNoController(ctx, store, actor, nodeID)
				require.True(t, exists)
				require.Equal(t, uint64(25), stakeStartBlock)
				require.Equal(t, uint64(50), stakeEndBlock)
				require.Equal(t, emission.GetStakingConfig().MinDelegatorStake, stakedAmount)
				require.Equal(t, actor, rewardAddress)
			},
			ExpectedOutputs: &DelegateUserStakeResult{
				Actor:              actor.String(),
				Receiver:           nodeID.String(),
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

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 10, StakeRewards: 20})

	delegateUserStakeBenchmark := &chaintest.ActionBenchmark{
		Name:  "DelegateUserStakeBenchmark",
		Actor: actor,
		Action: &DelegateUserStake{
			NodeID:          nodeID,
			StakeStartBlock: 25,
			StakeEndBlock:   50,
			StakedAmount:    emission.GetStakingConfig().MinDelegatorStake,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Register the validator
			require.NoError(storage.SetValidatorStake(context.Background(), store, nodeID, 25, 50, emission.GetStakingConfig().MinValidatorStake, 10, actor, actor))
			// Set the balance for the user
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, storage.NAIAddress, actor, emission.GetStakingConfig().MinDelegatorStake*2))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly deducted
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
			require.NoError(err)
			require.Equal(b, emission.GetStakingConfig().MinDelegatorStake, balance)

			// Check if the stake was created correctly
			exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, _, _ := storage.GetDelegatorStakeNoController(ctx, store, actor, nodeID)
			require.True(exists)
			require.Equal(b, uint64(25), stakeStartBlock)
			require.Equal(b, uint64(50), stakeEndBlock)
			require.Equal(b, emission.GetStakingConfig().MinDelegatorStake, stakedAmount)
			require.Equal(b, actor, rewardAddress)
		},
	}

	ctx := context.Background()
	delegateUserStakeBenchmark.Run(ctx, b)
}
