// Copyright (C) 2024, Nuklai. All rights reserved.
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
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

// [txPrefix] + [txID]
func TxKey(id ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen)
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
	units fees.Dimensions,
	fee uint64,
) error {
	k := TxKey(id)
	v := make([]byte, consts.Uint64Len+1+fees.DimensionsLen+consts.Uint64Len)
	binary.BigEndian.PutUint64(v, uint64(t))
	if success {
		v[consts.Uint64Len] = successByte
	} else {
		v[consts.Uint64Len] = failureByte
	}
	copy(v[consts.Uint64Len+1:], units.Bytes())
	binary.BigEndian.PutUint64(v[consts.Uint64Len+1+fees.DimensionsLen:], fee)
	return db.Put(k, v)
}

func GetTransaction(
	_ context.Context,
	db database.KeyValueReader,
	id ids.ID,
) (bool, int64, bool, fees.Dimensions, uint64, error) {
	k := TxKey(id)
	v, err := db.Get(k)
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, false, fees.Dimensions{}, 0, nil
	}
	if err != nil {
		return false, 0, false, fees.Dimensions{}, 0, err
	}
	t := int64(binary.BigEndian.Uint64(v))
	success := true
	if v[consts.Uint64Len] == failureByte {
		success = false
	}
	d, err := fees.UnpackDimensions(v[consts.Uint64Len+1 : consts.Uint64Len+1+fees.DimensionsLen])
	if err != nil {
		return false, 0, false, fees.Dimensions{}, 0, err
	}
	fee := binary.BigEndian.Uint64(v[consts.Uint64Len+1+fees.DimensionsLen:])
	return true, t, success, d, fee, nil
}

// [accountPrefix] + [address] + [asset]
func BalanceKey(addr codec.Address, asset ids.ID) (k []byte) {
	k = balanceKeyPool.Get().([]byte)
	k[0] = balancePrefix
	copy(k[1:], addr[:])
	copy(k[1+codec.AddressLen:], asset[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen+ids.IDLen:], BalanceChunks)
	return
}

// If locked is 0, then account does not exist
func GetBalance(
	ctx context.Context,
	im state.Immutable,
	addr codec.Address,
	asset ids.ID,
) (uint64, error) {
	key, bal, _, err := getBalance(ctx, im, addr, asset)
	balanceKeyPool.Put(key)
	return bal, err
}

func getBalance(
	ctx context.Context,
	im state.Immutable,
	addr codec.Address,
	asset ids.ID,
) ([]byte, uint64, bool, error) {
	k := BalanceKey(addr, asset)
	bal, exists, err := innerGetBalance(im.GetValue(ctx, k))
	return k, bal, exists, err
}

// Used to serve RPC queries
func GetBalanceFromState(
	ctx context.Context,
	f ReadState,
	addr codec.Address,
	asset ids.ID,
) (uint64, error) {
	k := BalanceKey(addr, asset)
	values, errs := f(ctx, [][]byte{k})
	bal, _, err := innerGetBalance(values[0], errs[0])
	balanceKeyPool.Put(k)
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
	asset ids.ID,
	balance uint64,
) error {
	k := BalanceKey(addr, asset)
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

func DeleteBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	asset ids.ID,
) error {
	return mu.Remove(ctx, BalanceKey(addr, asset))
}

func AddBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	asset ids.ID,
	amount uint64,
	create bool,
) error {
	key, bal, exists, err := getBalance(ctx, mu, addr, asset)
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
			"%w: could not add balance (asset=%s, bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			asset,
			bal,
			codec.MustAddressBech32(nconsts.HRP, addr),
			amount,
		)
	}
	return setBalance(ctx, mu, key, nbal)
}

func SubBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	asset ids.ID,
	amount uint64,
) error {
	key, bal, _, err := getBalance(ctx, mu, addr, asset)
	if err != nil {
		return err
	}
	nbal, err := smath.Sub(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not subtract balance (asset=%s, bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			asset,
			bal,
			codec.MustAddressBech32(nconsts.HRP, addr),
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
	k = make([]byte, 1+ids.IDLen*2)
	k[0] = incomingWarpPrefix
	copy(k[1:], sourceChainID[:])
	copy(k[1+ids.IDLen:], msgID[:])
	return k
}

func OutgoingWarpKeyPrefix(txID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen)
	k[0] = outgoingWarpPrefix
	copy(k[1:], txID[:])
	return k
}
