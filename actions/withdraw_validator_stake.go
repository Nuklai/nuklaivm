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

var _ chain.Action = (*WithdrawValidatorStake)(nil)

type WithdrawValidatorStake struct {
	NodeID        []byte        `json:"nodeID"`        // Node ID of the validator
	RewardAddress codec.Address `json:"rewardAddress"` // Address to receive rewards
}

func (*WithdrawValidatorStake) GetTypeID() uint8 {
	return nconsts.WithdrawValidatorStakeID
}

func (u *WithdrawValidatorStake) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	nodeID, _ := ids.ToNodeID(u.NodeID)
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
		string(storage.BalanceKey(u.RewardAddress, ids.Empty)),
		string(storage.RegisterValidatorStakeKey(nodeID)),
	}
}

func (*WithdrawValidatorStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks}
}

func (*WithdrawValidatorStake) OutputsWarpMessage() bool {
	return false
}

func (u *WithdrawValidatorStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	// Check if it's a valid nodeID
	nodeID, err := ids.ToNodeID(u.NodeID)
	if err != nil {
		return false, WithdrawValidatorStakeComputeUnits, OutputInvalidNodeID, nil, nil
	}

	// Check if the validator was already registered
	exists, _, stakeEndBlock, stakedAmount, _, _, ownerAddress, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if !exists {
		return false, WithdrawValidatorStakeComputeUnits, OutputValidatorAlreadyRegistered, nil, nil
	}
	if ownerAddress != actor {
		return false, WithdrawValidatorStakeComputeUnits, OutputUnauthorized, nil, nil
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Get last accepted block height
	lastBlockHeight := emissionInstance.GetLastAcceptedBlockHeight()
	// Check that lastBlockTime is after stakeStartBlock
	if lastBlockHeight < stakeEndBlock {
		return false, WithdrawValidatorStakeComputeUnits, OutputStakeNotStarted, nil, nil
	}

	// Withdraw in Emission Balancer
	rewardAmount, err := emissionInstance.WithdrawValidatorStake(nodeID)
	if err != nil {
		return false, WithdrawValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}

	if err := storage.AddBalance(ctx, mu, u.RewardAddress, ids.Empty, rewardAmount, true); err != nil {
		return false, WithdrawValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.DeleteRegisterValidatorStake(ctx, mu, nodeID); err != nil {
		return false, WithdrawValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.AddBalance(ctx, mu, actor, ids.Empty, stakedAmount, true); err != nil {
		return false, WithdrawValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}

	sr := &WithdrawStakeResult{stakedAmount, rewardAmount}
	output, err := sr.Marshal()
	if err != nil {
		return false, WithdrawValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}

	return true, WithdrawValidatorStakeComputeUnits, output, nil, nil
}

func (*WithdrawValidatorStake) MaxComputeUnits(chain.Rules) uint64 {
	return WithdrawValidatorStakeComputeUnits
}

func (*WithdrawValidatorStake) Size() int {
	return hconsts.NodeIDLen + codec.AddressLen
}

func (u *WithdrawValidatorStake) Marshal(p *codec.Packer) {
	p.PackBytes(u.NodeID)
	p.PackAddress(u.RewardAddress)
}

func UnmarshalWithdrawValidatorStake(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var unstake WithdrawValidatorStake
	p.UnpackBytes(hconsts.NodeIDLen, true, &unstake.NodeID)
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
	p := codec.NewReader(b, 2*hconsts.Uint64Len)
	var result WithdrawStakeResult
	result.StakedAmount = p.UnpackUint64(true)
	result.RewardAmount = p.UnpackUint64(true)
	return &result, p.Err()
}

func (s *WithdrawStakeResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(2*hconsts.Uint64Len, 2*hconsts.Uint64Len)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
