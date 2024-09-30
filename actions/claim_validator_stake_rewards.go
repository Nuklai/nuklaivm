// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	ClaimStakingRewardComputeUnits = 5
)

var _ chain.Action = (*ClaimValidatorStakeRewards)(nil)

type ClaimValidatorStakeRewards struct {
	NodeID ids.NodeID `serialize:"true" json:"node_id"` // Node ID of the validator where NAI
}

func (*ClaimValidatorStakeRewards) GetTypeID() uint8 {
	return nconsts.ClaimValidatorStakeRewardsID
}

func (c *ClaimValidatorStakeRewards) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):        state.All,
		string(storage.RegisterValidatorStakeKey(c.NodeID)): state.Read,
	}
}

func (*ClaimValidatorStakeRewards) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks}
}

func (c *ClaimValidatorStakeRewards) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check whether a validator is trying to claim its reward
	exists, stakeStartBlock, stakeEndBlock, stakeAmount, delegationFeeRate, rewardAddress, ownerAddress, _ := storage.GetRegisterValidatorStake(ctx, mu, c.NodeID)
	if !exists {
		return nil, ErrStakeMissing
	}
	if rewardAddress != actor {
		return nil, ErrUnauthorizedUser
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check that lastBlockHeight is after stakeEndBlock
	if emissionInstance.GetLastAcceptedBlockHeight() < stakeEndBlock {
		return nil, ErrStakeNotEnded
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(c.NodeID, codec.EmptyAddress)
	if err != nil {
		return nil, err
	}

	balance, err := storage.AddBalance(ctx, mu, rewardAddress, ids.Empty, rewardAmount, true)
	if err != nil {
		return nil, err
	}

	return &ClaimValidatorStakeRewardsResult{
		StakeStartBlock:      stakeStartBlock,
		StakeEndBlock:        stakeEndBlock,
		StakedAmount:         stakeAmount,
		DelegationFeeRate:    delegationFeeRate,
		BalanceBeforeUnstake: balance - rewardAmount,
		BalanceAfterUnstake:  balance,
		DistributedTo:        rewardAddress,
		ValidatorOwner:       ownerAddress,
	}, nil
}

func (*ClaimValidatorStakeRewards) ComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (*ClaimValidatorStakeRewards) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*ClaimValidatorStakeRewards)(nil)

func (*ClaimValidatorStakeRewards) Size() int {
	return ids.NodeIDLen
}

func (c *ClaimValidatorStakeRewards) Marshal(p *codec.Packer) {
	p.PackFixedBytes(c.NodeID.Bytes())
}

func UnmarshalClaimValidatorStakeRewards(p *codec.Packer) (chain.Action, error) {
	var claimRewards ClaimValidatorStakeRewards
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
	_ codec.Typed     = (*ClaimValidatorStakeRewardsResult)(nil)
	_ chain.Marshaler = (*ClaimValidatorStakeRewardsResult)(nil)
)

type ClaimValidatorStakeRewardsResult struct {
	StakeStartBlock      uint64        `serialize:"true" json:"stake_start_block"`
	StakeEndBlock        uint64        `serialize:"true" json:"stake_end_block"`
	StakedAmount         uint64        `serialize:"true" json:"staked_amount"`
	DelegationFeeRate    uint64        `serialize:"true" json:"delegation_fee_rate"`
	BalanceBeforeUnstake uint64        `serialize:"true" json:"balance_before_unstake"`
	BalanceAfterUnstake  uint64        `serialize:"true" json:"balance_after_unstake"`
	DistributedTo        codec.Address `serialize:"true" json:"distributed_to"`
	ValidatorOwner       codec.Address `serialize:"true" json:"validator_owner"`
}

func (*ClaimValidatorStakeRewardsResult) GetTypeID() uint8 {
	return nconsts.ClaimValidatorStakeRewardsID
}

func (*ClaimValidatorStakeRewardsResult) Size() int {
	return 6*consts.Uint64Len + 2*codec.AddressLen
}

func (r *ClaimValidatorStakeRewardsResult) Marshal(p *codec.Packer) {
	p.PackUint64(r.StakeStartBlock)
	p.PackUint64(r.StakeEndBlock)
	p.PackUint64(r.StakedAmount)
	p.PackUint64(r.DelegationFeeRate)
	p.PackUint64(r.BalanceBeforeUnstake)
	p.PackUint64(r.BalanceAfterUnstake)
	p.PackAddress(r.DistributedTo)
	p.PackAddress(r.ValidatorOwner)
}

func UnmarshalClaimValidatorStakeRewardsResult(p *codec.Packer) (codec.Typed, error) {
	var result ClaimValidatorStakeRewardsResult
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.StakedAmount = p.UnpackUint64(true)
	result.DelegationFeeRate = p.UnpackUint64(true)
	result.BalanceBeforeUnstake = p.UnpackUint64(false)
	result.BalanceAfterUnstake = p.UnpackUint64(true)
	p.UnpackAddress(&result.DistributedTo)
	p.UnpackAddress(&result.ValidatorOwner)
	return &result, p.Err()
}
