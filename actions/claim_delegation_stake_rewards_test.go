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

func TestClaimDelegationStakeRewardsActionFailure(t *testing.T) {
	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 10, StakeRewards: 20})

	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:  "StakeMissing",
			Actor: actor,
			Action: &ClaimDelegationStakeRewards{
				NodeID: nodeID, // Non-existent stake
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrStakeMissing,
		},
		{
			Name:  "StakeNotStarted",
			Actor: actor,
			Action: &ClaimDelegationStakeRewards{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set stake with end block greater than the current block height
				require.NoError(t, storage.SetDelegatorStake(context.Background(), store, actor, nodeID, 25, 50, 1000, actor))
				return store
			}(),
			ExpectedErr: ErrStakeNotStarted,
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func TestClaimDelegationStakeRewardsActionSuccess(t *testing.T) {
	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 51, StakeRewards: 20})

	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	tests := []chaintest.ActionTest{
		{
			Name:     "ValidClaim",
			ActionID: ids.GenerateTestID(),
			Actor:    actor,
			Action: &ClaimDelegationStakeRewards{
				NodeID: nodeID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set stake with end block less than the current block height
				require.NoError(t, storage.SetDelegatorStake(context.Background(), store, actor, nodeID, 25, 50, 1000, actor))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly updated
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
				require.NoError(t, err)
				require.Equal(t, uint64(20), balance)

				// Check if the stake was claimed correctly
				exists, _, _, _, rewardAddress, _, _ := storage.GetDelegatorStakeNoController(ctx, store, actor, nodeID)
				require.True(t, exists)
				require.Equal(t, actor, rewardAddress)
			},
			ExpectedOutputs: &ClaimDelegationStakeRewardsResult{
				Actor:              actor.String(),
				Receiver:           actor.String(),
				StakeStartBlock:    25,
				StakeEndBlock:      50,
				StakedAmount:       1000,
				BalanceBeforeClaim: 0,
				BalanceAfterClaim:  20,
				DistributedTo:      actor.String(),
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

// BenchmarkClaimDelegationStakeRewards remains unchanged.
func BenchmarkClaimDelegationStakeRewards(b *testing.B) {
	require := require.New(b)
	actor := codectest.NewRandomAddress()
	nodeID := ids.GenerateTestNodeID()

	emission.MockNewEmission(&emission.MockEmission{LastAcceptedBlockHeight: 51, StakeRewards: 20})

	claimStakeRewardsBenchmark := &chaintest.ActionBenchmark{
		Name:  "ClaimStakeRewardsBenchmark",
		Actor: actor,
		Action: &ClaimDelegationStakeRewards{
			NodeID: nodeID,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set stake with end block less than the current block height
			require.NoError(storage.SetDelegatorStake(context.Background(), store, actor, nodeID, 25, 50, 1000, actor))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly updated
			balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
			require.NoError(err)
			require.Equal(b, uint64(20), balance)
		},
	}

	ctx := context.Background()
	claimStakeRewardsBenchmark.Run(ctx, b)
}
