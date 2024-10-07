// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"

	"github.com/ava-labs/avalanchego/database"
	smath "github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	AssetAccountBalanceChunks uint16 = 1
	AssetInfoChunks           uint16 = 13
)

const (
	MaxNameSize          = 64
	MaxSymbolSize        = 24
	MaxAssetMetadataSize = 256
	MaxTextSize          = 256
	MaxAssetDecimals     = 9
)

func AssetAddress(assetType uint8, name []byte, symbol []byte, decimals uint8, metadata []byte, uri []byte, owner codec.Address) codec.Address {
	v := make([]byte, len(name)+len(symbol)+consts.Uint8Len+len(metadata)+len(uri)+codec.AddressLen)
	offset := 0
	copy(v[offset:], name)
	offset += len(name)
	copy(v[offset:], symbol)
	offset += len(symbol)
	v[offset] = decimals
	offset += consts.Uint8Len
	copy(v[offset:], metadata)
	offset += len(metadata)
	copy(v[offset:], uri)
	offset += len(uri)
	copy(v[offset:], owner[:])
	id := utils.ToID(v)
	return codec.CreateAddress(assetType, id)
}

func AssetAddressNFT(assetAddress codec.Address, metadata []byte, owner codec.Address) codec.Address {
	v := make([]byte, codec.AddressLen+len(metadata)+codec.AddressLen)
	offset := 0
	copy(v[offset:], assetAddress[:])
	offset += codec.AddressLen
	copy(v[offset:], metadata)
	offset += len(metadata)
	copy(v[offset:], owner[:])
	id := utils.ToID(v)
	return codec.CreateAddress(nconsts.AssetNonFungibleTokenID, id)
}

func AssetAddressFractional(assetAddress codec.Address) codec.Address {
	id := utils.ToID(assetAddress[:])
	return codec.CreateAddress(nconsts.AssetFractionalTokenID, id)
}

func AssetInfoKey(assetAddress codec.Address) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+consts.Uint16Len)               // Length of prefix + assetAddress + AssetInfoChunks
	k[0] = assetInfoPrefix                                              // assetInfoPrefix is a constant representing the asset category
	copy(k[1:1+codec.AddressLen], assetAddress[:])                      // Copy the assetAddress
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], AssetInfoChunks) // Adding AssetInfoChunks
	return
}

func AssetAccountBalanceKey(asset codec.Address, account codec.Address) []byte {
	k := make([]byte, 1+codec.AddressLen+codec.AddressLen+consts.Uint16Len)
	k[0] = assetAccountBalancePrefix
	copy(k[1:], asset[:])
	copy(k[1+codec.AddressLen:], account[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen+codec.AddressLen:], AssetAccountBalanceChunks)
	return k
}

