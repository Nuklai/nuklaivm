// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	RegisterValidatorStakeComputeUnits = 5
	StakeInfoSize                      = ids.NodeIDLen + 4*consts.Uint64Len + codec.AddressLen
)

var (
	ErrOutputDifferentSignerThanActor              = errors.New("output has a different signer than the actor")
	ErrNotValidatorOwner                           = errors.New("actor is not the owner of the validator")
	ErrInvalidNodeID                               = errors.New("invalid nodeID")
	ErrValidatorAlreadyRegistered                  = errors.New("validator is already registered")
	ErrValidatorStakedAmountInvalid                = errors.New("staked amount is invalid")
	ErrInvalidStakeStartBlock                      = errors.New("stakeStartBlock is invalid")
	ErrInvalidStakeEndBlock                        = errors.New("stakeEndBlock is invalid")
	ErrInvalidStakeDuration                        = errors.New("stake duration is invalid")
	ErrInvalidDelegationFeeRate                    = errors.New("delegation fee rate is invalid")
	_                                 chain.Action = (*RegisterValidatorStake)(nil)
)

type RegisterValidatorStake struct {
	NodeID        ids.NodeID `serialize:"true" json:"node_id"`        // Node ID of the validator
	StakeInfo     []byte     `serialize:"true" json:"stake_info"`     // StakeInfo of the validator
	AuthSignature []byte     `serialize:"true" json:"auth_signature"` // Auth BLS signature of the validator
}

func (*RegisterValidatorStake) GetTypeID() uint8 {
	return nconsts.RegisterValidatorStakeID
}

func (r *RegisterValidatorStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.BalanceKey(actor, ids.Empty)):        state.Read | state.Write,
		string(storage.RegisterValidatorStakeKey(r.NodeID)): state.Allocate | state.Write,
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
) (codec.Typed, error) {
	// Unmarshal the stake info
	packer := codec.NewReader(r.StakeInfo, len(r.StakeInfo))
	sInfo, err := UnmarshalRegisterValidatorStakeResult(packer)
	if err != nil {
		return nil, err
	}
	stakeInfo, ok := sInfo.(*RegisterValidatorStakeResult)
	if !ok {
		return nil, errors.New("failed to unmarshal stake info")
	}
	// Check if nodeID passed is the same as the one in the stake info
	if r.NodeID != stakeInfo.NodeID {
		return nil, ErrInvalidNodeID
	}

	// Check if it's a valid signature
	signer, err := VerifyAuthSignature(r.StakeInfo, r.AuthSignature)
	if err != nil {
		return nil, err
	}
	// Check whether the actor is the same as the one who signed the message
	if actor != signer {
		return nil, ErrOutputDifferentSignerThanActor
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
		if actor == signer {
			isValidatorOwner = true
			nodePublicKey = publicKey
			break
		}
	}
	if !isValidatorOwner {
		return nil, ErrNotValidatorOwner
	}

	// Check if the validator was already registered
	exists, _, _, _, _, _, _, _ := storage.GetRegisterValidatorStake(ctx, mu, stakeInfo.NodeID)
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
	err = emissionInstance.RegisterValidatorStake(stakeInfo.NodeID, nodePublicKey, stakeInfo.StakeStartBlock, stakeInfo.StakeEndBlock, stakeInfo.StakedAmount, stakeInfo.DelegationFeeRate)
	if err != nil {
		return nil, err
	}

	if _, err := storage.SubBalance(ctx, mu, actor, ids.Empty, stakeInfo.StakedAmount); err != nil {
		return nil, err
	}
	if err := storage.SetRegisterValidatorStake(ctx, mu, stakeInfo.NodeID, stakeInfo.StakeStartBlock, stakeInfo.StakeEndBlock, stakeInfo.StakedAmount, stakeInfo.DelegationFeeRate, stakeInfo.RewardAddress, actor); err != nil {
		return nil, err
	}
	return &RegisterValidatorStakeResult{
		NodeID:            stakeInfo.NodeID,
		StakeStartBlock:   stakeInfo.StakeStartBlock,
		StakeEndBlock:     stakeInfo.StakeEndBlock,
		StakedAmount:      stakeInfo.StakedAmount,
		DelegationFeeRate: stakeInfo.DelegationFeeRate,
		RewardAddress:     stakeInfo.RewardAddress,
	}, nil
}

