// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"

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

func (c *ClaimDelegationStakeRewards) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(c.NodeID)
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):        state.All,
		string(storage.DelegateUserStakeKey(actor, nodeID)): state.Read,
	}
}

func (*ClaimDelegationStakeRewards) StateKeysMaxChunks() []uint16 {
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
) ([][]byte, error) {
	nodeID, err := ids.ToNodeID(c.NodeID)
	if err != nil {
		return nil, ErrOutputInvalidNodeID
	}

	exists, stakeStartBlock, _, _, rewardAddress, _, _ := storage.GetDelegateUserStake(ctx, mu, c.UserStakeAddress, nodeID)
	if !exists {
		return nil, ErrOutputStakeMissing
	}
	if rewardAddress != actor {
		return nil, ErrOutputUnauthorized
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check that lastBlockHeight is after stakeStartBlock
	if emissionInstance.GetLastAcceptedBlockHeight() < stakeStartBlock {
		return nil, ErrOutputStakeNotEnded
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(nodeID, c.UserStakeAddress)
	if err != nil {
		return nil, err
	}

	if err := storage.AddBalance(ctx, mu, rewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return nil, err
	}

	sr := &ClaimRewardsResult{rewardAmount}
	output, err := sr.Marshal()
	if err != nil {
		return nil, err
	}
	return [][]byte{output}, nil
}

func (*ClaimDelegationStakeRewards) ComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (*ClaimDelegationStakeRewards) Size() int {
	return ids.NodeIDLen + codec.AddressLen
}

func (c *ClaimDelegationStakeRewards) Marshal(p *codec.Packer) {
	p.PackBytes(c.NodeID)
	p.PackAddress(c.UserStakeAddress)
}

func UnmarshalClaimDelegationStakeRewards(p *codec.Packer) (chain.Action, error) {
	var claimRewards ClaimDelegationStakeRewards
	p.UnpackBytes(ids.NodeIDLen, true, &claimRewards.NodeID)
	p.UnpackAddress(&claimRewards.UserStakeAddress)
	return &claimRewards, p.Err()
}

func (*ClaimDelegationStakeRewards) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