func SetAssetInfo(
	ctx context.Context,
	mu state.Mutable,
	assetAddress codec.Address,
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
	// Setup
	k := AssetInfoKey(assetAddress)
	nameLen := len(name)
	symbolLen := len(symbol)
	metadataLen := len(metadata)
	uriLen := len(uri)
	assetInfoSize := consts.Uint8Len + consts.Uint16Len + nameLen + consts.Uint16Len + symbolLen + consts.Uint8Len + consts.Uint16Len + metadataLen + consts.Uint16Len + uriLen + consts.Uint64Len*2 + codec.AddressLen*5
	v := make([]byte, assetInfoSize)

	// Populate
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

// Used to serve RPC queries
func GetAssetInfoFromState(
	ctx context.Context,
	f ReadState,
	asset codec.Address,
) (uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	values, errs := f(ctx, [][]byte{AssetInfoKey(asset)})
	if errs[0] != nil {
		return 0, nil, nil, 0, nil, nil, 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, errs[0]
	}
	return innerGetAssetInfo(values[0])
}

func GetAssetInfoNoController(
	ctx context.Context,
	im state.Immutable,
	asset codec.Address,
) (uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	k := AssetInfoKey(asset)
	v, err := im.GetValue(ctx, k)
	if err != nil {
		return 0, nil, nil, 0, nil, nil, 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, err
	}
	return innerGetAssetInfo(v)
}

func innerGetAssetInfo(
	v []byte,
) (uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	// Extract
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

	return assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, nil
}

func MintAsset(ctx context.Context, mu state.Mutable, asset codec.Address, to codec.Address, mintAmount uint64) (uint64, error) {
	// Get asset info + account
	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := GetAssetInfoNoController(ctx, mu, asset)
	if err != nil {
		return 0, err
	}
	balance, err := GetAssetAccountBalanceNoController(ctx, mu, asset, to)
	if err != nil {
		return 0, err
	}
	newTotalSupply, err := smath.Add(totalSupply, mintAmount)
	if err != nil {
		return 0, err
	}
	// Ensure minting doesn't exceed max supply
	if maxSupply != 0 && newTotalSupply > maxSupply {
		return 0, ErrMaxSupplyExceeded
	}
	newBalance, err := smath.Add(balance, mintAmount)
	if err != nil {
		return 0, err
	}
	// Update asset info
	if err := SetAssetInfo(ctx, mu, asset, assetType, name, symbol, decimals, metadata, uri, newTotalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return 0, err
	}
	// Update asset account
	if err := SetAssetAccountBalance(ctx, mu, asset, to, newBalance); err != nil {
		return 0, err
	}
	return newBalance, nil
}

func SetAssetAccountBalance(
	ctx context.Context,
	mu state.Mutable,
	assetAddress codec.Address,
	account codec.Address,
	balance uint64,
) error {
	k := AssetAccountBalanceKey(assetAddress, account)
	v := make([]byte, consts.Uint64Len)
	binary.BigEndian.PutUint64(v, balance)
	return mu.Insert(ctx, k, v)
}

func GetAssetAccountBalanceNoController(
	ctx context.Context,
	mu state.Immutable,
	assetAddress codec.Address,
	account codec.Address,
) (uint64, error) {
	k := AssetAccountBalanceKey(assetAddress, account)
	v, err := mu.GetValue(ctx, k)
	if err == database.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}

func GetAssetAccountBalanceFromState(
	ctx context.Context,
	f ReadState,
	assetAddress codec.Address,
	account codec.Address,
) (uint64, error) {
	k := AssetAccountBalanceKey(assetAddress, account)
	values, errs := f(ctx, [][]byte{k})
	if errs[0] == database.ErrNotFound {
		return 0, nil
	} else if errs[0] != nil {
		return 0, errs[0]
	}
	return binary.BigEndian.Uint64(values[0]), nil
}

func BurnAsset(
	ctx context.Context,
	mu state.Mutable,
	assetAddress codec.Address,
	from codec.Address,
	value uint64,
) (uint64, error) {
	balance, err := GetAssetAccountBalanceNoController(ctx, mu, assetAddress, from)
	if err != nil {
		return 0, err
	}
	// Ensure that the balance is sufficient
	if balance < value {
		return 0, ErrInsufficientAssetBalance
	}

	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := GetAssetInfoNoController(ctx, mu, assetAddress)
	if err != nil {
		return 0, err
	}

	newBalance, err := smath.Sub(balance, value)
	if err != nil {
		return 0, err
	}
	newTotalSupply, err := smath.Sub(totalSupply, value)
	if err != nil {
		return 0, err
	}

	if err = SetAssetAccountBalance(ctx, mu, assetAddress, from, newBalance); err != nil {
		return 0, err
	}
	if err = SetAssetInfo(ctx, mu, assetAddress, assetType, name, symbol, decimals, metadata, uri, newTotalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin); err != nil {
		return 0, err
	}

	return newBalance, nil
}

func TransferAsset(
	ctx context.Context,
	mu state.Mutable,
	assetAddress codec.Address,
	from codec.Address,
	to codec.Address,
	value uint64,
) (uint64, uint64, error) {
	fromBalance, err := GetAssetAccountBalanceNoController(ctx, mu, assetAddress, from)
	if err != nil {
		return 0, 0, err
	}
	toBalance, err := GetAssetAccountBalanceNoController(ctx, mu, assetAddress, to)
	if err != nil {
		return 0, 0, err
	}
	newFromBalance, err := smath.Sub(fromBalance, value)
	if err != nil {
		return 0, 0, err
	}
	newToBalance, err := smath.Add(toBalance, value)
	if err != nil {
		return 0, 0, err
	}
	if err = SetAssetAccountBalance(ctx, mu, assetAddress, from, newFromBalance); err != nil {
		return 0, 0, err
	}
	if err = SetAssetAccountBalance(ctx, mu, assetAddress, to, newToBalance); err != nil {
		return 0, 0, err
	}
	// Handle NFTs
	assetType, _, _, _, _, nftCollectionAddressBytes, _, _, _, _, _, _, _, err := GetAssetInfoNoController(ctx, mu, assetAddress)
	if err != nil {
		return 0, 0, err
	}
	if assetType == nconsts.AssetNonFungibleTokenID {
		nftCollectionAddress, err := codec.ToAddress(nftCollectionAddressBytes)
		if err != nil {
			return 0, 0, err
		}
		if err = SetAssetAccountBalance(ctx, mu, nftCollectionAddress, from, newFromBalance); err != nil {
			return 0, 0, err
		}
		if err = SetAssetAccountBalance(ctx, mu, nftCollectionAddress, to, newToBalance); err != nil {
			return 0, 0, err
		}
	}
	return newFromBalance, newToBalance, nil
}

func DeleteAsset(
	ctx context.Context,
	mu state.Mutable,
	assetAddress codec.Address,
) error {
	k := AssetInfoKey(assetAddress)
	return mu.Remove(ctx, k)
}

func AssetExists(
	ctx context.Context,
	mu state.Immutable,
	assetAddress codec.Address,
) bool {
	v, err := mu.GetValue(ctx, AssetInfoKey(assetAddress))
	return v != nil && err == nil
}
