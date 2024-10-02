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
	UndelegateUserStakeComputeUnits = 5
)

var _ chain.Action = (*UndelegateUserStake)(nil)

type UndelegateUserStake struct {
	NodeID ids.NodeID `serialize:"true" json:"node_id"` // Node ID of the validator where NAI is staked
}

func (*UndelegateUserStake) GetTypeID() uint8 {
	return nconsts.UndelegateUserStakeID
}

func (u *UndelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):          state.Read | state.Write,
		string(storage.DelegateUserStakeKey(actor, u.NodeID)): state.Read | state.Write,
	}
}

func (*UndelegateUserStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.DelegateUserStakeChunks}
}

func (*UndelegateUserStake) OutputsWarpMessage() bool {
	return false
}

func (u *UndelegateUserStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	exists, stakeStartBlock, stakeEndBlock, stakedAmount, _, ownerAddress, _ := storage.GetDelegateUserStake(ctx, mu, actor, u.NodeID)
	if !exists {
		return nil, ErrStakeMissing
	}
	if ownerAddress != actor {
		return nil, ErrUnauthorizedUser
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check that lastBlockHeight is after stakeEndBlock
	if emissionInstance.GetLastAcceptedBlockHeight() < stakeEndBlock {
		return nil, ErrStakeNotEnded
	}

	// Undelegate in Emission Balancer
	rewardAmount, err := emissionInstance.UndelegateUserStake(u.NodeID, actor)
	if err != nil {
		return nil, err
	}
	if err := storage.DeleteDelegateUserStake(ctx, mu, actor, u.NodeID); err != nil {
		return nil, err
	}
	balance, err := storage.AddBalance(ctx, mu, actor, ids.Empty, rewardAmount+stakedAmount, true)
	if err != nil {
		return nil, err
	}

	return &UndelegateUserStakeResult{
		StakeStartBlock:      stakeStartBlock,
		StakeEndBlock:        stakeEndBlock,
		UnstakedAmount:       stakedAmount,
		RewardAmount:         rewardAmount,
		BalanceBeforeUnstake: balance - rewardAmount - stakedAmount,
		BalanceAfterUnstake:  balance,
		DistributedTo:        actor,
	}, nil
}

func (*UndelegateUserStake) ComputeUnits(chain.Rules) uint64 {
	return UndelegateUserStakeComputeUnits
}

func (*UndelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*UndelegateUserStake)(nil)

func (*UndelegateUserStake) Size() int {
	return ids.NodeIDLen
}

func (u *UndelegateUserStake) Marshal(p *codec.Packer) {
	p.PackFixedBytes(u.NodeID.Bytes())
}

func UnmarshalUndelegateUserStake(p *codec.Packer) (chain.Action, error) {
	var unstake UndelegateUserStake
	nodeIDBytes := make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &nodeIDBytes)
	nodeID, err := ids.ToNodeID(nodeIDBytes)
	if err != nil {
		return nil, err
	}
	unstake.NodeID = nodeID
	return &unstake, p.Err()
}

var (
	_ codec.Typed     = (*UndelegateUserStakeResult)(nil)
	_ chain.Marshaler = (*UndelegateUserStakeResult)(nil)
)

type UndelegateUserStakeResult struct {
	StakeStartBlock      uint64        `serialize:"true" json:"stake_start_block"`
	StakeEndBlock        uint64        `serialize:"true" json:"stake_end_block"`
	UnstakedAmount       uint64        `serialize:"true" json:"unstaked_amount"`
	RewardAmount         uint64        `serialize:"true" json:"reward_amount"`
	BalanceBeforeUnstake uint64        `serialize:"true" json:"balance_before_unstake"`
	BalanceAfterUnstake  uint64        `serialize:"true" json:"balance_after_unstake"`
	DistributedTo        codec.Address `serialize:"true" json:"distributed_to"`
}

func (*UndelegateUserStakeResult) GetTypeID() uint8 {
	return nconsts.UndelegateUserStakeID
}

func (*UndelegateUserStakeResult) Size() int {
	return 6*consts.Uint64Len + codec.AddressLen
}

func (r *UndelegateUserStakeResult) Marshal(p *codec.Packer) {
	p.PackUint64(r.StakeStartBlock)
	p.PackUint64(r.StakeEndBlock)
	p.PackUint64(r.UnstakedAmount)
	p.PackUint64(r.RewardAmount)
	p.PackUint64(r.BalanceBeforeUnstake)
	p.PackUint64(r.BalanceAfterUnstake)
	p.PackAddress(r.DistributedTo)
}

func UnmarshalUndelegateUserStakeResult(p *codec.Packer) (codec.Typed, error) {
	var result UndelegateUserStakeResult
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.UnstakedAmount = p.UnpackUint64(false)
	result.RewardAmount = p.UnpackUint64(false)
	result.BalanceBeforeUnstake = p.UnpackUint64(false)
	result.BalanceAfterUnstake = p.UnpackUint64(true)
	p.UnpackAddress(&result.DistributedTo)
	return &result, p.Err()
}
