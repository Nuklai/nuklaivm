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

var _ chain.Action = (*UndelegateUserStake)(nil)

type UndelegateUserStake struct {
	NodeID        []byte        `json:"nodeID"`        // Node ID of the validator where NAI is staked
	RewardAddress codec.Address `json:"rewardAddress"` // Address to receive rewards

	// TODO: add boolean to indicate whether sender will
	// create recipient account
}

func (*UndelegateUserStake) GetTypeID() uint8 {
	return nconsts.UndelegateUserStakeID
}

func (u *UndelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(u.NodeID)
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):           state.Read | state.Write,
		string(storage.BalanceKey(u.RewardAddress, ids.Empty)): state.All,
		string(storage.DelegateUserStakeKey(actor, nodeID)):    state.Read | state.Write,
	}
}

func (*UndelegateUserStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.BalanceChunks, storage.DelegateUserStakeChunks}
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
) ([][]byte, error) {
	nodeID, err := ids.ToNodeID(u.NodeID)
	if err != nil {
		return nil, ErrOutputInvalidNodeID
	}

	exists, _, stakeEndBlock, stakedAmount, _, ownerAddress, _ := storage.GetDelegateUserStake(ctx, mu, actor, nodeID)
	if !exists {
		return nil, ErrOutputStakeMissing
	}
	if ownerAddress != actor {
		return nil, ErrOutputUnauthorized
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check that lastBlockHeight is after stakeEndBlock
	if emissionInstance.GetLastAcceptedBlockHeight() < stakeEndBlock {
		return nil, ErrOutputStakeNotEnded
	}

	// Undelegate in Emission Balancer
	rewardAmount, err := emissionInstance.UndelegateUserStake(nodeID, actor)
	if err != nil {
		return nil, err
	}
	if err := storage.AddBalance(ctx, mu, u.RewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return nil, err
	}

	if err := storage.DeleteDelegateUserStake(ctx, mu, ownerAddress, nodeID); err != nil {
		return nil, err
	}
	if err := storage.AddBalance(ctx, mu, ownerAddress, ids.Empty, stakedAmount, true); err != nil {
		return nil, err
	}

	sr := &UndelegateUserStakeResult{stakedAmount, rewardAmount}
	output, err := sr.Marshal()
	if err != nil {
		return nil, err
	}
	return [][]byte{output}, nil
}

func (*UndelegateUserStake) ComputeUnits(chain.Rules) uint64 {
	return UndelegateUserStakeComputeUnits
}

func (*UndelegateUserStake) Size() int {
	return ids.NodeIDLen + codec.AddressLen
}

func (u *UndelegateUserStake) Marshal(p *codec.Packer) {
	p.PackBytes(u.NodeID)
	p.PackAddress(u.RewardAddress)
}

func UnmarshalUndelegateUserStake(p *codec.Packer) (chain.Action, error) {
	var unstake UndelegateUserStake
	p.UnpackBytes(ids.NodeIDLen, true, &unstake.NodeID)
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
	p := codec.NewReader(b, 2*consts.Uint64Len)
	var result UndelegateUserStakeResult
	result.StakedAmount = p.UnpackUint64(true)
	result.RewardAmount = p.UnpackUint64(false)
	return &result, p.Err()
}

func (s *UndelegateUserStakeResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(2*consts.Uint64Len, 2*consts.Uint64Len)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
