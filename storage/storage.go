// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	hmath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

// Metadata
// 0x0/ (tx)
//   -> [txID] => timestamp
//
// State
// / (height) => store in root
//   -> [heightPrefix] => height
//
// 0x0/ (balance)
//   -> [owner|asset] => balance

// 0x1/ (hypersdk-height)
// 0x2/ (hypersdk-timestamp)
// 0x3/ (hypersdk-fee)

// 0x4/ (hypersdk-incoming warp)
// 0x5/ (hypersdk-outgoing warp)

// 0x6/ (assets)
//   -> [asset] => metadataLen|metadata|supply|owner|warp
// 0x7/ (loans)
//   -> [assetID|destination] => amount

// 0x8/ (stake)
//   -> [nodeID] => stakeStartBlock|stakeEndBlock|stakedAmount|delegationFeeRate|rewardAddress|ownerAddress
// 0x9/ (delegate)
//   -> [owner|nodeID] => stakeStartBlock|stakedAmount|rewardAddress|ownerAddress

const (
	// metaDB
	txPrefix = 0x0

	// stateDB
	balancePrefix = 0x0

	heightPrefix    = 0x1
	timestampPrefix = 0x2
	feePrefix       = 0x3

	incomingWarpPrefix = 0x4
	outgoingWarpPrefix = 0x5

	assetPrefix = 0x6
	loanPrefix  = 0x7

	registerValidatorStakePrefix = 0x8
	delegateUserStakePrefix      = 0x9
)

const (
	BalanceChunks                uint16 = 1
	AssetChunks                  uint16 = 5
	LoanChunks                   uint16 = 1
	RegisterValidatorStakeChunks uint16 = 5
	DelegateUserStakeChunks      uint16 = 3
)

var (
	failureByte  = byte(0x0)
	successByte  = byte(0x1)
	heightKey    = []byte{heightPrefix}
	timestampKey = []byte{timestampPrefix}
	feeKey       = []byte{feePrefix}

	balanceKeyPool = sync.Pool{
		New: func() any {
			return make([]byte, 1+codec.AddressLen+hconsts.IDLen+hconsts.Uint16Len)
		},
	}
)

// [txPrefix] + [txID]
func TxKey(id ids.ID) (k []byte) {
	k = make([]byte, 1+hconsts.IDLen)
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
	v := make([]byte, hconsts.Uint64Len+1+chain.DimensionsLen+hconsts.Uint64Len)
	binary.BigEndian.PutUint64(v, uint64(t))
	if success {
		v[hconsts.Uint64Len] = successByte
	} else {
		v[hconsts.Uint64Len] = failureByte
	}
	copy(v[hconsts.Uint64Len+1:], units.Bytes())
	binary.BigEndian.PutUint64(v[hconsts.Uint64Len+1+chain.DimensionsLen:], fee)
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
	if v[hconsts.Uint64Len] == failureByte {
		success = false
	}
	d, err := chain.UnpackDimensions(v[hconsts.Uint64Len+1 : hconsts.Uint64Len+1+chain.DimensionsLen])
	if err != nil {
		return false, 0, false, chain.Dimensions{}, 0, err
	}
	fee := binary.BigEndian.Uint64(v[hconsts.Uint64Len+1+chain.DimensionsLen:])
	return true, t, success, d, fee, nil
}

// [accountPrefix] + [address] + [asset]
func BalanceKey(addr codec.Address, asset ids.ID) (k []byte) {
	k = balanceKeyPool.Get().([]byte)
	k[0] = balancePrefix
	copy(k[1:], addr[:])
	copy(k[1+codec.AddressLen:], asset[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen+hconsts.IDLen:], BalanceChunks)
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
	nbal, err := hmath.Add64(bal, amount)
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
	nbal, err := hmath.Sub(bal, amount)
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

// [assetPrefix] + [address]
func AssetKey(asset ids.ID) (k []byte) {
	k = make([]byte, 1+hconsts.IDLen+hconsts.Uint16Len)
	k[0] = assetPrefix
	copy(k[1:], asset[:])
	binary.BigEndian.PutUint16(k[1+hconsts.IDLen:], AssetChunks)
	return
}

// Used to serve RPC queries
func GetAssetFromState(
	ctx context.Context,
	f ReadState,
	asset ids.ID,
) (bool, []byte, uint8, []byte, uint64, codec.Address, bool, error) {
	values, errs := f(ctx, [][]byte{AssetKey(asset)})
	return innerGetAsset(values[0], errs[0])
}

func GetAsset(
	ctx context.Context,
	im state.Immutable,
	asset ids.ID,
) (bool, []byte, uint8, []byte, uint64, codec.Address, bool, error) {
	k := AssetKey(asset)
	return innerGetAsset(im.GetValue(ctx, k))
}

func innerGetAsset(
	v []byte,
	err error,
) (bool, []byte, uint8, []byte, uint64, codec.Address, bool, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, nil, 0, nil, 0, codec.EmptyAddress, false, nil
	}
	if err != nil {
		return false, nil, 0, nil, 0, codec.EmptyAddress, false, err
	}
	symbolLen := binary.BigEndian.Uint16(v)
	symbol := v[hconsts.Uint16Len : hconsts.Uint16Len+symbolLen]
	decimals := v[hconsts.Uint16Len+symbolLen]
	metadataLen := binary.BigEndian.Uint16(v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len:])
	metadata := v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len : hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen]
	supply := binary.BigEndian.Uint64(v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen:])
	var addr codec.Address
	copy(addr[:], v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen+hconsts.Uint64Len:])
	warp := v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen+hconsts.Uint64Len+codec.AddressLen] == 0x1
	return true, symbol, decimals, metadata, supply, addr, warp, nil
}

