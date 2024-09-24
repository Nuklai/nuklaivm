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

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

const (
	BalanceChunks  uint16 = 1
	AssetChunks    uint16 = 16
	AssetNFTChunks uint16 = 10
)

var (
	heightKey    = []byte{heightPrefix}
	timestampKey = []byte{timestampPrefix}
	feeKey       = []byte{feePrefix}
)

// [accountPrefix] + [address] + [asset]
func BalanceKey(addr codec.Address, asset ids.ID) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+ids.IDLen+consts.Uint16Len)
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
	_, bal, _, err := getBalance(ctx, im, addr, asset)
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

func AddBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	asset ids.ID,
	amount uint64,
	create bool,
) (uint64, error) {
	key, bal, exists, err := getBalance(ctx, mu, addr, asset)
	if err != nil {
		return 0, err
	}
	// Don't add balance if account doesn't exist. This
	// can be useful when processing fee refunds.
	if !exists && !create {
		return 0, nil
	}
	nbal, err := smath.Add(bal, amount)
	if err != nil {
		return 0, fmt.Errorf(
			"%w: could not add balance (asset=%v, bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			asset,
			bal,
			addr,
			amount,
		)
	}
	return nbal, setBalance(ctx, mu, key, nbal)
}

func SubBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	asset ids.ID,
	amount uint64,
) (uint64, error) {
	key, bal, ok, err := getBalance(ctx, mu, addr, asset)
	if !ok {
		return 0, ErrInvalidAddress
	}
	if err != nil {
		return 0, err
	}
	nbal, err := smath.Sub(bal, amount)
	if err != nil {
		return 0, fmt.Errorf(
			"%w: could not subtract balance (asset=%v, bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			asset,
			bal,
			addr,
			amount,
		)
	}
	if nbal == 0 {
		// If there is no balance left, we should delete the record instead of
		// setting it to 0.
		return 0, mu.Remove(ctx, key)
	}
	return nbal, setBalance(ctx, mu, key, nbal)
}

func DeleteBalance(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	asset ids.ID,
) error {
	return mu.Remove(ctx, BalanceKey(addr, asset))
}

// [assetID]
func AssetKey(asset ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len)           // Length of prefix + assetID + AssetChunks
	k[0] = assetPrefix                                       // assetPrefix is a constant representing the asset category
	copy(k[1:], asset[:])                                    // Copy the assetID
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], AssetChunks) // Adding AssetChunks
	return
}

// Used to serve RPC queries
func GetAssetFromState(
	ctx context.Context,
	f ReadState,
	asset ids.ID,
) (bool, uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	values, errs := f(ctx, [][]byte{AssetKey(asset)})
	return innerGetAsset(values[0], errs[0])
}

func GetAsset(
	ctx context.Context,
	im state.Immutable,
	asset ids.ID,
) (bool, uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	k := AssetKey(asset)
	return innerGetAsset(im.GetValue(ctx, k))
}

func innerGetAsset(
	v []byte,
	err error,
) (bool, uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, nil, nil, 0, nil, nil, 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, nil
	}
	if err != nil {
		return false, 0, nil, nil, 0, nil, nil, 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, err
	}

	offset := uint16(0)
	assetType := v[offset]
	offset += consts.Uint8Len
	nameLen := binary.BigEndian.Uint16(v[offset:])
	offset += consts.Uint16Len
	name := v[offset : offset+nameLen]
	offset += nameLen
	symbolLen := binary.BigEndian.Uint16(v[offset:])
	offset += consts.Uint16Len
	symbol := v[offset : offset+symbolLen]
	offset += symbolLen
	decimals := v[offset]
	offset += consts.Uint8Len
	metadataLen := binary.BigEndian.Uint16(v[offset:])
	offset += consts.Uint16Len
	metadata := v[offset : offset+metadataLen]
	offset += metadataLen
	uriLen := binary.BigEndian.Uint16(v[offset:])
	offset += consts.Uint16Len
	uri := v[offset : offset+uriLen]
	offset += uriLen
	totalSupply := binary.BigEndian.Uint64(v[offset:])
	offset += consts.Uint64Len
	maxSupply := binary.BigEndian.Uint64(v[offset:])
	offset += consts.Uint64Len

	var owner codec.Address
	copy(owner[:], v[offset:])
	offset += codec.AddressLen
	var mintAdmin codec.Address
	copy(mintAdmin[:], v[offset:])
	offset += codec.AddressLen
	var pauseUnpauseAdmin codec.Address
	copy(pauseUnpauseAdmin[:], v[offset:])
	offset += codec.AddressLen
	var freezeUnfreezeAdmin codec.Address
	copy(freezeUnfreezeAdmin[:], v[offset:])
	offset += codec.AddressLen
	var enableDisableKYCAccountAdmin codec.Address
	copy(enableDisableKYCAccountAdmin[:], v[offset:])

	return true, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, nil
}

