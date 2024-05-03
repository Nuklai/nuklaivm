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

var _ chain.Action = (*UndelegateUserStake)(nil)

type UndelegateUserStake struct {
	NodeID        []byte        `json:"nodeID"`        // Node ID of the validator where NAI is staked
	RewardAddress codec.Address `json:"rewardAddress"` // Address to receive rewards
}

func (*UndelegateUserStake) GetTypeID() uint8 {
	return nconsts.UndelegateUserStakeID
}

func (u *UndelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(u.NodeID)
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
		string(storage.BalanceKey(u.RewardAddress, ids.Empty)),
		string(storage.DelegateUserStakeKey(actor, nodeID)),
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
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	nodeID, err := ids.ToNodeID(u.NodeID)
	if err != nil {
		return false, UndelegateUserStakeComputeUnits, OutputInvalidNodeID, nil, nil
	}

	exists, stakeStartBlock, stakedAmount, _, ownerAddress, _ := storage.GetDelegateUserStake(ctx, mu, actor, nodeID)
	if !exists {
		return false, UndelegateUserStakeComputeUnits, OutputStakeMissing, nil, nil
	}
	if ownerAddress != actor {
		return false, UndelegateUserStakeComputeUnits, OutputUnauthorized, nil, nil
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Get last accepted block height
	lastBlockHeight := emissionInstance.GetLastAcceptedBlockHeight()
	// Check that lastBlockHeight is after stakeStartBlock
	if lastBlockHeight < stakeStartBlock {
		return false, UndelegateUserStakeComputeUnits, OutputStakeNotStarted, nil, nil
	}

	// Undelegate in Emission Balancer
	rewardAmount, err := emissionInstance.UndelegateUserStake(nodeID, actor, stakedAmount)
	if err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.AddBalance(ctx, mu, u.RewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}

	if err := storage.DeleteDelegateUserStake(ctx, mu, ownerAddress, nodeID); err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.AddBalance(ctx, mu, ownerAddress, ids.Empty, stakedAmount, true); err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}

	sr := &UndelegateUserStakeResult{stakedAmount, rewardAmount}
	output, err := sr.Marshal()
	if err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, UndelegateUserStakeComputeUnits, output, nil, nil
}

func (*UndelegateUserStake) MaxComputeUnits(chain.Rules) uint64 {
	return UndelegateUserStakeComputeUnits
}

func (*UndelegateUserStake) Size() int {
	return hconsts.NodeIDLen + codec.AddressLen
}

func (u *UndelegateUserStake) Marshal(p *codec.Packer) {
	p.PackBytes(u.NodeID)
	p.PackAddress(u.RewardAddress)
}

func UnmarshalUndelegateUserStake(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var unstake UndelegateUserStake
	p.UnpackBytes(hconsts.NodeIDLen, true, &unstake.NodeID)
	p.UnpackAddress(&unstake.RewardAddress)
	return &unstake, p.Err()
}

func (*UndelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

type UndelegateUserStakeResult struct {
	StakedAmount uint64
	RewardAmount uint64
}

func UnmarshalUndelegateUserStakeResult(b []byte) (*UndelegateUserStakeResult, error) {
	p := codec.NewReader(b, 2*hconsts.Uint64Len)
	var result UndelegateUserStakeResult
	result.StakedAmount = p.UnpackUint64(true)
	result.RewardAmount = p.UnpackUint64(true)
	return &result, p.Err()
}

func (s *UndelegateUserStakeResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(2*hconsts.Uint64Len, 2*hconsts.Uint64Len)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
