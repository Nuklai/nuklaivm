// Copyright (C) 2024, Nuklai. All rights reserved.
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
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/fees"
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

// 0x7/ (stake)
//   -> [nodeID] => stakeStartBlock|stakeEndBlock|stakedAmount|delegationFeeRate|rewardAddress|ownerAddress
// 0x8/ (delegate)
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

	assetPrefix    = 0x6
	assetNFTPrefix = 0x7

	registerValidatorStakePrefix = 0x8
	delegateUserStakePrefix      = 0x9

	datasetPrefix = 0xA
)

const (
	BalanceChunks                uint16 = 1
	AssetChunks                  uint16 = 9
	AssetNFTChunks               uint16 = 3
	RegisterValidatorStakeChunks uint16 = 4
	DelegateUserStakeChunks      uint16 = 2
	DatasetChunks                uint16 = 91
)

var (
	failureByte  = byte(0x0)
	successByte  = byte(0x1)
	heightKey    = []byte{heightPrefix}
	timestampKey = []byte{timestampPrefix}
	feeKey       = []byte{feePrefix}

	balanceKeyPool = sync.Pool{
		New: func() any {
			return make([]byte, 1+codec.AddressLen+ids.IDLen+consts.Uint16Len)
		},
	}
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
) (bool, uint8, []byte, []byte, uint8, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	values, errs := f(ctx, [][]byte{AssetKey(asset)})
	return innerGetAsset(values[0], errs[0])
}

func GetAsset(
	ctx context.Context,
	im state.Immutable,
	asset ids.ID,
) (bool, uint8, []byte, []byte, uint8, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	k := AssetKey(asset)
	return innerGetAsset(im.GetValue(ctx, k))
}

func innerGetAsset(
	v []byte,
	err error,
) (bool, uint8, []byte, []byte, uint8, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, nil, nil, 0, nil, 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, nil
	}
	if err != nil {
		return false, 0, nil, nil, 0, nil, 0, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, err
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
	totalSupply := binary.BigEndian.Uint64(v[offset:])
	offset += consts.Uint64Len
	maxSupply := binary.BigEndian.Uint64(v[offset:])
	offset += consts.Uint64Len

	var updateAssetActor codec.Address
	copy(updateAssetActor[:], v[offset:])
	offset += codec.AddressLen
	var mintActor codec.Address
	copy(mintActor[:], v[offset:])
	offset += codec.AddressLen
	var pauseUnpauseActor codec.Address
	copy(pauseUnpauseActor[:], v[offset:])
	offset += codec.AddressLen
	var freezeUnfreezeActor codec.Address
	copy(freezeUnfreezeActor[:], v[offset:])
	offset += codec.AddressLen
	var enableDisableKYCAccountActor codec.Address
	copy(enableDisableKYCAccountActor[:], v[offset:])
	offset += codec.AddressLen
	var deleteActor codec.Address
	copy(deleteActor[:], v[offset:])

	return true, assetType, name, symbol, decimals, metadata, totalSupply, maxSupply, updateAssetActor, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, deleteActor, nil
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
	totalSupply uint64,
	maxSupply uint64,
	updateAssetActor codec.Address,
	mintActor codec.Address,
	pauseUnpauseActor codec.Address,
	freezeUnfreezeActor codec.Address,
	enableDisableKYCAccountActor codec.Address,
	deleteActor codec.Address,
) error {
	k := AssetKey(asset)
	nameLen := len(name)
	symbolLen := len(symbol)
	metadataLen := len(metadata)

	v := make([]byte, consts.Uint8Len+consts.Uint16Len+nameLen+consts.Uint16Len+symbolLen+consts.Uint8Len+consts.Uint16Len+metadataLen+consts.Uint64Len*2+codec.AddressLen*6)

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
	binary.BigEndian.PutUint64(v[offset:], totalSupply)
	offset += consts.Uint64Len
	binary.BigEndian.PutUint64(v[offset:], maxSupply)
	offset += consts.Uint64Len
	copy(v[offset:], updateAssetActor[:])
	offset += codec.AddressLen
	copy(v[offset:], mintActor[:])
	offset += codec.AddressLen
	copy(v[offset:], pauseUnpauseActor[:])
	offset += codec.AddressLen
	copy(v[offset:], freezeUnfreezeActor[:])
	offset += codec.AddressLen
	copy(v[offset:], enableDisableKYCAccountActor[:])
	offset += codec.AddressLen
	copy(v[offset:], deleteActor[:])

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
) (bool, ids.ID, uint64, []byte, codec.Address, error) {
	values, errs := f(ctx, [][]byte{AssetNFTKey(nftID)})
	return innerGetAssetNFT(values[0], errs[0])
}

