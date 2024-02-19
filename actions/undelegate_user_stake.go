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
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*UndelegateUserStake)(nil)

type UndelegateUserStake struct {
	NodeID []byte `json:"nodeID"` // Node ID of the validator where NAI is staked
}

func (*UndelegateUserStake) GetTypeID() uint8 {
	return nconsts.UndelegateUserStakeID
}

func (u *UndelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(u.NodeID)
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
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
	timestamp int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	nodeID, err := ids.ToNodeID(u.NodeID)
	if err != nil {
		return false, UndelegateUserStakeComputeUnits, OutputInvalidNodeID, nil, nil
	}

	exists, _, stakeEndTime, stakedAmount, _, ownerAddress, _ := storage.GetDelegateUserStake(ctx, mu, actor, nodeID)
	if !exists {
		return false, UndelegateUserStakeComputeUnits, OutputStakeMissing, nil, nil
	}
	if ownerAddress != actor {
		return false, UndelegateUserStakeComputeUnits, OutputUnauthorized, nil, nil
	}

	// Get current time
	currentTime := time.Unix(int64(timestamp), 0).UTC()
	// Convert Unix timestamps to Go's time.Time for easier manipulation
	endTime := time.Unix(int64(stakeEndTime), 0).UTC()
	// Check that currentTime is after stakeEndTime
	if currentTime.Before(endTime) {
		return false, UndelegateUserStakeComputeUnits, OutputStakeNotEnded, nil, nil
	}

	if err := storage.DeleteDelegateUserStake(ctx, mu, ownerAddress, nodeID); err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.AddBalance(ctx, mu, ownerAddress, ids.Empty, stakedAmount, true); err != nil {
		return false, UndelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	// TODO: Claim Delegated Staking Rewards

	sr := &DelegateStakeResult{stakedAmount, stakeEndTime}
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
	return hconsts.NodeIDLen
}

func (u *UndelegateUserStake) Marshal(p *codec.Packer) {
	p.PackBytes(u.NodeID)
}

func UnmarshalUndelegateUserStake(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var unstake UndelegateUserStake
	p.UnpackBytes(hconsts.NodeIDLen, true, &unstake.NodeID)
	return &unstake, p.Err()
}

func (*UndelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

type DelegateStakeResult struct {
	StakedAmount uint64
	StakeEndTime uint64
}

func UnmarshalDelegateUserStakeResult(b []byte) (*DelegateStakeResult, error) {
	p := codec.NewReader(b, hconsts.Uint64Len*2)
	var result DelegateStakeResult
	result.StakedAmount = p.UnpackUint64(true)
	result.StakeEndTime = p.UnpackUint64(true)
	return &result, p.Err()
}

func (s *DelegateStakeResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(hconsts.Uint64Len*2, hconsts.Uint64Len*2)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.StakeEndTime)
	return p.Bytes(), p.Err()
}