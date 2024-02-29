// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*ClaimStakingRewards)(nil)

type ClaimStakingRewards struct {
	NodeID           []byte        `json:"nodeID"`           // Node ID of the validator where NAI is staked
	UserStakeAddress codec.Address `json:"userStakeAddress"` // Pass this if you're claiming rewards for delegated user stake
}

func (*ClaimStakingRewards) GetTypeID() uint8 {
	return nconsts.ClaimStakingRewardsID
}

func (c *ClaimStakingRewards) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(c.NodeID)
	if c.UserStakeAddress == codec.EmptyAddress {
		return []string{
			string(storage.BalanceKey(actor, ids.Empty)),
			string(storage.RegisterValidatorStakeKey(nodeID)),
			string(storage.DelegateUserStakeKey(actor, nodeID)),
		}
	}
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
		string(storage.DelegateUserStakeKey(actor, nodeID)),
	}
}

func (c *ClaimStakingRewards) StateKeysMaxChunks() []uint16 {
	if c.UserStakeAddress == codec.EmptyAddress {
		return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks, storage.DelegateUserStakeChunks}
	}
	return []uint16{storage.BalanceChunks, storage.DelegateUserStakeChunks}
}

func (*ClaimStakingRewards) OutputsWarpMessage() bool {
	return false
}

func (c *ClaimStakingRewards) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	timestamp int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	nodeID, err := ids.ToNodeID(c.NodeID)
	if err != nil {
		return false, ClaimStakingRewardComputeUnits, OutputInvalidNodeID, nil, nil
	}

	// Check whether a validator is trying to claim its reward
	exists, _, stakeEndTime, _, _, rewardAddress, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if c.UserStakeAddress != codec.EmptyAddress {
		exists, _, _, rewardAddress, _, _ = storage.GetDelegateUserStake(ctx, mu, c.UserStakeAddress, nodeID)
	}
	if !exists {
		return false, ClaimStakingRewardComputeUnits, OutputStakeMissing, nil, nil
	}
	if rewardAddress != actor {
		return false, ClaimStakingRewardComputeUnits, OutputUnauthorized, nil, nil
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check if the stake has ended for the validator
	if c.UserStakeAddress == codec.EmptyAddress {
		// Get current time
		currentTime := emissionInstance.GetLastAcceptedBlockTimestamp()
		// Convert Unix timestamps to Go's time.Time for easier manipulation
		endTime := time.Unix(int64(stakeEndTime), 0).UTC()
		// Check that currentTime is after stakeEndTime
		if currentTime.Before(endTime) {
			return false, ClaimStakingRewardComputeUnits, OutputStakeNotEnded, nil, nil
		}
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(nodeID, c.UserStakeAddress)
	if err != nil {
		return false, ClaimStakingRewardComputeUnits, utils.ErrBytes(err), nil, nil
	}

	if err := storage.AddBalance(ctx, mu, rewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return false, ClaimStakingRewardComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, ClaimStakingRewardComputeUnits, nil, nil, nil
}

func (*ClaimStakingRewards) MaxComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (c *ClaimStakingRewards) Size() int {
	return hconsts.NodeIDLen + codec.AddressLen
}

func (c *ClaimStakingRewards) Marshal(p *codec.Packer) {
	p.PackBytes(c.NodeID)
	p.PackAddress(c.UserStakeAddress)
}

func UnmarshalClaimStakingRewards(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var claimRewards ClaimStakingRewards
	p.UnpackBytes(hconsts.NodeIDLen, true, &claimRewards.NodeID)
	p.UnpackAddress(&claimRewards.UserStakeAddress)
	return &claimRewards, p.Err()
}

func (*ClaimStakingRewards) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