func GetAssetNFT(
	ctx context.Context,
	im state.Immutable,
	nftID ids.ID,
) (bool, ids.ID, uint64, []byte, codec.Address, error) {
	k := AssetNFTKey(nftID)
	return innerGetAssetNFT(im.GetValue(ctx, k))
}

func innerGetAssetNFT(v []byte, err error) (bool, ids.ID, uint64, []byte, codec.Address, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, ids.Empty, 0, nil, codec.Address{}, nil
	}
	if err != nil {
		return false, ids.Empty, 0, nil, codec.Address{}, err
	}

	collectionID, err := ids.ToID(v[:ids.IDLen])
	if err != nil {
		return false, ids.Empty, 0, nil, codec.Address{}, err
	}
	uniqueID := binary.BigEndian.Uint64(v[ids.IDLen:])
	uriLen := binary.BigEndian.Uint16(v[ids.IDLen+consts.Uint64Len:])
	uri := v[ids.IDLen+consts.Uint64Len+consts.Uint16Len : ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen]
	var to codec.Address
	copy(to[:], v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen:])

	return true, collectionID, uniqueID, uri, to, nil
}

func SetAssetNFT(ctx context.Context, mu state.Mutable, collectionID ids.ID, uniqueID uint64, nftID ids.ID, uri []byte, to codec.Address) error {
	k := AssetNFTKey(nftID)
	uriLen := len(uri)

	v := make([]byte, ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen+codec.AddressLen)
	copy(v, collectionID[:])
	binary.BigEndian.PutUint64(v[ids.IDLen:], uniqueID)
	binary.BigEndian.PutUint16(v[ids.IDLen+consts.Uint64Len:], uint16(uriLen))
	copy(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len:], uri)
	copy(v[ids.IDLen+consts.Uint64Len+consts.Uint16Len+uriLen:], to[:])

	return mu.Insert(ctx, k, v)
}

func DeleteAssetNFT(ctx context.Context, mu state.Mutable, nftID ids.ID) error {
	k := AssetNFTKey(nftID)
	return mu.Remove(ctx, k)
}

// [datasetID]
func AssetDatasetKey(datasetID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len) // Length of prefix + datasetID + DatasetChunks
	k[0] = datasetPrefix                           // datasetPrefix is a constant representing the dataset category
	copy(k[1:], datasetID[:])                      // Copy the datasetID
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], DatasetChunks)
	return
}

// Used to serve RPC queries
func GetAssetDatasetFromState(
	ctx context.Context,
	f ReadState,
	datasetID ids.ID,
) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	values, errs := f(ctx, [][]byte{AssetDatasetKey(datasetID)})
	return innerGetAssetDataset(values[0], errs[0])
}

func GetAssetDataset(
	ctx context.Context,
	im state.Immutable,
	datasetID ids.ID,
) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	k := AssetDatasetKey(datasetID)
	return innerGetAssetDataset(im.GetValue(ctx, k))
}

func innerGetAssetDataset(v []byte, err error) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, nil, nil, nil, nil, nil, nil, nil, false, false, ids.Empty, 0, 0, 0, 0, 0, codec.EmptyAddress, nil
	}
	if err != nil {
		return false, nil, nil, nil, nil, nil, nil, nil, false, false, ids.Empty, 0, 0, 0, 0, 0, codec.EmptyAddress, err
	}

	offset := uint16(0)
	nameLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	name := v[offset : offset+nameLen]
	offset += nameLen
	descriptionLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	description := v[offset : offset+descriptionLen]
	offset += descriptionLen
	categoriesLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	categories := v[offset : offset+categoriesLen]
	offset += categoriesLen
	licenseNameLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	licenseName := v[offset : offset+licenseNameLen]
	offset += licenseNameLen
	licenseSymbolLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	licenseSymbol := v[offset : offset+licenseSymbolLen]
	offset += licenseSymbolLen
	licenseURLLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	licenseURL := v[offset : offset+licenseURLLen]
	offset += licenseURLLen
	metadataLen := binary.BigEndian.Uint16(v[offset : offset+consts.Uint16Len])
	offset += consts.Uint16Len
	metadata := v[offset : offset+metadataLen]
	offset += metadataLen

	isCommunityDataset := v[offset] == successByte
	offset += consts.BoolLen
	onSale := v[offset] == successByte
	offset += consts.BoolLen

	baseAsset, err := ids.ToID(v[offset : offset+ids.IDLen])
	if err != nil {
		return false, nil, nil, nil, nil, nil, nil, nil, false, false, ids.Empty, 0, 0, 0, 0, 0, codec.EmptyAddress, err
	}
	offset += ids.IDLen

	basePrice := binary.BigEndian.Uint64(v[offset : offset+consts.Uint64Len])
	offset += consts.Uint64Len
	revenueModelDataShare := v[offset]
	offset += consts.Uint8Len
	revenueModelMetadataShare := v[offset]
	offset += consts.Uint8Len
	revenueModelDataOwnerCut := v[offset]
	offset += consts.Uint8Len
	revenueModelMetadataOwnerCut := v[offset]
	offset += consts.Uint8Len

	var owner codec.Address
	copy(owner[:], v[offset:])
	return true, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, nil
}

