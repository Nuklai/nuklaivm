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

func TestClaimValidatorStakeRewardsActionFailure(t *testing.T) {
	_, err := emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 10, StakeRewards: 20})
	require.NoError(t, err)

	addr := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "StakeMissing",
			Actor: addr,
			Action: &ClaimValidatorStakeRewards{
				NodeID: nodeID, // Non-existent stake
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrStakeMissing,
		},
		{
			Name:  "StakeNotStarted",
			Actor: addr,
			Action: &ClaimValidatorStakeRewards{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set validator stake with end block greater than the current block height
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 25, 50, 5000, 10, addr, addr))
				return store
			}(),
			ExpectedErr: ErrStakeNotStarted,
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func TestClaimValidatorStakeRewardsActionSuccess(t *testing.T) {
	_, err := emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 51, StakeRewards: 20})
	require.NoError(t, err)

	addr := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:     "ValidClaim",
			ActionID: ids.GenerateTestID(),
			Actor:    addr,
			Action: &ClaimValidatorStakeRewards{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set validator stake with end block less than the current block height
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 25, 50, emission.GetStakingConfig().MinValidatorStake, 10, addr, addr))
				// Set the balance for the validator
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, ids.Empty, 0))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly updated
				balance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, uint64(20), balance)

				// Check if the stake still exists after claiming rewards
				exists, _, _, _, _, rewardAddress, _, _ := storage.GetRegisterValidatorStake(ctx, store, nodeID)
				require.True(t, exists)
				require.Equal(t, addr, rewardAddress)
			},
			ExpectedOutputs: &ClaimValidatorStakeRewardsResult{
				StakeStartBlock:    25,
				StakeEndBlock:      50,
				StakedAmount:       emission.GetStakingConfig().MinValidatorStake,
				DelegationFeeRate:  10,
				BalanceBeforeClaim: 0,
				BalanceAfterClaim:  20,
				DistributedTo:      addr,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkClaimValidatorStakeRewards(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	_, err := emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 51, StakeRewards: 20})
	require.NoError(err)

	claimValidatorStakeRewardsBenchmark := &chaintest.ActionBenchmark{
		Name:  "ClaimValidatorStakeRewardsBenchmark",
		Actor: actor,
		Action: &ClaimValidatorStakeRewards{
			NodeID: nodeID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set validator stake with end block less than the current block height
			require.NoError(storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 25, 50, emission.GetStakingConfig().MinValidatorStake, 10, actor, actor))
			// Set the balance for the validator
			require.NoError(storage.SetBalance(context.Background(), store, actor, ids.Empty, 0))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly updated after claiming rewards
			balance, err := storage.GetBalance(ctx, store, actor, ids.Empty)
			require.NoError(err)
			require.Equal(b, uint64(20), balance) // Reward amount set by emission instance
		},
	}

	ctx := context.Background()
	claimValidatorStakeRewardsBenchmark.Run(ctx, b)
}
