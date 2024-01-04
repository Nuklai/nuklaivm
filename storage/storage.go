// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	mconsts "github.com/nuklai/nuklaivm/consts"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

// Metadata
// 0x0/ (tx)
//   -> [txID] => timestamp
//
// State
// / (height) => store in root
//   -> [heightPrefix] => height
// 0x0/ (balance)
//   -> [owner] => balance
// 0x1/ (hypersdk-height)
// 0x2/ (hypersdk-timestamp)
// 0x3/ (hypersdk-fee)
// 0x4/ (hypersdk-incoming warp)
// 0x5/ (hypersdk-outgoing warp)

const (
	// metaDB
	txPrefix = 0x0

	// stateDB
	incomingWarpPrefix = 0x0
	outgoingWarpPrefix = 0x1
	balancePrefix      = 0x2
	heightPrefix       = 0x3
	timestampPrefix    = 0x4
	feePrefix          = 0x5
	stakePrefix        = 0x6
)

const (
	BalanceChunks uint16 = 1
	StakeChunks   uint16 = 2
)

var (
	failureByte  = byte(0x0)
	successByte  = byte(0x1)
	heightKey    = []byte{heightPrefix}
	timestampKey = []byte{timestampPrefix}
	feeKey       = []byte{feePrefix}
)

// [txPrefix] + [txID]
func TxKey(id ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = txPrefix
	copy(k[1:], id[:])
	return
}

func StoreTransaction(
	_ context.Context,
	db database.KeyValueWriter,
	id ids.ID,
	t int64,
	success bool,
	units chain.Dimensions,
	fee uint64,
) error {
	k := TxKey(id)
	v := make([]byte, consts.Uint64Len+1+chain.DimensionsLen+consts.Uint64Len)
	binary.BigEndian.PutUint64(v, uint64(t))
	if success {
		v[consts.Uint64Len] = successByte
	} else {
		v[consts.Uint64Len] = failureByte
	}
	copy(v[consts.Uint64Len+1:], units.Bytes())
	binary.BigEndian.PutUint64(v[consts.Uint64Len+1+chain.DimensionsLen:], fee)
	return db.Put(k, v)
}

func GetTransaction(
	_ context.Context,
	db database.KeyValueReader,
	id ids.ID,
) (bool, int64, bool, chain.Dimensions, uint64, error) {
	k := TxKey(id)
	v, err := db.Get(k)
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, false, chain.Dimensions{}, 0, nil
	}
	if err != nil {
		return false, 0, false, chain.Dimensions{}, 0, err
	}
	t := int64(binary.BigEndian.Uint64(v))
	success := true
	if v[consts.Uint64Len] == failureByte {
		success = false
	}
	d, err := chain.UnpackDimensions(v[consts.Uint64Len+1 : consts.Uint64Len+1+chain.DimensionsLen])
	if err != nil {
		return false, 0, false, chain.Dimensions{}, 0, err
	}
	fee := binary.BigEndian.Uint64(v[consts.Uint64Len+1+chain.DimensionsLen:])
	return true, t, success, d, fee, nil
}

// [balancePrefix] + [address]
func BalanceKey(addr codec.Address) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+consts.Uint16Len)
	k[0] = balancePrefix
	copy(k[1:], addr[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], BalanceChunks)
	return
}

// If locked is 0, then account does not exist
func GetBalance(
	ctx context.Context,
	im state.Immutable,
	addr codec.Address,
) (uint64, error) {
	_, bal, _, err := getBalance(ctx, im, addr)
	return bal, err
}

func getBalance(
	ctx context.Context,
	im state.Immutable,
	addr codec.Address,
) ([]byte, uint64, bool, error) {
	k := BalanceKey(addr)
	bal, exists, err := innerGetBalance(im.GetValue(ctx, k))
	return k, bal, exists, err
}

// Used to serve RPC queries
func GetBalanceFromState(
	ctx context.Context,
	f ReadState,
	addr codec.Address,
) (uint64, error) {
	k := BalanceKey(addr)
	values, errs := f(ctx, [][]byte{k})
	bal, _, err := innerGetBalance(values[0], errs[0])
	return bal, err
}

func innerGetBalance(
	v []byte,
	err error,
) (uint64, bool, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return binary.BigEndian.Uint64(v), true, nil
}

func SetBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	balance uint64,
) error {
	k := BalanceKey(addr)
	return setBalance(ctx, mu, k, balance)
}

func setBalance(
	ctx context.Context,
	mu state.Mutable,
	key []byte,
	balance uint64,
) error {
	return mu.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, balance))
}

func AddBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	amount uint64,
	create bool,
) error {
	key, bal, exists, err := getBalance(ctx, mu, addr)
	if err != nil {
		return err
	}
	// Don't add balance if account doesn't exist. This
	// can be useful when processing fee refunds.
	if !exists && !create {
		return nil
	}
	nbal, err := smath.Add64(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not add balance (bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			bal,
			codec.MustAddressBech32(mconsts.HRP, addr),
			amount,
		)
	}
	return setBalance(ctx, mu, key, nbal)
}

func SubBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	amount uint64,
) error {
	key, bal, _, err := getBalance(ctx, mu, addr)
	if err != nil {
		return err
	}
	nbal, err := smath.Sub(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not subtract balance (bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			bal,
			codec.MustAddressBech32(mconsts.HRP, addr),
			amount,
		)
	}
	if nbal == 0 {
		// If there is no balance left, we should delete the record instead of
		// setting it to 0.
		return mu.Remove(ctx, key)
	}
	return setBalance(ctx, mu, key, nbal)
}

// [stakePrefix] + [txID]
func StakeKey(txID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen+consts.Uint16Len) // Length of prefix + txID + stakeChunks
	k[0] = stakePrefix                                // stakePrefix is a constant representing the staking category
	copy(k[1:], txID[:])
	binary.BigEndian.PutUint16(k[1+consts.IDLen:], StakeChunks) // Adding StakeChunks
	return
}

func SetStake(
	ctx context.Context,
	mu state.Mutable,
	stake ids.ID,
	nodeID []byte,
	stakedAmount uint64,
	endLockUp uint64,
	owner codec.Address,
) error {
	key := StakeKey(stake)
	v := make([]byte, consts.NodeIDLen+(2*consts.Uint64Len)+codec.AddressLen) // Calculate the length of the encoded data

	id, err := ids.ToNodeID(nodeID)
	if err != nil {
		return err
	}

	offset := 0
	copy(v[offset:], id[:])
	offset += consts.NodeIDLen

	binary.BigEndian.PutUint64(v[offset:], stakedAmount)
	offset += consts.Uint64Len

	binary.BigEndian.PutUint64(v[offset:], endLockUp)
	offset += consts.Uint64Len

	copy(v[offset:], owner[:])

	return mu.Insert(ctx, key, v)
}

func GetStake(
	ctx context.Context,
	im state.Immutable,
	stake ids.ID,
) (bool, // exists
	ids.NodeID, // NodeID
	uint64, // StakedAmount
	uint64, // EndLockUp
	codec.Address, // Owner
	error) {
	key := StakeKey(stake)
	v, err := im.GetValue(ctx, key)
	return innerGetStake(v, err)
}

// Used to serve RPC queries
func GetStakeFromState(
	ctx context.Context,
	f ReadState,
	stake ids.ID,
) (bool, // exists
	ids.NodeID, // NodeID
	uint64, // StakedAmount
	uint64, // EndLockUp
	codec.Address, // Owner
	error) {
	values, errs := f(ctx, [][]byte{StakeKey(stake)})
	return innerGetStake(values[0], errs[0])
}

func innerGetStake(v []byte, err error) (
	bool, // exists
	ids.NodeID, // NodeID
	uint64, // StakedAmount
	uint64, // EndLockUp
	codec.Address, // Owner
	error,
) {
	if errors.Is(err, database.ErrNotFound) {
		return false, ids.EmptyNodeID, 0, 0, codec.Address{}, nil
	}
	if err != nil {
		return false, ids.EmptyNodeID, 0, 0, codec.Address{}, err
	}

	offset := 0
	var nodeID ids.NodeID
	copy(nodeID[:], v[offset:offset+consts.NodeIDLen])
	offset += consts.NodeIDLen

	stakedAmount := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len

	endLockUp := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len

	var walletAddress codec.Address
	copy(walletAddress[:], v[offset:offset+codec.AddressLen])

	return true, nodeID, stakedAmount, endLockUp, walletAddress, nil
}

func DeleteStake(
	ctx context.Context,
	mu state.Mutable,
	stake ids.ID,
) error {
	return mu.Remove(ctx, StakeKey(stake))
}

func HeightKey() (k []byte) {
	return heightKey
}

func TimestampKey() (k []byte) {
	return timestampKey
}

func FeeKey() (k []byte) {
	return feeKey
}

func IncomingWarpKeyPrefix(sourceChainID ids.ID, msgID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen*2)
	k[0] = incomingWarpPrefix
	copy(k[1:], sourceChainID[:])
	copy(k[1+consts.IDLen:], msgID[:])
	return k
}

func OutgoingWarpKeyPrefix(txID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = outgoingWarpPrefix
	copy(k[1:], txID[:])
	return k
}
