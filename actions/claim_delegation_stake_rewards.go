// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var (
	ErrStakeMissing                  = errors.New("stake is missing")
	ErrUnauthorizedUser              = errors.New("user is not authorized")
	ErrStakeNotEnded                 = errors.New("stake has not ended")
	_                   chain.Action = (*ClaimDelegationStakeRewards)(nil)
)

type ClaimDelegationStakeRewards struct {
	NodeID ids.NodeID `serialize:"true" json:"node_id"` // Node ID of the validator where NAI is staked
}

func (*ClaimDelegationStakeRewards) GetTypeID() uint8 {
	return nconsts.ClaimDelegationStakeRewards
}

func (c *ClaimDelegationStakeRewards) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):          state.All,
		string(storage.DelegateUserStakeKey(actor, c.NodeID)): state.Read,
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
) (codec.Typed, error) {
	exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, _, _ := storage.GetDelegateUserStake(ctx, mu, actor, c.NodeID)
	if !exists {
		return nil, ErrStakeMissing
	}
	if rewardAddress != actor {
		return nil, ErrUnauthorizedUser
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check that lastBlockHeight is after stakeStartBlock
	if emissionInstance.GetLastAcceptedBlockHeight() < stakeStartBlock {
		return nil, ErrStakeNotEnded
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(c.NodeID, actor)
	if err != nil {
		return nil, err
	}

	balance, err := storage.AddBalance(ctx, mu, rewardAddress, ids.Empty, rewardAmount, true)
	if err != nil {
		return nil, err
	}

	return &ClaimDelegationStakeRewardsResult{
		StakeStartBlock:    stakeStartBlock,
		StakeEndBlock:      stakeEndBlock,
		StakedAmount:       stakedAmount,
		BalanceBeforeClaim: balance - rewardAmount,
		BalanceAfterClaim:  balance,
		DistributedTo:      rewardAddress,
	}, nil
}

func (*ClaimDelegationStakeRewards) ComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (*ClaimDelegationStakeRewards) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*ClaimDelegationStakeRewards)(nil)

func (*ClaimDelegationStakeRewards) Size() int {
	return ids.NodeIDLen
}

func (c *ClaimDelegationStakeRewards) Marshal(p *codec.Packer) {
	p.PackFixedBytes(c.NodeID.Bytes())
}

func UnmarshalClaimDelegationStakeRewards(p *codec.Packer) (chain.Action, error) {
	var claimRewards ClaimDelegationStakeRewards
	nodeIDBytes := make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &nodeIDBytes)
	nodeID, err := ids.ToNodeID(nodeIDBytes)
	if err != nil {
		return nil, err
	}
	claimRewards.NodeID = nodeID
	return &claimRewards, p.Err()
}

var (
	_ codec.Typed     = (*ClaimDelegationStakeRewardsResult)(nil)
	_ chain.Marshaler = (*ClaimDelegationStakeRewardsResult)(nil)
)

type ClaimDelegationStakeRewardsResult struct {
	StakeStartBlock    uint64        `serialize:"true" json:"stake_start_block"`
	StakeEndBlock      uint64        `serialize:"true" json:"stake_end_block"`
	StakedAmount       uint64        `serialize:"true" json:"staked_amount"`
	BalanceBeforeClaim uint64        `serialize:"true" json:"balance_before_claim"`
	BalanceAfterClaim  uint64        `serialize:"true" json:"balance_after_claim"`
	DistributedTo      codec.Address `serialize:"true" json:"distributed_to"`
}

func (*ClaimDelegationStakeRewardsResult) GetTypeID() uint8 {
	return nconsts.ClaimDelegationStakeRewards
}

func (*ClaimDelegationStakeRewardsResult) Size() int {
	return 5*consts.Uint64Len + codec.AddressLen
}

func (r *ClaimDelegationStakeRewardsResult) Marshal(p *codec.Packer) {
	p.PackUint64(r.StakeStartBlock)
	p.PackUint64(r.StakeEndBlock)
	p.PackUint64(r.StakedAmount)
	p.PackUint64(r.BalanceBeforeClaim)
	p.PackUint64(r.BalanceAfterClaim)
	p.PackAddress(r.DistributedTo)
}

func UnmarshalClaimDelegationStakeRewardsResult(p *codec.Packer) (codec.Typed, error) {
	var result ClaimDelegationStakeRewardsResult
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.StakedAmount = p.UnpackUint64(true)
	result.BalanceBeforeClaim = p.UnpackUint64(false)
	result.BalanceAfterClaim = p.UnpackUint64(true)
	p.UnpackAddress(&result.DistributedTo)
	return &result, p.Err()
}
