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
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
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

func (c *ClaimValidatorStakeRewards) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.ValidatorStakeKey(c.NodeID)):                       state.Read,
		string(storage.AssetAccountBalanceKey(storage.NAIAddress, actor)): state.All,
	}
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
	exists, stakeStartBlock, stakeEndBlock, stakeAmount, delegationFeeRate, rewardAddress, _, _ := storage.GetValidatorStakeNoController(ctx, mu, c.NodeID)
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
		return nil, ErrStakeNotStarted
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(c.NodeID, codec.EmptyAddress)
	if err != nil {
		return nil, err
	}

	// Get the reward
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, storage.NAIAddress, rewardAddress)
	if err != nil {
		return nil, err
	}
	newBalance, err := smath.Add(balance, rewardAmount)
	if err != nil {
		return nil, err
	}
	if err = storage.SetAssetAccountBalance(ctx, mu, storage.NAIAddress, rewardAddress, newBalance); err != nil {
		return nil, err
	}

	return &ClaimValidatorStakeRewardsResult{
		Actor:              actor.String(),
		Receiver:           actor.String(),
		StakeStartBlock:    stakeStartBlock,
		StakeEndBlock:      stakeEndBlock,
		StakedAmount:       stakeAmount,
		DelegationFeeRate:  delegationFeeRate,
		BalanceBeforeClaim: balance,
		BalanceAfterClaim:  newBalance,
		DistributedTo:      rewardAddress.String(),
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

var _ codec.Typed = (*ClaimValidatorStakeRewardsResult)(nil)

type ClaimValidatorStakeRewardsResult struct {
	Actor              string `serialize:"true" json:"actor"`
	Receiver           string `serialize:"true" json:"receiver"`
	StakeStartBlock    uint64 `serialize:"true" json:"stake_start_block"`
	StakeEndBlock      uint64 `serialize:"true" json:"stake_end_block"`
	StakedAmount       uint64 `serialize:"true" json:"staked_amount"`
	DelegationFeeRate  uint64 `serialize:"true" json:"delegation_fee_rate"`
	BalanceBeforeClaim uint64 `serialize:"true" json:"balance_before_claim"`
	BalanceAfterClaim  uint64 `serialize:"true" json:"balance_after_claim"`
	DistributedTo      string `serialize:"true" json:"distributed_to"`
}

func (*ClaimValidatorStakeRewardsResult) GetTypeID() uint8 {
	return nconsts.ClaimValidatorStakeRewardsID
}

func UnmarshalClaimValidatorStakeRewardsResult(p *codec.Packer) (codec.Typed, error) {
	var result ClaimValidatorStakeRewardsResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.StakedAmount = p.UnpackUint64(true)
	result.DelegationFeeRate = p.UnpackUint64(true)
	result.BalanceBeforeClaim = p.UnpackUint64(false)
	result.BalanceAfterClaim = p.UnpackUint64(true)
	result.DistributedTo = p.UnpackString(true)
	return &result, p.Err()
}
