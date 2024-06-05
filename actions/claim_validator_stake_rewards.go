// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*ClaimValidatorStakeRewards)(nil)

type ClaimValidatorStakeRewards struct {
	NodeID []byte `json:"nodeID"` // Node ID of the validator where NAI
}

func (*ClaimValidatorStakeRewards) GetTypeID() uint8 {
	return nconsts.ClaimValidatorStakeRewardsID
}

func (c *ClaimValidatorStakeRewards) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(c.NodeID)
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):      state.All,
		string(storage.RegisterValidatorStakeKey(nodeID)): state.Read,
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
) ([][]byte, error) {
	nodeID, err := ids.ToNodeID(c.NodeID)
	if err != nil {
		return nil, ErrInvalidNodeID
	}

	// Check whether a validator is trying to claim its reward
	exists, _, stakeEndBlock, _, _, rewardAddress, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if !exists {
		return nil, ErrStakeMissing
	}
	if rewardAddress != actor {
		return nil, ErrUnauthorized
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check that lastBlockHeight is after stakeEndBlock
	if emissionInstance.GetLastAcceptedBlockHeight() < stakeEndBlock {
		return nil, ErrStakeNotEnded
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(nodeID, codec.EmptyAddress)
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

func (*ClaimValidatorStakeRewards) ComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (*ClaimValidatorStakeRewards) Size() int {
	return ids.NodeIDLen
}

func (c *ClaimValidatorStakeRewards) Marshal(p *codec.Packer) {
	p.PackBytes(c.NodeID)
}

func UnmarshalClaimValidatorStakeRewards(p *codec.Packer) (chain.Action, error) {
	var claimRewards ClaimValidatorStakeRewards
	p.UnpackBytes(ids.NodeIDLen, true, &claimRewards.NodeID)
	return &claimRewards, p.Err()
}

func (*ClaimValidatorStakeRewards) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

type ClaimRewardsResult struct {
	RewardAmount uint64
}

func UnmarshalClaimRewardsResult(b []byte) (*ClaimRewardsResult, error) {
	p := codec.NewReader(b, consts.Uint64Len)
	var result ClaimRewardsResult
	result.RewardAmount = p.UnpackUint64(false)
	return &result, p.Err()
}

func (s *ClaimRewardsResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(consts.Uint64Len, consts.Uint64Len)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
