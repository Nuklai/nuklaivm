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
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	WithdrawValidatorStakeComputeUnits = 5
)

var (
	ErrNotValidator                 = errors.New("node is not a validator")
	ErrStakeNotStarted              = errors.New("stake not started")
	ErrStakeNotEnded                = errors.New("stake has not ended")
	_                  chain.Action = (*WithdrawValidatorStake)(nil)
)

type WithdrawValidatorStake struct {
	NodeID ids.NodeID `serialize:"true" json:"node_id"` // Node ID of the validator
}

func (*WithdrawValidatorStake) GetTypeID() uint8 {
	return nconsts.WithdrawValidatorStakeID
}

func (u *WithdrawValidatorStake) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.ValidatorStakeKey(u.NodeID)):                       state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(storage.NAIAddress, actor)): state.Read | state.Write,
	}
}

func (u *WithdrawValidatorStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the validator was already registered
	exists, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, _, ownerAddress, _ := storage.GetValidatorStakeNoController(ctx, mu, u.NodeID)
	if !exists {
		return nil, ErrNotValidator
	}
	if ownerAddress != actor {
		return nil, ErrNotValidatorOwner
	}

	// Get the emission instance
	emissionInstance := emission.GetEmission()

	// Get last accepted block height
	lastBlockHeight := emissionInstance.GetLastAcceptedBlockHeight()
	// Check that lastBlockTime is after stakeStartBlock
	if lastBlockHeight < stakeEndBlock {
		return nil, ErrStakeNotStarted
	}

	// Withdraw in Emission Balancer
	rewardAmount, err := emissionInstance.WithdrawValidatorStake(u.NodeID)
	if err != nil {
		return nil, err
	}

	if err := storage.DeleteValidatorStake(ctx, mu, u.NodeID); err != nil {
		return nil, err
	}

	// Get the reward + staked amount
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, storage.NAIAddress, actor)
	if err != nil {
		return nil, err
	}
	newBalance, err := smath.Add(balance, rewardAmount+stakedAmount)
	if err != nil {
		return nil, err
	}
	if err = storage.SetAssetAccountBalance(ctx, mu, storage.NAIAddress, actor, newBalance); err != nil {
		return nil, err
	}

	return &WithdrawValidatorStakeResult{
		Actor:                actor.String(),
		Receiver:             actor.String(),
		StakeStartBlock:      stakeStartBlock,
		StakeEndBlock:        stakeEndBlock,
		UnstakedAmount:       stakedAmount,
		DelegationFeeRate:    delegationFeeRate,
		RewardAmount:         rewardAmount,
		BalanceBeforeUnstake: balance,
		BalanceAfterUnstake:  newBalance,
		DistributedTo:        actor.String(),
	}, nil
}

func (*WithdrawValidatorStake) ComputeUnits(chain.Rules) uint64 {
	return WithdrawValidatorStakeComputeUnits
}

func (*WithdrawValidatorStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*WithdrawValidatorStake)(nil)

func (*WithdrawValidatorStake) Size() int {
	return ids.NodeIDLen
}

func (u *WithdrawValidatorStake) Marshal(p *codec.Packer) {
	p.PackFixedBytes(u.NodeID.Bytes())
}

func UnmarshalWithdrawValidatorStake(p *codec.Packer) (chain.Action, error) {
	var unstake WithdrawValidatorStake
	nodeIDBytes := make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &nodeIDBytes)
	nodeID, err := ids.ToNodeID(nodeIDBytes)
	if err != nil {
		return nil, err
	}
	unstake.NodeID = nodeID
	return &unstake, p.Err()
}

var _ codec.Typed = (*WithdrawValidatorStakeResult)(nil)

type WithdrawValidatorStakeResult struct {
	Actor                string `serialize:"true" json:"actor"`
	Receiver             string `serialize:"true" json:"receiver"`
	StakeStartBlock      uint64 `serialize:"true" json:"stake_start_block"`
	StakeEndBlock        uint64 `serialize:"true" json:"stake_end_block"`
	UnstakedAmount       uint64 `serialize:"true" json:"unstaked_amount"`
	DelegationFeeRate    uint64 `serialize:"true" json:"delegation_fee_rate"`
	RewardAmount         uint64 `serialize:"true" json:"reward_amount"`
	BalanceBeforeUnstake uint64 `serialize:"true" json:"balance_before_unstake"`
	BalanceAfterUnstake  uint64 `serialize:"true" json:"balance_after_unstake"`
	DistributedTo        string `serialize:"true" json:"distributed_to"`
}

func (*WithdrawValidatorStakeResult) GetTypeID() uint8 {
	return nconsts.WithdrawValidatorStakeID
}

func UnmarshalWithdrawValidatorStakeResult(p *codec.Packer) (codec.Typed, error) {
	var result WithdrawValidatorStakeResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.UnstakedAmount = p.UnpackUint64(false)
	result.DelegationFeeRate = p.UnpackUint64(false)
	result.RewardAmount = p.UnpackUint64(false)
	result.BalanceBeforeUnstake = p.UnpackUint64(false)
	result.BalanceAfterUnstake = p.UnpackUint64(true)
	result.DistributedTo = p.UnpackString(true)
	return &result, p.Err()
}
