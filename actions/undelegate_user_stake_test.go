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

func TestUndelegateUserStakeActionFailure(t *testing.T) {
	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 25, StakeRewards: 20})

	addr := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "StakeMissing",
			Actor: addr,
			Action: &UndelegateUserStake{
				NodeID: nodeID, // Non-existent stake
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrStakeMissing,
		},
		{
			Name:  "StakeNotEnded",
			Actor: addr,
			Action: &UndelegateUserStake{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set stake with end block greater than the current block height
				require.NoError(t, storage.SetDelegateUserStake(context.Background(), store, addr, nodeID, 25, 50, 1000, addr))
				return store
			}(),
			ExpectedErr: ErrStakeNotEnded,
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func TestUndelegateUserStakeActionSuccess(t *testing.T) {
	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 51, StakeRewards: 20})

	addr := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "ValidUnstake",
			Actor: addr,
			Action: &UndelegateUserStake{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set stake with end block less than the current block height
				require.NoError(t, storage.SetDelegateUserStake(context.Background(), store, addr, nodeID, 25, 50, 1000, addr))
				// Set user balance before unstaking
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, ids.Empty, 0))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly updated after unstaking
				balance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, uint64(1020), balance)

				// Check if the stake was deleted
				exists, _, _, _, _, _, _ := storage.GetDelegateUserStake(ctx, store, addr, nodeID)
				require.False(t, exists)
			},
			ExpectedOutputs: &UndelegateUserStakeResult{
				StakeStartBlock:      25,
				StakeEndBlock:        50,
				UnstakedAmount:       1000,
				RewardAmount:         20,
				BalanceBeforeUnstake: 0,
				BalanceAfterUnstake:  1020,
				DistributedTo:        addr,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkUndelegateUserStake(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 51, StakeRewards: 20})

	undelegateUserStakeBenchmark := &chaintest.ActionBenchmark{
		Name:  "UndelegateUserStakeBenchmark",
		Actor: actor,
		Action: &UndelegateUserStake{
			NodeID: nodeID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set stake with end block less than the current block height
			require.NoError(storage.SetDelegateUserStake(context.Background(), store, actor, nodeID, 25, 50, 1000, actor))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly updated after unstaking
			balance, err := storage.GetBalance(ctx, store, actor, ids.Empty)
			require.NoError(err)
			require.Equal(b, uint64(1020), balance)

			// Check if the stake was deleted
			exists, _, _, _, _, _, _ := storage.GetDelegateUserStake(ctx, store, actor, nodeID)
			require.False(exists)
		},
	}

	ctx := context.Background()
	undelegateUserStakeBenchmark.Run(ctx, b)
}