func SetAssetDataset(ctx context.Context, mu state.Mutable, datasetID ids.ID, name []byte, description []byte, categories []byte, licenseName []byte, licenseSymbol []byte, licenseURL []byte, metadata []byte, isCommunityDataset bool, onSale bool, baseAsset ids.ID, basePrice uint64, revenueModelDataShare uint8, revenueModelMetadataShare uint8, revenueModeldataOwnerCut uint8, revenueModelMetadataOwnerCut uint8, owner codec.Address) error {
	k := AssetDatasetKey(datasetID)
	nameLen := len(name)
	descriptionLen := len(description)
	categoriesLen := len(categories)
	licenseNameLen := len(licenseName)
	licenseSymbolLen := len(licenseSymbol)
	licenseURLLen := len(licenseURL)
	metadataLen := len(metadata)

	v := make([]byte, consts.Uint16Len+nameLen+consts.Uint16Len+descriptionLen+consts.Uint16Len+categoriesLen+consts.Uint16Len+licenseNameLen+consts.Uint16Len+licenseSymbolLen+consts.Uint16Len+licenseURLLen+consts.Uint16Len+metadataLen+consts.BoolLen+consts.BoolLen+ids.IDLen+consts.Uint64Len+consts.Uint8Len+consts.Uint8Len+consts.Uint8Len+consts.Uint8Len+codec.AddressLen)

	offset := 0
	binary.BigEndian.PutUint16(v[offset:], uint16(nameLen))
	offset += consts.Uint16Len
	copy(v[offset:], name)
	offset += nameLen
	binary.BigEndian.PutUint16(v[offset:], uint16(descriptionLen))
	offset += consts.Uint16Len
	copy(v[offset:], description)
	offset += descriptionLen
	binary.BigEndian.PutUint16(v[offset:], uint16(categoriesLen))
	offset += consts.Uint16Len
	copy(v[offset:], categories)
	offset += categoriesLen
	binary.BigEndian.PutUint16(v[offset:], uint16(licenseNameLen))
	offset += consts.Uint16Len
	copy(v[offset:], licenseName)
	offset += licenseNameLen
	binary.BigEndian.PutUint16(v[offset:], uint16(licenseSymbolLen))
	offset += consts.Uint16Len
	copy(v[offset:], licenseSymbol)
	offset += licenseSymbolLen
	binary.BigEndian.PutUint16(v[offset:], uint16(licenseURLLen))
	offset += consts.Uint16Len
	copy(v[offset:], licenseURL)
	offset += licenseURLLen
	binary.BigEndian.PutUint16(v[offset:], uint16(metadataLen))
	offset += consts.Uint16Len
	copy(v[offset:], metadata)
	offset += metadataLen

	if isCommunityDataset {
		v[offset] = successByte
	} else {
		v[offset] = failureByte
	}
	offset += consts.BoolLen
	if onSale {
		v[offset] = successByte
	} else {
		v[offset] = failureByte
	}
	offset += consts.BoolLen

	copy(v[offset:], baseAsset[:])
	offset += ids.IDLen

	binary.BigEndian.PutUint64(v[offset:], basePrice)
	offset += consts.Uint64Len
	v[offset] = revenueModelDataShare
	offset += consts.Uint8Len
	v[offset] = revenueModelMetadataShare
	offset += consts.Uint8Len
	v[offset] = revenueModeldataOwnerCut
	offset += consts.Uint8Len
	v[offset] = revenueModelMetadataOwnerCut
	offset += consts.Uint8Len

	copy(v[offset:], owner[:])

	return mu.Insert(ctx, k, v)
}

func DeleteAssetDataset(ctx context.Context, mu state.Mutable, datasetID ids.ID) error {
	k := AssetDatasetKey(datasetID)
	return mu.Remove(ctx, k)
}

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
