// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/state"

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

func (r *RegisterValidatorStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	// TODO: How to better handle a case where the NodeID is invalid?
	stakeInfo, _ := UnmarshalValidatorStakeInfo(r.StakeInfo)
	nodeID, _ := ids.ToNodeID(stakeInfo.NodeID)
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):      state.Read | state.Write,
		string(storage.RegisterValidatorStakeKey(nodeID)): state.Allocate | state.Write,
	}
}

func (*RegisterValidatorStake) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.RegisterValidatorStakeChunks}
}

func (r *RegisterValidatorStake) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if it's a valid signature
	signer, err := VerifyAuthSignature(r.StakeInfo, r.AuthSignature)
	if err != nil {
		return nil, err
	}
	// Check whether the actor is the same as the one who signed the message
	actorAddress := codec.MustAddressBech32(nconsts.HRP, actor)
	if actorAddress != codec.MustAddressBech32(nconsts.HRP, signer) {
		return nil, ErrDifferentSignerThanActor
	}

	// Check if the tx actor has signing permission for this NodeID
	isValidatorOwner := false

	// Get the emission instance
	emissionInstance := emission.GetEmission()
	currentValidators := emissionInstance.GetAllValidators(ctx)

	var nodePublicKey *bls.PublicKey
	for _, validator := range currentValidators {
		publicKey, err := bls.PublicKeyFromBytes(validator.PublicKey)
		if err != nil {
			return nil, err
		}
		signer := auth.NewBLSAddress(publicKey)
		if actorAddress == codec.MustAddressBech32(nconsts.HRP, signer) {
			isValidatorOwner = true
			nodePublicKey = publicKey
			break
		}
	}
	if !isValidatorOwner {
		return nil, ErrUnauthorized
	}

	// Unmarshal the stake info
	stakeInfo, err := UnmarshalValidatorStakeInfo(r.StakeInfo)
	if err != nil {
		return nil, err
	}
	// Check if it's a valid nodeID
	nodeID, err := ids.ToNodeID(stakeInfo.NodeID)
	if err != nil {
		return nil, ErrInvalidNodeID
	}

	// Check if the validator was already registered
	exists, _, _, _, _, _, _, _ := storage.GetRegisterValidatorStake(ctx, mu, nodeID)
	if exists {
		return nil, ErrValidatorAlreadyRegistered
	}

	stakingConfig := emission.GetStakingConfig()

	// Check if the staked amount is a valid amount
	if stakeInfo.StakedAmount < stakingConfig.MinValidatorStake || stakeInfo.StakedAmount > stakingConfig.MaxValidatorStake {
		return nil, ErrValidatorStakedAmountInvalid
	}

	// Get last accepted block height
	lastBlockHeight := emissionInstance.GetLastAcceptedBlockHeight()

	// Check that stakeStartBlock is after lastBlockHeight
	if stakeInfo.StakeStartBlock < lastBlockHeight {
		return nil, ErrInvalidStakeStartBlock
	}
	// Check that stakeEndBlock is after stakeStartBlock
	if stakeInfo.StakeEndBlock < stakeInfo.StakeStartBlock {
		return nil, ErrInvalidStakeEndBlock
	}

	// Check that the total staking period is at least the minimum staking period
	stakeDuration := stakeInfo.StakeEndBlock - stakeInfo.StakeStartBlock
	if stakeDuration < stakingConfig.MinValidatorStakeDuration || stakeDuration > stakingConfig.MaxValidatorStakeDuration {
		return nil, ErrInvalidStakeDuration
	}

	// Check if the delegation fee rate is valid
	if stakeInfo.DelegationFeeRate < stakingConfig.MinDelegationFee || stakeInfo.DelegationFeeRate > 100 {
		return nil, ErrInvalidDelegationFeeRate
	}

	// Register in Emission Balancer
	err = emissionInstance.RegisterValidatorStake(nodeID, nodePublicKey, stakeInfo.StakeStartBlock, stakeInfo.StakeEndBlock, stakeInfo.StakedAmount, stakeInfo.DelegationFeeRate)
	if err != nil {
		return nil, err
	}

	if err := storage.SubBalance(ctx, mu, actor, ids.Empty, stakeInfo.StakedAmount); err != nil {
		return nil, err
	}
	if err := storage.SetRegisterValidatorStake(ctx, mu, nodeID, stakeInfo.StakeStartBlock, stakeInfo.StakeEndBlock, stakeInfo.StakedAmount, stakeInfo.DelegationFeeRate, stakeInfo.RewardAddress, actor); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*RegisterValidatorStake) ComputeUnits(chain.Rules) uint64 {
	return RegisterValidatorStakeComputeUnits
}

func (*RegisterValidatorStake) Size() int {
	return ids.NodeIDLen + 4*consts.Uint64Len + codec.AddressLen + bls.PublicKeyLen + bls.SignatureLen
}

func (r *RegisterValidatorStake) Marshal(p *codec.Packer) {
	p.PackBytes(r.StakeInfo)
	p.PackBytes(r.AuthSignature)
}

func UnmarshalRegisterValidatorStake(p *codec.Packer) (chain.Action, error) {
	var stake RegisterValidatorStake
	p.UnpackBytes(ids.NodeIDLen+4*consts.Uint64Len+codec.AddressLen, true, &stake.StakeInfo)
	p.UnpackBytes(bls.PublicKeyLen+bls.SignatureLen, true, &stake.AuthSignature)
	return &stake, p.Err()
}

func (*RegisterValidatorStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func VerifyAuthSignature(content, authSignature []byte) (codec.Address, error) {
	p := codec.NewReader(authSignature, len(authSignature))
	sig, err := auth.UnmarshalBLS(p)
	if err != nil {
		return codec.EmptyAddress, err
	}
	return sig.Actor(), sig.Verify(context.TODO(), content)
}

type ValidatorStakeInfo struct {
	NodeID            []byte        `json:"nodeID"`            // NodeID of the validator
	StakeStartBlock   uint64        `json:"stakeStartBlock"`   // Start block of the stake
	StakeEndBlock     uint64        `json:"stakeEndBlock"`     // End block of the stake
	StakedAmount      uint64        `json:"stakedAmount"`      // Amount of NAI staked
	DelegationFeeRate uint64        `json:"delegationFeeRate"` // Delegation fee rate
	RewardAddress     codec.Address `json:"rewardAddress"`     // Address to receive rewards
}

func UnmarshalValidatorStakeInfo(stakeInfo []byte) (*ValidatorStakeInfo, error) {
	p := codec.NewReader(stakeInfo, ids.NodeIDLen+4*consts.Uint64Len+codec.AddressLen)
	var result ValidatorStakeInfo
	result.NodeID = make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &result.NodeID)
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.StakedAmount = p.UnpackUint64(true)
	result.DelegationFeeRate = p.UnpackUint64(true)
	p.UnpackAddress(&result.RewardAddress)
	return &result, p.Err()
}

func (s *ValidatorStakeInfo) Marshal() ([]byte, error) {
	p := codec.NewWriter(ids.NodeIDLen+4*consts.Uint64Len+codec.AddressLen, ids.NodeIDLen+4*consts.Uint64Len+codec.AddressLen)
	p.PackFixedBytes(s.NodeID)
	p.PackUint64(s.StakeStartBlock)
	p.PackUint64(s.StakeEndBlock)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.DelegationFeeRate)
	p.PackAddress(s.RewardAddress)
	return p.Bytes(), p.Err()
}