func (*RegisterValidatorStake) ComputeUnits(chain.Rules) uint64 {
	return RegisterValidatorStakeComputeUnits
}

func (*RegisterValidatorStake) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*RegisterValidatorStake)(nil)

func (*RegisterValidatorStake) Size() int {
	return ids.NodeIDLen + StakeInfoSize + auth.BLSSize
}

func (r *RegisterValidatorStake) Marshal(p *codec.Packer) {
	p.PackFixedBytes(r.NodeID.Bytes())
	p.PackBytes(r.StakeInfo)
	p.PackBytes(r.AuthSignature)
}

func UnmarshalRegisterValidatorStake(p *codec.Packer) (chain.Action, error) {
	var stake RegisterValidatorStake
	nodeIDBytes := make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &nodeIDBytes)
	nodeID, err := ids.ToNodeID(nodeIDBytes)
	if err != nil {
		return nil, err
	}
	stake.NodeID = nodeID
	p.UnpackBytes(StakeInfoSize, true, &stake.StakeInfo)
	p.UnpackBytes(auth.BLSSize, true, &stake.AuthSignature)
	return &stake, p.Err()
}

func VerifyAuthSignature(content, authSignature []byte) (codec.Address, error) {
	p := codec.NewReader(authSignature, len(authSignature))
	sig, err := auth.UnmarshalBLS(p)
	if err != nil {
		return codec.EmptyAddress, err
	}
	return sig.Actor(), sig.Verify(context.TODO(), content)
}

var (
	_ codec.Typed     = (*RegisterValidatorStakeResult)(nil)
	_ chain.Marshaler = (*RegisterValidatorStakeResult)(nil)
)

type RegisterValidatorStakeResult struct {
	NodeID            ids.NodeID    `serialize:"true" json:"node_id"`
	StakeStartBlock   uint64        `serialize:"true" json:"stake_start_block"`
	StakeEndBlock     uint64        `serialize:"true" json:"stake_end_block"`
	StakedAmount      uint64        `serialize:"true" json:"staked_amount"`
	DelegationFeeRate uint64        `serialize:"true" json:"delegation_fee_rate"`
	RewardAddress     codec.Address `serialize:"true" json:"reward_address"`
}

func (*RegisterValidatorStakeResult) GetTypeID() uint8 {
	return nconsts.RegisterValidatorStakeID
}

func (*RegisterValidatorStakeResult) Size() int {
	return StakeInfoSize
}

func (r *RegisterValidatorStakeResult) Marshal(p *codec.Packer) {
	p.PackFixedBytes(r.NodeID.Bytes())
	p.PackUint64(r.StakeStartBlock)
	p.PackUint64(r.StakeEndBlock)
	p.PackUint64(r.StakedAmount)
	p.PackUint64(r.DelegationFeeRate)
	p.PackAddress(r.RewardAddress)
}

func UnmarshalRegisterValidatorStakeResult(p *codec.Packer) (codec.Typed, error) {
	var result RegisterValidatorStakeResult
	nodeIDBytes := make([]byte, ids.NodeIDLen)
	p.UnpackFixedBytes(ids.NodeIDLen, &nodeIDBytes)
	nodeID, err := ids.ToNodeID(nodeIDBytes)
	if err != nil {
		return nil, err
	}
	result.NodeID = nodeID
	result.StakeStartBlock = p.UnpackUint64(true)
	result.StakeEndBlock = p.UnpackUint64(true)
	result.StakedAmount = p.UnpackUint64(true)
	result.DelegationFeeRate = p.UnpackUint64(false)
	p.UnpackAddress(&result.RewardAddress)
	return &result, p.Err()
}
