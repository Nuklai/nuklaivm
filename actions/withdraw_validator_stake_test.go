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

func TestWithdrawValidatorStakeAction(t *testing.T) {
	emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 200, // Mock block height
		StakeRewards:            100, // Mock reward amount
	})

	addr := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "ValidatorNotYetRegistered",
			Actor: addr,
			Action: &WithdrawValidatorStake{
				NodeID: nodeID, // Non-existent validator
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrNotValidator,
		},
		{
			Name:  "StakeNotStarted",
			Actor: addr,
			Action: &WithdrawValidatorStake{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the validator with stake end block greater than the current block height
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 150, 300, 10000, 10, addr, addr))
				return store
			}(),
			ExpectedErr: ErrStakeNotStarted,
		},
		{
			Name:     "ValidWithdrawal",
			ActionID: ids.GenerateTestID(),
			Actor:    addr,
			Action: &WithdrawValidatorStake{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set validator stake with end block less than the current block height
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, 10000, 10, addr, addr))
				// Set the balance for the validator
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, ids.Empty, 0))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly updated
				balance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, uint64(10100), balance) // Reward amount + staked amount

				// Check if the stake was successfully withdrawn
				exists, _, _, _, _, _, _, _ := storage.GetRegisterValidatorStake(ctx, store, nodeID)
				require.False(t, exists) // Stake should no longer exist
			},
			ExpectedOutputs: &WithdrawValidatorStakeResult{
				StakeStartBlock:      50,
				StakeEndBlock:        150,
				UnstakedAmount:       10000,
				DelegationFeeRate:    10,
				RewardAmount:         100,
				BalanceBeforeUnstake: 0,
				BalanceAfterUnstake:  10100, // Reward + Stake amount
				DistributedTo:        addr,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkWithdrawValidatorStake(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 200, // Mock block height
		StakeRewards:            100, // Mock reward amount
	})

	withdrawValidatorStakeBenchmark := &chaintest.ActionBenchmark{
		Name:  "WithdrawValidatorStakeBenchmark",
		Actor: actor,
		Action: &WithdrawValidatorStake{
			NodeID: nodeID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set validator stake with end block less than the current block height
			require.NoError(storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, 10000, 10, actor, actor))
			// Set the balance for the validator
			require.NoError(storage.SetBalance(context.Background(), store, actor, ids.Empty, 0))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly updated
			balance, err := storage.GetBalance(ctx, store, actor, ids.Empty)
			require.NoError(err)
			require.Equal(b, uint64(10100), balance) // Reward amount + staked amount
		},
	}

	ctx := context.Background()
	withdrawValidatorStakeBenchmark.Run(ctx, b)
}
