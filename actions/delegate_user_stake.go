// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	DelegateUserStakeComputeUnits = 5
)

var (
	ErrValidatorNotYetRegistered                = errors.New("validator is not yet registered")
	ErrUserAlreadyStaked                        = errors.New("user has already staked to this validator")
	ErrDelegateStakedAmountInvalid              = errors.New("staked amount is invalid")
	_                              chain.Action = (*DelegateUserStake)(nil)
)

type DelegateUserStake struct {
	NodeID          ids.NodeID `serialize:"true" json:"node_id"`           // Node ID of the validator to stake to
	StakeStartBlock uint64     `serialize:"true" json:"stake_start_block"` // Block height at which the stake should be made
	StakeEndBlock   uint64     `serialize:"true" json:"stake_end_block"`   // Block height at which the stake should end
	StakedAmount    uint64     `serialize:"true" json:"staked_amount"`     // Amount of NAI staked
}

func (*DelegateUserStake) GetTypeID() uint8 {
	return nconsts.DelegateUserStakeID
}

func (s *DelegateUserStake) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.DelegatorStakeKey(actor, s.NodeID)):                state.Allocate | state.Write,
		string(storage.ValidatorStakeKey(s.NodeID)):                       state.Read,
		string(storage.AssetAccountBalanceKey(storage.NAIAddress, actor)): state.Read | state.Write,
	}
}

func (s *DelegateUserStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the validator the user is trying to delegate to is registered for staking
	exists, stakeStartBlock, stakeEndBlock, _, _, _, _, _ := storage.GetValidatorStakeNoController(ctx, mu, s.NodeID)
	if !exists {
		return nil, ErrValidatorNotYetRegistered
	}

	// Check if the user has already delegated to this validator node before
	exists, _, _, _, _, _, _ = storage.GetDelegatorStakeNoController(ctx, mu, actor, s.NodeID)
	if exists {
		return nil, ErrUserAlreadyStaked
	}

	stakingConfig := emission.GetStakingConfig()

	// Check if the staked amount is a valid amount
	if s.StakedAmount < stakingConfig.MinDelegatorStake {
		return nil, ErrDelegateStakedAmountInvalid
	}

	// Check if stakeStartBlock is greater than stakeEndBlock
	if s.StakeStartBlock >= s.StakeEndBlock {
		return nil, ErrInvalidStakeEndBlock
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Check if stakeStartBlock is smaller than the current block height
	if s.StakeStartBlock < emissionInstance.GetLastAcceptedBlockHeight() || s.StakeStartBlock >= stakeEndBlock {
		return nil, ErrInvalidStakeStartBlock
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
	err := emissionInstance.DelegateUserStake(s.NodeID, actor, s.StakeStartBlock, s.StakeEndBlock, s.StakedAmount)
	if err != nil {
		return nil, err
	}

	// Ensure that the balance is sufficient and subtract the staked amount
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, storage.NAIAddress, actor)
	if err != nil {
		return nil, err
	}
	if balance < s.StakedAmount {
		return nil, storage.ErrInsufficientAssetBalance
	}
	newBalance, err := smath.Sub(balance, s.StakedAmount)
	if err != nil {
		return nil, err
	}
	if err = storage.SetAssetAccountBalance(ctx, mu, storage.NAIAddress, actor, newBalance); err != nil {
		return nil, err
	}

	if err := storage.SetDelegatorStake(ctx, mu, actor, s.NodeID, s.StakeStartBlock, s.StakeEndBlock, s.StakedAmount, actor); err != nil {
		return nil, err
	}
	return &DelegateUserStakeResult{
		Actor:              actor.String(),
		Receiver:           s.NodeID.String(),
		StakedAmount:       s.StakedAmount,
		BalanceBeforeStake: balance,
		BalanceAfterStake:  newBalance,
	}, nil
}

func (*DelegateUserStake) ComputeUnits(chain.Rules) uint64 {
	return DelegateUserStakeComputeUnits
}

func (*DelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*DelegateUserStake)(nil)

func (*DelegateUserStake) Size() int {
	return ids.NodeIDLen + 3*consts.Uint64Len
}

func (s *DelegateUserStake) Marshal(p *codec.Packer) {
	p.PackFixedBytes(s.NodeID.Bytes())
	p.PackUint64(s.StakeStartBlock)
	p.PackUint64(s.StakeEndBlock)
	p.PackUint64(s.StakedAmount)
}

func UnmarshalDelegateUserStake(p *codec.Packer) (chain.Action, error) {
	var stake DelegateUserStake
	nodeIDBytes := make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &nodeIDBytes)
	nodeID, err := ids.ToNodeID(nodeIDBytes)
	if err != nil {
		return nil, err
	}
	stake.NodeID = nodeID
	stake.StakeStartBlock = p.UnpackUint64(true)
	stake.StakeEndBlock = p.UnpackUint64(true)
	stake.StakedAmount = p.UnpackUint64(true)
	return &stake, p.Err()
}

var _ codec.Typed = (*DelegateUserStakeResult)(nil)

type DelegateUserStakeResult struct {
	Actor              string `serialize:"true" json:"actor"`
	Receiver           string `serialize:"true" json:"receiver"`
	StakedAmount       uint64 `serialize:"true" json:"staked_amount"`
	BalanceBeforeStake uint64 `serialize:"true" json:"balance_before_stake"`
	BalanceAfterStake  uint64 `serialize:"true" json:"balance_after_stake"`
}

func (*DelegateUserStakeResult) GetTypeID() uint8 {
	return nconsts.DelegateUserStakeID
}

func UnmarshalDelegateUserStakeResult(p *codec.Packer) (codec.Typed, error) {
	var result DelegateUserStakeResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.StakedAmount = p.UnpackUint64(true)
	result.BalanceBeforeStake = p.UnpackUint64(true)
	result.BalanceAfterStake = p.UnpackUint64(false)
	return &result, p.Err()
}
