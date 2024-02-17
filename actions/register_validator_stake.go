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
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*RegisterValidatorStake)(nil)

type RegisterValidatorStake struct {
	StakeInfo     []byte `json:"stakeInfo"`     // StakeInfo of the validator
	AuthSignature []byte `json:"authSignature"` // Auth BLS signature of the validator
}

func (*RegisterValidatorStake) GetTypeID() uint8 {
	return nconsts.RegisterValidatorStakeID
}

func (r *RegisterValidatorStake) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	if signer, err := VerifyAuthSignature(r.StakeInfo, r.AuthSignature); err == nil && codec.MustAddressBech32(nconsts.HRP, actor) == codec.MustAddressBech32(nconsts.HRP, signer) {
		if stakeInfo, err := UnmarshalValidatorStakeInfo(r.StakeInfo); err == nil {
			if nodeID, err := ids.ToNodeID(stakeInfo.NodeID); err == nil {
				return []string{
					string(storage.BalanceKey(actor, ids.Empty)),
					string(storage.RegisterValidatorStakeKey(nodeID)),
				}
			}
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
	// Check if it's a valid signature
	signer, err := VerifyAuthSignature(r.StakeInfo, r.AuthSignature)
	if err != nil {
		return false, RegisterValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	// Check whether the actor is the same as the one who signed the message
	if codec.MustAddressBech32(nconsts.HRP, actor) != codec.MustAddressBech32(nconsts.HRP, signer) {
		return false, RegisterValidatorStakeComputeUnits, OutputDifferentSignerThanActor, nil, nil
	}

	// Unmarshal the stake info
	stakeInfo, err := UnmarshalValidatorStakeInfo(r.StakeInfo)
	if err != nil {
		return false, RegisterValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	// Check if it's a valid nodeID
	nodeID, err := ids.ToNodeID(stakeInfo.NodeID)
	if err != nil {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidNodeID, nil, nil
	}

	// Check if the validator was already registered
	exists, _, _, _, _, _, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if exists {
		return false, RegisterValidatorStakeComputeUnits, OutputValidatorAlreadyRegistered, nil, nil
	}

	stakingConfig := emission.GetStakingConfig()

	// Check if the staked amount is a valid amount
	if stakeInfo.StakedAmount < stakingConfig.MinValidatorStake && stakeInfo.StakedAmount > stakingConfig.MaxValidatorStake {
		return false, RegisterValidatorStakeComputeUnits, OutputValidatorStakedAmountInvalid, nil, nil
	}

	// Get current time
	currentTime := time.Now().UTC()
	// Convert Unix timestamps to Go's time.Time for easier manipulation
	startTime := time.Unix(int64(stakeInfo.StakeStartTime), 0).UTC()
	if startTime.Before(currentTime) {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeStartTime, nil, nil
	}
	endTime := time.Unix(int64(stakeInfo.StakeEndTime), 0).UTC()
	// Check that stakeEndTime is greater than stakeStartTime
	if endTime.Before(startTime) {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeEndTime, nil, nil
	}
	// Check that the total staking period is at least the minimum staking period
	stakeDuration := endTime.Sub(startTime)
	if stakeDuration < stakingConfig.MinValidatorStakeDuration || stakeDuration > stakingConfig.MaxValidatorStakeDuration {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidStakeDuration, nil, nil
	}

	if stakeInfo.DelegationFeeRate < stakingConfig.MinDelegationFee || stakeInfo.DelegationFeeRate > 100 {
		return false, RegisterValidatorStakeComputeUnits, OutputInvalidDelegationFeeRate, nil, nil
	}

	if err := storage.SubBalance(ctx, mu, actor, ids.Empty, stakeInfo.StakedAmount); err != nil {
		return false, RegisterValidatorStakeComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.SetRegisterValidatorStake(ctx, mu, nodeID, stakeInfo.StakeStartTime, stakeInfo.StakeEndTime, stakeInfo.StakedAmount, stakeInfo.DelegationFeeRate, stakeInfo.RewardAddress, actor); err != nil {
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
	p.PackBytes(r.StakeInfo)
	p.PackBytes(r.AuthSignature)
}

func UnmarshalRegisterValidatorStake(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var stake RegisterValidatorStake
	p.UnpackBytes(hconsts.NodeIDLen+4*hconsts.Uint64Len+codec.AddressLen, true, &stake.StakeInfo)
	p.UnpackBytes(bls.PublicKeyLen+bls.SignatureLen, true, &stake.AuthSignature)
	return &stake, p.Err()
}

func (*RegisterValidatorStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func VerifyAuthSignature(stakeInfo, authSignature []byte) (codec.Address, error) {
	p := codec.NewReader(authSignature, len(authSignature))
	sig, err := auth.UnmarshalBLS(p, nil)
	if err != nil {
		return codec.EmptyAddress, err
	}
	return sig.Actor(), sig.Verify(context.TODO(), stakeInfo)
}

type ValidatorStakeInfo struct {
	NodeID            []byte        `json:"nodeID"`            // NodeID of the validator
	StakeStartTime    uint64        `json:"stakeStartTime"`    // Start date of the stake
	StakeEndTime      uint64        `json:"stakeEndTime"`      // End date of the stake
	StakedAmount      uint64        `json:"stakedAmount"`      // Amount of NAI staked
	DelegationFeeRate uint64        `json:"delegationFeeRate"` // Delegation fee rate
	RewardAddress     codec.Address `json:"rewardAddress"`     // Address to receive rewards
}

func UnmarshalValidatorStakeInfo(stakeInfo []byte) (*ValidatorStakeInfo, error) {
	p := codec.NewReader(stakeInfo, hconsts.NodeIDLen+4*hconsts.Uint64Len+codec.AddressLen)
	var result ValidatorStakeInfo
	result.NodeID = make([]byte, hconsts.NodeIDLen)
	p.UnpackFixedBytes(hconsts.NodeIDLen, &result.NodeID)
	result.StakeStartTime = p.UnpackUint64(true)
	result.StakeEndTime = p.UnpackUint64(true)
	result.StakedAmount = p.UnpackUint64(true)
	result.DelegationFeeRate = p.UnpackUint64(true)
	p.UnpackAddress(&result.RewardAddress)
	return &result, p.Err()
}

func (s *ValidatorStakeInfo) Marshal() ([]byte, error) {
	p := codec.NewWriter(hconsts.NodeIDLen+4*hconsts.Uint64Len+codec.AddressLen, hconsts.NodeIDLen+4*hconsts.Uint64Len+codec.AddressLen)
	p.PackFixedBytes(s.NodeID)
	p.PackUint64(s.StakeStartTime)
	p.PackUint64(s.StakeEndTime)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.DelegationFeeRate)
	p.PackAddress(s.RewardAddress)
	return p.Bytes(), p.Err()
}
