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
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*DelegateUserStake)(nil)

type DelegateUserStake struct {
	NodeID         []byte        `json:"nodeID"`         // Node ID of the validator to stake to
	StakeStartTime uint64        `json:"stakeStartTime"` // Start date of the stake
	StakeEndTime   uint64        `json:"stakeEndTime"`   // End date of the stake
	StakedAmount   uint64        `json:"stakedAmount"`   // Amount of NAI staked
	RewardAddress  codec.Address `json:"rewardAddress"`  // Address to receive rewards
}

func (*DelegateUserStake) GetTypeID() uint8 {
	return nconsts.DelegateUserStakeID
}

func (s *DelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) []string {
	if nodeID, err := ids.ToNodeID(s.NodeID); err == nil {
		return []string{
			string(storage.BalanceKey(actor, ids.Empty)),
			string(storage.DelegateUserStakeKey(actor, nodeID)),
		}
	}
	return []string{string(storage.BalanceKey(actor, ids.Empty))}
}

func (*DelegateUserStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.DelegateUserStakeChunks}
}

func (*DelegateUserStake) OutputsWarpMessage() bool {
	return false
}

func (s *DelegateUserStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	txID ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	nodeID, err := ids.ToNodeID(s.NodeID)
	if err != nil {
		return false, DelegateUserStakeComputeUnits, OutputInvalidNodeID, nil, nil
	}
	// Check if the user has already delegated before
	exists, _, _, _, _, _, _ := storage.GetDelegateUserStake(ctx, mu, actor, nodeID)
	if exists {
		return false, DelegateUserStakeComputeUnits, OutputUserAlreadyStaked, nil, nil
	}

	stakingConfig := emission.GetStakingConfig()

	// Check if the staked amount is a valid amount
	if s.StakedAmount < stakingConfig.MinDelegatorStake {
		return false, DelegateUserStakeComputeUnits, OutputDelegateStakedAmountInvalid, nil, nil
	}

	// Get current time
	currentTime := time.Now().UTC()
	// Convert Unix timestamps to Go's time.Time for easier manipulation
	startTime := time.Unix(int64(s.StakeStartTime), 0).UTC()
	if startTime.Before(currentTime) {
		return false, DelegateUserStakeComputeUnits, OutputInvalidStakeStartTime, nil, nil
	}
	endTime := time.Unix(int64(s.StakeEndTime), 0).UTC()
	// Check that stakeEndTime is greater than stakeStartTime
	if endTime.Before(startTime) {
		return false, DelegateUserStakeComputeUnits, OutputInvalidStakeEndTime, nil, nil
	}
	// Check that the total staking period is at least the minimum staking period
	stakeDuration := endTime.Sub(startTime)
	if stakeDuration < stakingConfig.MinDelegatorStakeDuration {
		return false, DelegateUserStakeComputeUnits, OutputInvalidStakeDuration, nil, nil
	}

	if err := storage.SubBalance(ctx, mu, actor, ids.Empty, s.StakedAmount); err != nil {
		return false, DelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.SetDelegateUserStake(ctx, mu, actor, nodeID, s.StakeStartTime, s.StakeEndTime, s.StakedAmount, s.RewardAddress); err != nil {
		return false, DelegateUserStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, DelegateUserStakeComputeUnits, nil, nil, nil
}

func (*DelegateUserStake) MaxComputeUnits(chain.Rules) uint64 {
	return DelegateUserStakeComputeUnits
}

func (*DelegateUserStake) Size() int {
	return hconsts.NodeIDLen + 3*hconsts.Uint64Len + codec.AddressLen
}

func (s *DelegateUserStake) Marshal(p *codec.Packer) {
	p.PackBytes(s.NodeID)
	p.PackUint64(s.StakeStartTime)
	p.PackUint64(s.StakeEndTime)
	p.PackUint64(s.StakedAmount)
	p.PackAddress(s.RewardAddress)
}

func UnmarshalDelegateUserStake(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var stake DelegateUserStake
	p.UnpackBytes(hconsts.NodeIDLen, true, &stake.NodeID)
	stake.StakeStartTime = p.UnpackUint64(true)
	stake.StakeEndTime = p.UnpackUint64(true)
	stake.StakedAmount = p.UnpackUint64(true)
	p.UnpackAddress(&stake.RewardAddress)
	return &stake, p.Err()
}

func (*DelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
