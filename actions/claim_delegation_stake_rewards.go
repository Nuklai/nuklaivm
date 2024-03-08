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

var _ chain.Action = (*ClaimDelegationStakeRewards)(nil)

type ClaimDelegationStakeRewards struct {
	NodeID           []byte        `json:"nodeID"`           // Node ID of the validator where NAI is staked
	UserStakeAddress codec.Address `json:"userStakeAddress"` // The address of the user who delegated the stake
}

func (*ClaimDelegationStakeRewards) GetTypeID() uint8 {
	return nconsts.ClaimDelegationStakeRewards
}

func (c *ClaimDelegationStakeRewards) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(c.NodeID)
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
		string(storage.DelegateUserStakeKey(actor, nodeID)),
	}
}

func (c *ClaimDelegationStakeRewards) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.DelegateUserStakeChunks}
}

func (*ClaimDelegationStakeRewards) OutputsWarpMessage() bool {
	return false
}

func (c *ClaimDelegationStakeRewards) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	nodeID, err := ids.ToNodeID(c.NodeID)
	if err != nil {
		return false, ClaimStakingRewardComputeUnits, OutputInvalidNodeID, nil, nil
	}

	exists, stakeStartTime, _, rewardAddress, _, _ := storage.GetDelegateUserStake(ctx, mu, c.UserStakeAddress, nodeID)
	if !exists {
		return false, ClaimStakingRewardComputeUnits, OutputStakeMissing, nil, nil
	}
	if rewardAddress != actor {
		return false, ClaimStakingRewardComputeUnits, OutputUnauthorized, nil, nil
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Get current time
	currentTime := time.Now().UTC()
	// Get last accepted block time
	lastBlockTime := emissionInstance.GetLastAcceptedBlockTimestamp()
	// Convert Unix timestamps to Go's time.Time for easier manipulation
	startTime := time.Unix(int64(stakeStartTime), 0).UTC()
	// Check that currentTime and lastBlockTime are after stakeStartTime
	if currentTime.Before(startTime) || lastBlockTime.Before(startTime) {
		return false, ClaimStakingRewardComputeUnits, OutputStakeNotStarted, nil, nil
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(nodeID, c.UserStakeAddress)
	if err != nil {
		return false, ClaimStakingRewardComputeUnits, utils.ErrBytes(err), nil, nil
	}

	if err := storage.AddBalance(ctx, mu, rewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return false, ClaimStakingRewardComputeUnits, utils.ErrBytes(err), nil, nil
	}

	sr := &ClaimRewardsResult{rewardAmount}
	output, err := sr.Marshal()
	if err != nil {
		return false, ClaimStakingRewardComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, ClaimStakingRewardComputeUnits, output, nil, nil
}

func (*ClaimDelegationStakeRewards) MaxComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (*ClaimDelegationStakeRewards) Size() int {
	return hconsts.NodeIDLen + codec.AddressLen
}

func (c *ClaimDelegationStakeRewards) Marshal(p *codec.Packer) {
	p.PackBytes(c.NodeID)
	p.PackAddress(c.UserStakeAddress)
}

func UnmarshalClaimDelegationStakeRewards(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var claimRewards ClaimDelegationStakeRewards
	p.UnpackBytes(hconsts.NodeIDLen, true, &claimRewards.NodeID)
	p.UnpackAddress(&claimRewards.UserStakeAddress)
	return &claimRewards, p.Err()
}

func (*ClaimDelegationStakeRewards) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