func SetAsset(
	ctx context.Context,
	mu state.Mutable,
	asset ids.ID,
	symbol []byte,
	decimals uint8,
	metadata []byte,
	supply uint64,
	owner codec.Address,
	warp bool,
) error {
	k := AssetKey(asset)
	symbolLen := len(symbol)
	metadataLen := len(metadata)
	v := make([]byte, hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen+hconsts.Uint64Len+codec.AddressLen+1)
	binary.BigEndian.PutUint16(v, uint16(symbolLen))
	copy(v[hconsts.Uint16Len:], symbol)
	v[hconsts.Uint16Len+symbolLen] = decimals
	binary.BigEndian.PutUint16(v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len:], uint16(metadataLen))
	copy(v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len:], metadata)
	binary.BigEndian.PutUint64(v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen:], supply)
	copy(v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen+hconsts.Uint64Len:], owner[:])
	b := byte(0x0)
	if warp {
		b = 0x1
	}
	v[hconsts.Uint16Len+symbolLen+hconsts.Uint8Len+hconsts.Uint16Len+metadataLen+hconsts.Uint64Len+codec.AddressLen] = b
	return mu.Insert(ctx, k, v)
}

func DeleteAsset(ctx context.Context, mu state.Mutable, asset ids.ID) error {
	k := AssetKey(asset)
	return mu.Remove(ctx, k)
}

// [loanPrefix] + [asset] + [destination]
func LoanKey(asset ids.ID, destination ids.ID) (k []byte) {
	k = make([]byte, 1+hconsts.IDLen*2+hconsts.Uint16Len)
	k[0] = loanPrefix
	copy(k[1:], asset[:])
	copy(k[1+hconsts.IDLen:], destination[:])
	binary.BigEndian.PutUint16(k[1+hconsts.IDLen*2:], LoanChunks)
	return
}

// Used to serve RPC queries
func GetLoanFromState(
	ctx context.Context,
	f ReadState,
	asset ids.ID,
	destination ids.ID,
) (uint64, error) {
	values, errs := f(ctx, [][]byte{LoanKey(asset, destination)})
	return innerGetLoan(values[0], errs[0])
}

func innerGetLoan(v []byte, err error) (uint64, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}

func GetLoan(
	ctx context.Context,
	im state.Immutable,
	asset ids.ID,
	destination ids.ID,
) (uint64, error) {
	k := LoanKey(asset, destination)
	v, err := im.GetValue(ctx, k)
	return innerGetLoan(v, err)
}

func SetLoan(
	ctx context.Context,
	mu state.Mutable,
	asset ids.ID,
	destination ids.ID,
	amount uint64,
) error {
	k := LoanKey(asset, destination)
	return mu.Insert(ctx, k, binary.BigEndian.AppendUint64(nil, amount))
}

func AddLoan(
	ctx context.Context,
	mu state.Mutable,
	asset ids.ID,
	destination ids.ID,
	amount uint64,
) error {
	loan, err := GetLoan(ctx, mu, asset, destination)
	if err != nil {
		return err
	}
	nloan, err := hmath.Add64(loan, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not add loan (asset=%s, destination=%s, amount=%d)",
			ErrInvalidBalance,
			asset,
			destination,
			amount,
		)
	}
	return SetLoan(ctx, mu, asset, destination, nloan)
}

func SubLoan(
	ctx context.Context,
	mu state.Mutable,
	asset ids.ID,
	destination ids.ID,
	amount uint64,
) error {
	loan, err := GetLoan(ctx, mu, asset, destination)
	if err != nil {
		return err
	}
	nloan, err := hmath.Sub(loan, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not subtract loan (asset=%s, destination=%s, amount=%d)",
			ErrInvalidBalance,
			asset,
			destination,
			amount,
		)
	}
	if nloan == 0 {
		// If there is no balance left, we should delete the record instead of
		// setting it to 0.
		return mu.Remove(ctx, LoanKey(asset, destination))
	}
	return SetLoan(ctx, mu, asset, destination, nloan)
}

