// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

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

var _ chain.Action = (*ClaimValidatorStakeRewards)(nil)

type ClaimValidatorStakeRewards struct {
	NodeID []byte `json:"nodeID"` // Node ID of the validator where NAI
}

func (*ClaimValidatorStakeRewards) GetTypeID() uint8 {
	return nconsts.ClaimValidatorStakeRewardsID
}

func (c *ClaimValidatorStakeRewards) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(c.NodeID)
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
		string(storage.RegisterValidatorStakeKey(nodeID)),
	}
}

func (*ClaimValidatorStakeRewards) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks}
}

func (*ClaimValidatorStakeRewards) OutputsWarpMessage() bool {
	return false
}

func (c *ClaimValidatorStakeRewards) Execute(
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

	// Check whether a validator is trying to claim its reward
	exists, stakeStartBlock, _, _, _, rewardAddress, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if !exists {
		return false, ClaimStakingRewardComputeUnits, OutputStakeMissing, nil, nil
	}
	if rewardAddress != actor {
		return false, ClaimStakingRewardComputeUnits, OutputUnauthorized, nil, nil
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Get last accepted block height
	lastBlockHeight := emissionInstance.GetLastAcceptedBlockHeight()
	// Check that lastBlockHeight is after stakeStartBlock
	if lastBlockHeight < stakeStartBlock {
		return false, ClaimStakingRewardComputeUnits, OutputStakeNotEnded, nil, nil
	}

	// Claim rewards in Emission Balancer
	rewardAmount, err := emissionInstance.ClaimStakingRewards(nodeID, codec.EmptyAddress)
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

func (*ClaimValidatorStakeRewards) MaxComputeUnits(chain.Rules) uint64 {
	return ClaimStakingRewardComputeUnits
}

func (*ClaimValidatorStakeRewards) Size() int {
	return hconsts.NodeIDLen
}

func (c *ClaimValidatorStakeRewards) Marshal(p *codec.Packer) {
	p.PackBytes(c.NodeID)
}

func UnmarshalClaimValidatorStakeRewards(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var claimRewards ClaimValidatorStakeRewards
	p.UnpackBytes(hconsts.NodeIDLen, true, &claimRewards.NodeID)
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
	p := codec.NewReader(b, hconsts.Uint64Len)
	var result ClaimRewardsResult
	result.RewardAmount = p.UnpackUint64(true)
	return &result, p.Err()
}

func (s *ClaimRewardsResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(hconsts.Uint64Len, hconsts.Uint64Len)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
