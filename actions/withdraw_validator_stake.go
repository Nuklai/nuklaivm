// Copyright (C) 2024, Nuklai. All rights reserved.
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

var _ chain.Action = (*WithdrawValidatorStake)(nil)

type WithdrawValidatorStake struct {
	NodeID        []byte        `json:"nodeID"`        // Node ID of the validator
	RewardAddress codec.Address `json:"rewardAddress"` // Address to receive rewards

	// TODO: add boolean to indicate whether sender will
	// create recipient account
}

func (*WithdrawValidatorStake) GetTypeID() uint8 {
	return nconsts.WithdrawValidatorStakeID
}

func (u *WithdrawValidatorStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(u.NodeID)
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):           state.Read | state.Write,
		string(storage.BalanceKey(u.RewardAddress, ids.Empty)): state.All,
		string(storage.RegisterValidatorStakeKey(nodeID)):      state.Read | state.Write,
	}
}

func (*WithdrawValidatorStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks}
}

func (u *WithdrawValidatorStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if it's a valid nodeID
	nodeID, err := ids.ToNodeID(u.NodeID)
	if err != nil {
		return nil, ErrOutputInvalidNodeID
	}

	// Check if the validator was already registered
	exists, _, stakeEndBlock, stakedAmount, _, _, ownerAddress, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if !exists {
		return nil, ErrOutputNotValidator
	}
	if ownerAddress != actor {
		return nil, ErrOutputUnauthorized
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Get last accepted block height
	lastBlockHeight := emissionInstance.GetLastAcceptedBlockHeight()
	// Check that lastBlockTime is after stakeStartBlock
	if lastBlockHeight < stakeEndBlock {
		return nil, ErrOutputStakeNotStarted
	}

	// Withdraw in Emission Balancer
	rewardAmount, err := emissionInstance.WithdrawValidatorStake(nodeID)
	if err != nil {
		return nil, err
	}

	if err := storage.AddBalance(ctx, mu, u.RewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return nil, err
	}
	if err := storage.DeleteRegisterValidatorStake(ctx, mu, nodeID); err != nil {
		return nil, err
	}
	if err := storage.AddBalance(ctx, mu, actor, ids.Empty, stakedAmount, true); err != nil {
		return nil, err
	}

	sr := &WithdrawStakeResult{stakedAmount, rewardAmount}
	output, err := sr.Marshal()
	if err != nil {
		return nil, err
	}

	return [][]byte{output}, nil
}

func (*WithdrawValidatorStake) ComputeUnits(chain.Rules) uint64 {
	return WithdrawValidatorStakeComputeUnits
}

func (*WithdrawValidatorStake) Size() int {
	return ids.NodeIDLen + codec.AddressLen
}

func (u *WithdrawValidatorStake) Marshal(p *codec.Packer) {
	p.PackBytes(u.NodeID)
	p.PackAddress(u.RewardAddress)
}

func UnmarshalWithdrawValidatorStake(p *codec.Packer) (chain.Action, error) {
	var unstake WithdrawValidatorStake
	p.UnpackBytes(ids.NodeIDLen, true, &unstake.NodeID)
	p.UnpackAddress(&unstake.RewardAddress)
	return &unstake, p.Err()
}

func (*WithdrawValidatorStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

type WithdrawStakeResult struct {
	StakedAmount uint64
	RewardAmount uint64
}

func UnmarshalWithdrawValidatorStakeResult(b []byte) (*WithdrawStakeResult, error) {
	p := codec.NewReader(b, 2*consts.Uint64Len)
	var result WithdrawStakeResult
	result.StakedAmount = p.UnpackUint64(true)
	result.RewardAmount = p.UnpackUint64(false)
	return &result, p.Err()
}

func (s *WithdrawStakeResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(2*consts.Uint64Len, 2*consts.Uint64Len)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