func SetAsset(
	ctx context.Context,
	mu state.Mutable,
	asset ids.ID,
	assetType uint8,
	name []byte,
	symbol []byte,
	decimals uint8,
	metadata []byte,
	uri []byte,
	totalSupply uint64,
	maxSupply uint64,
	owner codec.Address,
	mintAdmin codec.Address,
	pauseUnpauseAdmin codec.Address,
	freezeUnfreezeAdmin codec.Address,
	enableDisableKYCAccountAdmin codec.Address,
) error {
	k := AssetKey(asset)
	nameLen := len(name)
	symbolLen := len(symbol)
	metadataLen := len(metadata)
	uriLen := len(uri)

	v := make([]byte, consts.Uint8Len+consts.Uint16Len+nameLen+consts.Uint16Len+symbolLen+consts.Uint8Len+consts.Uint16Len+metadataLen+consts.Uint16Len+uriLen+consts.Uint64Len*2+codec.AddressLen*5)

	offset := 0
	v[offset] = assetType
	offset += consts.Uint8Len
	binary.BigEndian.PutUint16(v[offset:], uint16(nameLen))
	offset += consts.Uint16Len
	copy(v[offset:], name)
	offset += nameLen
	binary.BigEndian.PutUint16(v[offset:], uint16(symbolLen))
	offset += consts.Uint16Len
	copy(v[offset:], symbol)
	offset += symbolLen
	v[offset] = decimals
	offset += consts.Uint8Len
	binary.BigEndian.PutUint16(v[offset:], uint16(metadataLen))
	offset += consts.Uint16Len
	copy(v[offset:], metadata)
	offset += metadataLen
	binary.BigEndian.PutUint16(v[offset:], uint16(uriLen))
	offset += consts.Uint16Len
	copy(v[offset:], uri)
	offset += uriLen
	binary.BigEndian.PutUint64(v[offset:], totalSupply)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], maxSupply)
	offset += consts.Uint64Len
	copy(v[offset:], owner[:])
	offset += codec.AddressLen
	copy(v[offset:], mintAdmin[:])
	offset += codec.AddressLen
	copy(v[offset:], pauseUnpauseAdmin[:])
	offset += codec.AddressLen
	copy(v[offset:], freezeUnfreezeAdmin[:])
	offset += codec.AddressLen
	copy(v[offset:], enableDisableKYCAccountAdmin[:])

	return mu.Insert(ctx, k, v)
}

func DeleteAsset(ctx context.Context, mu state.Mutable, asset ids.ID) error {
	k := AssetKey(asset)
	return mu.Remove(ctx, k)
}

// [nftID]
func AssetNFTKey(nftID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len)              // Length of prefix + nftID + AssetNFTChunks
	k[0] = assetNFTPrefix                                       // assetNFTPrefix is a constant representing the assetNFT category
	copy(k[1:], nftID[:])                                       // Copy the nftID
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], AssetNFTChunks) // Adding AssetNFTChunks
	return
}

// Used to serve RPC queries
func GetAssetNFTFromState(
	ctx context.Context,
	f ReadState,
	nftID ids.ID,
) (bool, ids.ID, uint64, []byte, []byte, codec.Address, error) {
	values, errs := f(ctx, [][]byte{AssetNFTKey(nftID)})
	return innerGetAssetNFT(values[0], errs[0])
}

func GetAssetNFT(
	ctx context.Context,
	im state.Immutable,
	nftID ids.ID,
) (bool, ids.ID, uint64, []byte, []byte, codec.Address, error) {
	k := AssetNFTKey(nftID)
	return innerGetAssetNFT(im.GetValue(ctx, k))
}

func innerGetAssetNFT(v []byte, err error) (bool, ids.ID, uint64, []byte, []byte, codec.Address, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, ids.Empty, 0, nil, nil, codec.Address{}, nil
	}
	if err != nil {
		return false, ids.Empty, 0, nil, nil, codec.Address{}, err
	}

	collectionID, err := ids.ToID(v[:ids.IDLen])
	if err != nil {
		return false, ids.Empty, 0, nil, nil, codec.Address{}, err
	}
	uniqueID := binary.BigEndian.Uint64(v[ids.IDLen:])
	uriLen := binary.BigEndian.Uint16(v[ids.IDLen+consts.Uint64Len:])
	uri := v[ids.IDLen+consts.Uint64Len+consts.Uint16Len : ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen]
	metadataLen := binary.BigEndian.Uint16(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen:])
	metadata := v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+consts.Uint16Len : ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+consts.Uint16Len+metadataLen]
	var owner codec.Address
	copy(owner[:], v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+consts.Uint16Len+metadataLen:])

	return true, collectionID, uniqueID, uri, metadata, owner, nil
}

func SetAssetNFT(ctx context.Context, mu state.Mutable, collectionID ids.ID, uniqueID uint64, nftID ids.ID, uri []byte, metadata []byte, owner codec.Address) error {
	k := AssetNFTKey(nftID)
	uriLen := len(uri)
	metadataLen := len(metadata)

	v := make([]byte, ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+consts.Uint16Len+metadataLen+codec.AddressLen)
	copy(v, collectionID[:])
	binary.BigEndian.PutUint64(v[ids.IDLen:], uniqueID)
	binary.BigEndian.PutUint16(v[ids.IDLen+consts.Uint64Len:], uint16(uriLen))
	copy(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len:], uri)
	binary.BigEndian.PutUint16(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen:], uint16(metadataLen))
	copy(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+consts.Uint16Len:], metadata)
	copy(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+consts.Uint16Len+metadataLen:], owner[:])

	return mu.Insert(ctx, k, v)
}

func DeleteAssetNFT(ctx context.Context, mu state.Mutable, nftID ids.ID) error {
	k := AssetNFTKey(nftID)
	return mu.Remove(ctx, k)
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
