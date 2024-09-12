// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
)

// [registerValidatorStakePrefix] + [nodeID]
func RegisterValidatorStakeKey(nodeID ids.NodeID) (k []byte) {
	k = make([]byte, 1+ids.NodeIDLen+consts.Uint16Len) // Length of prefix + nodeID + RegisterValidatorStakeChunks
	k[0] = registerValidatorStakePrefix                // registerValidatorStakePrefix is a constant representing the registerValidatorStake category
	copy(k[1:], nodeID[:])
	binary.BigEndian.PutUint16(k[1+ids.NodeIDLen:], RegisterValidatorStakeChunks) // Adding RegisterValidatorStakeChunks
	return
}

func SetRegisterValidatorStake(
	ctx context.Context,
	mu state.Mutable,
	nodeID ids.NodeID,
	stakeStartBlock uint64,
	stakeEndBlock uint64,
	stakedAmount uint64,
	delegationFeeRate uint64,
	rewardAddress codec.Address,
	ownerAddress codec.Address,
) error {
	key := RegisterValidatorStakeKey(nodeID)
	v := make([]byte, (4*consts.Uint64Len)+(2*codec.AddressLen)) // Calculate the length of the encoded data

	offset := 0
	binary.BigEndian.PutUint64(v[offset:], stakeStartBlock)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakeEndBlock)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakedAmount)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], delegationFeeRate)
	offset += consts.Uint64Len

	copy(v[offset:], rewardAddress[:])
	offset += codec.AddressLen

	copy(v[offset:], ownerAddress[:])

	return mu.Insert(ctx, key, v)
}

func GetRegisterValidatorStake(
	ctx context.Context,
	im state.Immutable,
	nodeID ids.NodeID,
) (bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	uint64, // DelegationFeeRate
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	key := RegisterValidatorStakeKey(nodeID)
	v, err := im.GetValue(ctx, key)
	return innerGetRegisterValidatorStake(v, err)
}

// Used to serve RPC queries
func GetRegisterValidatorStakeFromState(
	ctx context.Context,
	f ReadState,
	nodeID ids.NodeID,
) (bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	uint64, // DelegationFeeRate
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	values, errs := f(ctx, [][]byte{RegisterValidatorStakeKey(nodeID)})
	return innerGetRegisterValidatorStake(values[0], errs[0])
}

func innerGetRegisterValidatorStake(v []byte, err error) (
	bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	uint64, // DelegationFeeRate
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, 0, 0, 0, codec.Address{}, codec.Address{}, nil
	}
	if err != nil {
		return false, 0, 0, 0, 0, codec.Address{}, codec.Address{}, nil
	}

	offset := 0
	stakeStartBlock := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len
	stakeEndBlock := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len
	stakedAmount := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len
	delegationFeeRate := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len

	var rewardAddress codec.Address
	copy(rewardAddress[:], v[offset:offset+codec.AddressLen])
	offset += codec.AddressLen

	var ownerAddress codec.Address
	copy(ownerAddress[:], v[offset:offset+codec.AddressLen])

	return true, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, nil
}

func DeleteRegisterValidatorStake(
	ctx context.Context,
	mu state.Mutable,
	nodeID ids.NodeID,
) error {
	return mu.Remove(ctx, RegisterValidatorStakeKey(nodeID))
}

// [delegateUserStakePrefix] + [txID]
func DelegateUserStakeKey(owner codec.Address, nodeID ids.NodeID) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+ids.NodeIDLen+consts.Uint16Len) // Length of prefix + owner + nodeID + DelegateUserStakeChunks
	k[0] = delegateUserStakePrefix                                      // delegateUserStakePrefix is a constant representing the staking category
	copy(k[1:], owner[:])
	copy(k[1+codec.AddressLen:], nodeID[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen+ids.NodeIDLen:], DelegateUserStakeChunks) // Adding DelegateUserStakeChunks
	return
}

func SetDelegateUserStake(
	ctx context.Context,
	mu state.Mutable,
	owner codec.Address,
	nodeID ids.NodeID,
	stakeStartBlock uint64,
	stakeEndBlock uint64,
	stakedAmount uint64,
	rewardAddress codec.Address,
) error {
	key := DelegateUserStakeKey(owner, nodeID)
	v := make([]byte, 3*consts.Uint64Len+2*codec.AddressLen) // Calculate the length of the encoded data

	offset := 0

	binary.BigEndian.PutUint64(v[offset:], stakeStartBlock)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakeEndBlock)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakedAmount)
	offset += consts.Uint64Len

	copy(v[offset:], rewardAddress[:])
	offset += codec.AddressLen
	copy(v[offset:], owner[:])

	return mu.Insert(ctx, key, v)
}

func GetDelegateUserStake(
	ctx context.Context,
	im state.Immutable,
	owner codec.Address,
	nodeID ids.NodeID,
) (bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	key := DelegateUserStakeKey(owner, nodeID)
	v, err := im.GetValue(ctx, key)
	return innerGetDelegateUserStake(v, err)
}

// Used to serve RPC queries
func GetDelegateUserStakeFromState(
	ctx context.Context,
	f ReadState,
	owner codec.Address,
	nodeID ids.NodeID,
) (bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	values, errs := f(ctx, [][]byte{DelegateUserStakeKey(owner, nodeID)})
	return innerGetDelegateUserStake(values[0], errs[0])
}

func innerGetDelegateUserStake(v []byte, err error) (
	bool, // exists
	uint64, // StakeStartBlock
	uint64, // StakeEndBlock
	uint64, // StakedAmount
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, 0, 0, codec.Address{}, codec.Address{}, nil
	}
	if err != nil {
		return false, 0, 0, 0, codec.Address{}, codec.Address{}, nil
	}

	offset := 0

	stakeStartBlock := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len
	stakeEndBlock := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len
	stakedAmount := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len

	var rewardAddress codec.Address
	copy(rewardAddress[:], v[offset:offset+codec.AddressLen])
	offset += codec.AddressLen
	var ownerAddress codec.Address
	copy(ownerAddress[:], v[offset:offset+codec.AddressLen])

	return true, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, nil
}

func DeleteDelegateUserStake(
	ctx context.Context,
	mu state.Mutable,
	owner codec.Address,
	nodeID ids.NodeID,
) error {
	return mu.Remove(ctx, DelegateUserStakeKey(owner, nodeID))
}
