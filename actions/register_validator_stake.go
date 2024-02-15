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
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*RegisterValidatorStake)(nil)

type RegisterValidatorStake struct {
	NodeID            []byte        `json:"nodeID"`            // NodeID of the validator
	StakeStartTime    uint64        `json:"stakeStartTime"`    // Start date of the stake
	StakeEndTime      uint64        `json:"stakeEndTime"`      // End date of the stake
	StakedAmount      uint64        `json:"stakedAmount"`      // Amount of NAI staked
	DelegationFeeRate uint64        `json:"delegationFeeRate"` // Delegation fee rate
	RewardAddress     codec.Address `json:"rewardAddress"`     // Address to receive rewards
}

func (*RegisterValidatorStake) GetTypeID() uint8 {
	return nconsts.RegisterValidatorStakeID
}

func (r *RegisterValidatorStake) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	if nodeID, err := ids.ToNodeID(r.NodeID); err == nil {
		return []string{
			string(storage.BalanceKey(actor, ids.Empty)),
			string(storage.RegisterValidatorStakeKey(nodeID)),
		}
	}
	return []string{string(storage.BalanceKey(actor, ids.Empty))}
}

func (*RegisterValidatorStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks}
}

func (*RegisterValidatorStake) OutputsWarpMessage() bool {
	return false
}

func (r *RegisterValidatorStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	// Check if it's a valid nodeID
	nodeID, err := ids.ToNodeID(r.NodeID)
	if err != nil {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidNodeID, nil, nil
	}

	// Check if the validator was already registered
	exists, _, _, _, _, _, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if exists {
		return false, RegisterValidatorStakeComputeUnits, OutputValidatorAlreadyRegistered, nil, nil
	}

	// Check if the staked amount is greater than or equal to 1.5 million NAI
	minAmountToStake, _ := utils.ParseBalance("1500000", nconsts.Decimals)
	if r.StakedAmount < minAmountToStake {
		return false, RegisterValidatorStakeComputeUnits, OutputStakedAmountZero, nil, nil
	}

	// Get current time
	currentTime := time.Now().UTC()
	// Convert Unix timestamps to Go's time.Time for easier manipulation
	startTime := time.Unix(int64(r.StakeStartTime), 0).UTC()
	if startTime.Before(currentTime) {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeStartTime, nil, nil
	}
	endTime := time.Unix(int64(r.StakeEndTime), 0).UTC()
	// Check that stakeEndTime is greater than stakeStartTime
	if endTime.Before(startTime) {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeEndTime, nil, nil
	}
	// TODO: Disable this when we go to production
	// Check that the total staking period is at least 60 seconds
	if r.StakeEndTime-r.StakeStartTime < 60 {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeEndTime, nil, nil
	}
	// TODO: Enable this when we go to production
	// Check that stakeEndTime is at least 6 months after stakeStartTime
	// Adding 6 months to startTime
	/*
		sixMonthsAfterStart := startTime.AddDate(0, 6, 0)
		if endTime.Before(sixMonthsAfterStart) {
			return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeEndTime, nil, nil
		}
	*/
	if r.DelegationFeeRate < 2 || r.DelegationFeeRate > 100 {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidDelegationFeeRate, nil, nil
	}
	// TODO: Check if the NodeID belongs to the actor

	if err := storage.SubBalance(ctx, mu, actor, ids.Empty, r.StakedAmount); err != nil {
		return false, RegisterValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.SetRegisterValidatorStake(ctx, mu, nodeID, r.StakeStartTime, r.StakeEndTime, r.StakedAmount, r.DelegationFeeRate, r.RewardAddress, actor); err != nil {
		return false, RegisterValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, RegisterValidatorStakeComputeUnits, nil, nil, nil
}

func (*RegisterValidatorStake) MaxComputeUnits(chain.Rules) uint64 {
	return RegisterValidatorStakeComputeUnits
}

func (*RegisterValidatorStake) Size() int {
	return hconsts.NodeIDLen + 4*hconsts.Uint64Len + codec.AddressLen
}

func (r *RegisterValidatorStake) Marshal(p *codec.Packer) {
	p.PackBytes(r.NodeID)
	p.PackUint64(r.StakeStartTime)
	p.PackUint64(r.StakeEndTime)
	p.PackUint64(r.StakedAmount)
	p.PackUint64(r.DelegationFeeRate)
	p.PackAddress(r.RewardAddress)
}

func UnmarshalRegisterValidatorStake(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var stake RegisterValidatorStake
	p.UnpackBytes(hconsts.NodeIDLen, false, &stake.NodeID)
	stake.StakeStartTime = p.UnpackUint64(true)
	stake.StakeEndTime = p.UnpackUint64(true)
	stake.StakedAmount = p.UnpackUint64(true)
	stake.DelegationFeeRate = p.UnpackUint64(true)
	p.UnpackAddress(&stake.RewardAddress)
	return &stake, p.Err()
}

func (*RegisterValidatorStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