// [registerValidatorStakePrefix] + [nodeID]
func RegisterValidatorStakeKey(nodeID ids.NodeID) (k []byte) {
	k = make([]byte, 1+hconsts.NodeIDLen+hconsts.Uint16Len) // Length of prefix + nodeID + RegisterValidatorStakeChunks
	k[0] = registerValidatorStakePrefix                     // registerValidatorStakePrefix is a constant representing the registerValidatorStake category
	copy(k[1:], nodeID[:])
	binary.BigEndian.PutUint16(k[1+hconsts.NodeIDLen:], RegisterValidatorStakeChunks) // Adding RegisterValidatorStakeChunks
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
	v := make([]byte, (4*hconsts.Uint64Len)+(2*codec.AddressLen)) // Calculate the length of the encoded data

	offset := 0
	binary.BigEndian.PutUint64(v[offset:], stakeStartBlock)
	offset += hconsts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakeEndBlock)
	offset += hconsts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakedAmount)
	offset += hconsts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], delegationFeeRate)
	offset += hconsts.Uint64Len

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
	stakeStartBlock := binary.BigEndian.Uint64(v[offset : offset+hconsts.Uint64Len])
	offset += hconsts.Uint64Len
	stakeEndBlock := binary.BigEndian.Uint64(v[offset : offset+hconsts.Uint64Len])
	offset += hconsts.Uint64Len
	stakedAmount := binary.BigEndian.Uint64(v[offset : offset+hconsts.Uint64Len])
	offset += hconsts.Uint64Len
	delegationFeeRate := binary.BigEndian.Uint64(v[offset : offset+hconsts.Uint64Len])
	offset += hconsts.Uint64Len

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
	k = make([]byte, 1+codec.AddressLen+hconsts.NodeIDLen+hconsts.Uint16Len) // Length of prefix + owner + nodeID + DelegateUserStakeChunks
	k[0] = delegateUserStakePrefix                                           // delegateUserStakePrefix is a constant representing the staking category
	copy(k[1:], owner[:])
	copy(k[1+codec.AddressLen:], nodeID[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen+hconsts.NodeIDLen:], DelegateUserStakeChunks) // Adding DelegateUserStakeChunks
	return
}

func SetDelegateUserStake(
	ctx context.Context,
	mu state.Mutable,
	owner codec.Address,
	nodeID ids.NodeID,
	stakeStartBlock uint64,
	stakedAmount uint64,
	rewardAddress codec.Address,
) error {
	key := DelegateUserStakeKey(owner, nodeID)
	v := make([]byte, 2*hconsts.Uint64Len+2*codec.AddressLen) // Calculate the length of the encoded data

	offset := 0

	binary.BigEndian.PutUint64(v[offset:], stakeStartBlock)
	offset += hconsts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], stakedAmount)
	offset += hconsts.Uint64Len

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
	uint64, // StakedAmount
	codec.Address, // RewardAddress
	codec.Address, // OwnerAddress
	error,
) {
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, 0, codec.Address{}, codec.Address{}, nil
	}
	if err != nil {
		return false, 0, 0, codec.Address{}, codec.Address{}, nil
	}

	offset := 0

	stakeStartBlock := binary.BigEndian.Uint64(v[offset : offset+hconsts.Uint64Len])
	offset += hconsts.Uint64Len
	stakedAmount := binary.BigEndian.Uint64(v[offset : offset+hconsts.Uint64Len])
	offset += hconsts.Uint64Len

	var rewardAddress codec.Address
	copy(rewardAddress[:], v[offset:offset+codec.AddressLen])
	offset += codec.AddressLen
	var ownerAddress codec.Address
	copy(ownerAddress[:], v[offset:offset+codec.AddressLen])

	return true, stakeStartBlock, stakedAmount, rewardAddress, ownerAddress, nil
}

func DeleteDelegateUserStake(
	ctx context.Context,
	mu state.Mutable,
	owner codec.Address,
	nodeID ids.NodeID,
) error {
	return mu.Remove(ctx, DelegateUserStakeKey(owner, nodeID))
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
	k = make([]byte, 1+hconsts.IDLen*2)
	k[0] = incomingWarpPrefix
	copy(k[1:], sourceChainID[:])
	copy(k[1+hconsts.IDLen:], msgID[:])
	return k
}

func OutgoingWarpKeyPrefix(txID ids.ID) (k []byte) {
	k = make([]byte, 1+hconsts.IDLen)
	k[0] = outgoingWarpPrefix
	copy(k[1:], txID[:])
	return k
}
