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
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*DelegateUserStake)(nil)

type DelegateUserStake struct {
	NodeID          []byte        `json:"nodeID"`          // Node ID of the validator to stake to
	StakeStartBlock uint64        `json:"stakeStartBlock"` // Block height at which the stake should be made
	StakeEndBlock   uint64        `json:"stakeEndBlock"`   // Block height at which the stake should end
	StakedAmount    uint64        `json:"stakedAmount"`    // Amount of NAI staked
	RewardAddress   codec.Address `json:"rewardAddress"`   // Address to receive rewards
}

func (*DelegateUserStake) GetTypeID() uint8 {
	return nconsts.DelegateUserStakeID
}

func (s *DelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	nodeID, _ := ids.ToNodeID(s.NodeID)
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):        state.Read | state.Write,
		string(storage.DelegateUserStakeKey(actor, nodeID)): state.Allocate | state.Write,
		string(storage.RegisterValidatorStakeKey(nodeID)):   state.Read,
	}
}

func (*DelegateUserStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.DelegateUserStakeChunks, storage.RegisterValidatorStakeChunks}
}

func (s *DelegateUserStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	nodeID, err := ids.ToNodeID(s.NodeID)
	if err != nil {
		return nil, ErrOutputInvalidNodeID
	}
	if _, err := codec.AddressBech32(nconsts.HRP, s.RewardAddress); err != nil {
		return nil, err
	}

	// Check if the validator the user is trying to delegate to is registered for staking
	exists, stakeStartBlock, stakeEndBlock, _, _, _, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if !exists {
		return nil, ErrOutputValidatorNotYetRegistered
	}

	// Check if the user has already delegated to this validator node before
	exists, _, _, _, _, _, _ = storage.GetDelegateUserStake(ctx, mu, actor, nodeID)
	if exists {
		return nil, ErrOutputUserAlreadyStaked
	}

	stakingConfig := emission.GetStakingConfig()

	// Check if the staked amount is a valid amount
	if s.StakedAmount < stakingConfig.MinDelegatorStake {
		return nil, ErrOutputDelegateStakedAmountInvalid
	}

	// Check if stakeStartBlock is greater than stakeEndBlock
	if s.StakeStartBlock >= s.StakeEndBlock {
		return nil, ErrOutputInvalidStakeEndBlock
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check if stakeStartBlock is smaller than the current block height
	if s.StakeStartBlock < emissionInstance.GetLastAcceptedBlockHeight() || s.StakeStartBlock >= stakeEndBlock {
		return nil, ErrOutputInvalidStakeStartBlock
	}

	// Check if the delegator stakeStartBlock is smaller than the validator stakeStartBlock and if so, set the delegator stakeStartBlock to the validator stakeStartBlock
	if s.StakeStartBlock < stakeStartBlock {
		s.StakeStartBlock = stakeStartBlock
	}

	// Check if the delegator stakeEndBlock is greater than the validator stakeEndBlock
	if s.StakeEndBlock > stakeEndBlock {
		s.StakeEndBlock = stakeEndBlock
	}

	// Delegate in Emission Balancer
	err = emissionInstance.DelegateUserStake(nodeID, actor, s.StakeStartBlock, s.StakeEndBlock, s.StakedAmount)
	if err != nil {
		return nil, err
	}

	if err := storage.SubBalance(ctx, mu, actor, ids.Empty, s.StakedAmount); err != nil {
		return nil, err
	}
	if err := storage.SetDelegateUserStake(ctx, mu, actor, nodeID, s.StakeStartBlock, s.StakeEndBlock, s.StakedAmount, s.RewardAddress); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*DelegateUserStake) ComputeUnits(chain.Rules) uint64 {
	return DelegateUserStakeComputeUnits
}

func (*DelegateUserStake) Size() int {
	return ids.NodeIDLen + 3*consts.Uint64Len + codec.AddressLen
}

func (s *DelegateUserStake) Marshal(p *codec.Packer) {
	p.PackBytes(s.NodeID)
	p.PackUint64(s.StakeStartBlock)
	p.PackUint64(s.StakeEndBlock)
	p.PackUint64(s.StakedAmount)
	p.PackAddress(s.RewardAddress)
}

func UnmarshalDelegateUserStake(p *codec.Packer) (chain.Action, error) {
	var stake DelegateUserStake
	p.UnpackBytes(ids.NodeIDLen, true, &stake.NodeID)
	stake.StakeStartBlock = p.UnpackUint64(true)
	stake.StakeEndBlock = p.UnpackUint64(true)
	stake.StakedAmount = p.UnpackUint64(true)
	p.UnpackAddress(&stake.RewardAddress)
	return &stake, p.Err()
}

func (*DelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
