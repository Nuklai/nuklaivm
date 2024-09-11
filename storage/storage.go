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

	assetPrefix                    = 0x6
	assetNFTPrefix                 = 0x7
	assetCollectionPagePrefix      = 0x8
	assetCollectionPageCountPrefix = 0x9

	registerValidatorStakePrefix = 0xA
	delegateUserStakePrefix      = 0xB

	datasetPrefix = 0xC
)

const (
	BalanceChunks                  uint16 = 1
	AssetChunks                    uint16 = 16
	AssetNFTChunks                 uint16 = 3
	AssetCollectionPageChunks      uint16 = 11
	AssetCollectionPageCountChunks uint16 = 1
	AssetDatasetChunks             uint16 = 3
	RegisterValidatorStakeChunks   uint16 = 4
	DelegateUserStakeChunks        uint16 = 2
	DatasetChunks                  uint16 = 101
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

	var admin codec.Address
	copy(admin[:], v[offset:])
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

	return true, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, nil
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
	admin codec.Address,
	mintActor codec.Address,
	pauseUnpauseActor codec.Address,
	freezeUnfreezeActor codec.Address,
	enableDisableKYCAccountActor codec.Address,
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
	copy(v[offset:], admin[:])
	offset += codec.AddressLen
	copy(v[offset:], mintActor[:])
	offset += codec.AddressLen
	copy(v[offset:], pauseUnpauseActor[:])
	offset += codec.AddressLen
	copy(v[offset:], freezeUnfreezeActor[:])
	offset += codec.AddressLen
	copy(v[offset:], enableDisableKYCAccountActor[:])

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

// [collectionID + pageIndex]
func AssetCollectionPageKey(collectionID ids.ID, pageIndex uint64) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint64Len+consts.Uint16Len)                         // Length of prefix + collectionID + pageIndex + AssetCollectionPageChunks
	k[0] = assetCollectionPagePrefix                                                        // assetCollectionPagePrefix is a constant representing a page in a collection
	copy(k[1:], collectionID[:])                                                            // Copy the collectionID
	binary.BigEndian.PutUint64(k[1+ids.IDLen:], pageIndex)                                  // Adding pageIndex for dynamic growth
	binary.BigEndian.PutUint16(k[1+ids.IDLen+consts.Uint64Len:], AssetCollectionPageChunks) // Adding AssetCollectionPageChunks
	return
}

// [collectionID]
func AssetCollectionPageCountKey(collectionID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len)                              // Length of prefix + collectionID + AssetCollectionPageCountChunks
	k[0] = assetCollectionPageCountPrefix                                       // assetCollectionPageCountPrefix is a constant representing the count of pages in a collection
	copy(k[1:], collectionID[:])                                                // Copy the collectionID
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], AssetCollectionPageCountChunks) // Adding AssetCollectionPageCountChunks
	return
}

func GetAssetNFTsByCollectionFromState(
	ctx context.Context,
	f ReadState,
	collectionID ids.ID,
) ([]ids.ID, error) {
	// Retrieve the page count key to determine how many pages are in the collection.
	pageCountKey := AssetCollectionPageCountKey(collectionID)
	pageCountValues, pageCountErrs := f(ctx, [][]byte{pageCountKey})

	// Handle errors for retrieving the page count.
	if len(pageCountErrs) > 0 {
		if err := handleGetCollectionNFTError(pageCountErrs[0]); err != nil {
			return nil, err
		}
	}

	// Use the helper function to get all NFTs.
	return innerGetAssetNFTsByCollection(pageCountValues[0], func(ctx context.Context, key []byte) ([]byte, error) {
		values, errs := f(ctx, [][]byte{key})
		if len(errs) > 0 {
			if err := handleGetCollectionNFTError(errs[0]); err != nil {
				return nil, err
			}
		}
		return values[0], nil
	}, ctx, collectionID)
}

func GetAssetNFTsByCollection(
	ctx context.Context,
	im state.Immutable,
	collectionID ids.ID,
) ([]ids.ID, error) {
	// Retrieve the page count key to determine how many pages are in the collection.
	pageCountKey := AssetCollectionPageCountKey(collectionID)
	pageCountValue, err := im.GetValue(ctx, pageCountKey)
	if err := handleGetCollectionNFTError(err); err != nil {
		return nil, err
	}

	// Use the helper function to get all NFTs.
	return innerGetAssetNFTsByCollection(pageCountValue, im.GetValue, ctx, collectionID)
}

func innerGetAssetNFTsByCollection(
	pageCountValue []byte,
	getValueFunc func(ctx context.Context, key []byte) ([]byte, error),
	ctx context.Context,
	collectionID ids.ID,
) ([]ids.ID, error) {
	// Decode the page count.
	pageCount, err := decodeNFTCountInCollection(pageCountValue)
	if err != nil {
		return nil, err
	}

	// Initialize a slice to hold all NFT IDs.
	allNFTs := []ids.ID{}

	// Iterate over all pages to fetch each NFT ID.
	for i := uint64(0); i <= pageCount; i++ {
		pageKey := AssetCollectionPageKey(collectionID, i)
		pageValue, err := getValueFunc(ctx, pageKey)
		if err := handleGetCollectionNFTError(err); err != nil {
			return nil, err
		}
		if len(pageValue) == 0 {
			continue // No value found; skip to the next.
		}

		// Decode the list of NFTs from the current page.
		nftList := decodeNFTList(pageValue)
		allNFTs = append(allNFTs, nftList...)
	}

	return allNFTs, nil
}

func AddAssetNFT(ctx context.Context, mu state.Mutable, collectionID ids.ID, nftID ids.ID) error {
	// Fetch the current page count of NFTs in the collection.
	pageCountKey := AssetCollectionPageCountKey(collectionID)
	currentPageCountBytes, err := mu.GetValue(ctx, pageCountKey)
	if err := handleGetCollectionNFTError(err); err != nil {
		return err
	}
	currentPageCount, _ := decodeNFTCountInCollection(currentPageCountBytes)

	// Load the current page of NFTs.
	pageKey := AssetCollectionPageKey(collectionID, currentPageCount)
	currentPage, err := mu.GetValue(ctx, pageKey)
	if err := handleGetCollectionNFTError(err); err != nil {
		return err
	}

	// Determine if the current page has space.
	const maxNFTsPerPage = 10 // Define a maximum number of NFTs per page.
	var nftList []ids.ID
	if len(currentPage) > 0 {
		nftList = decodeNFTList(currentPage) // Decode the list of NFTs from the current page.
	}

	if len(nftList) >= maxNFTsPerPage {
		// If the current page is full, increment the page count and create a new page.
		currentPageCount++
		pageKey = AssetCollectionPageKey(collectionID, currentPageCount)
		nftList = []ids.ID{} // Start a new list for the new page.
	}

	// Add the new NFT to the list.
	nftList = append(nftList, nftID)

	// Encode the updated list and store it in the state.
	updatedPageData := encodeNFTList(nftList)
	if err := mu.Insert(ctx, pageKey, updatedPageData); err != nil {
		return err
	}

	// Update the page count if necessary.
	if len(nftList) == 1 { // Only update the page count if a new page was created.
		newPageCountBytes := make([]byte, consts.Uint64Len)
		binary.BigEndian.PutUint64(newPageCountBytes, currentPageCount)
		if err := mu.Insert(ctx, pageCountKey, newPageCountBytes); err != nil {
			return err
		}
	}

	return nil
}

func encodeNFTList(nftList []ids.ID) []byte {
	// Each ID is a fixed size. Calculate the total size of the byte array.
	idSize := ids.IDLen
	totalSize := len(nftList) * idSize

	// Create a byte array to hold all IDs.
	encoded := make([]byte, totalSize)

	// Copy each ID into the byte array.
	for i, nftID := range nftList {
		start := i * idSize
		end := start + idSize
		copy(encoded[start:end], nftID[:])
	}

	return encoded
}

func decodeNFTList(data []byte) []ids.ID {
	idSize := ids.IDLen
	totalNFTs := len(data) / idSize

	// Create a slice to hold the decoded IDs.
	nftList := make([]ids.ID, totalNFTs)

	// Extract each ID from the byte array.
	for i := 0; i < totalNFTs; i++ {
		start := i * idSize
		end := start + idSize
		copy(nftList[i][:], data[start:end])
	}

	return nftList
}

func handleGetCollectionNFTError(err error) error {
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil // No action needed for not found.
		}
		return err // Return any other error.
	}
	return nil
}

func decodeNFTCountInCollection(countValue []byte) (uint64, error) {
	if len(countValue) == 0 {
		return 0, nil // No NFTs in the collection.
	}
	return binary.BigEndian.Uint64(countValue), nil
}

func DeleteAssetCollectionNFT(
	ctx context.Context,
	mu state.Mutable,
	collectionID ids.ID,
) error {
	// Retrieve the page count key to determine how many pages are in the collection.
	pageCountKey := AssetCollectionPageCountKey(collectionID)
	pageCountValue, err := mu.GetValue(ctx, pageCountKey)
	if err := handleGetCollectionNFTError(err); err != nil {
		return err
	}

	// Decode the page count.
	pageCount, err := decodeNFTCountInCollection(pageCountValue)
	if err != nil {
		return err
	}

	// Iterate over the indices to delete each page of NFTs.
	for i := uint64(0); i <= pageCount; i++ {
		pageKey := AssetCollectionPageKey(collectionID, i)
		if err := handleGetCollectionNFTError(mu.Remove(ctx, pageKey)); err != nil {
			return err
		}
	}

	// Finally, delete the page count key for the collection.
	return mu.Remove(ctx, pageCountKey)
}

// [datasetID]
func DatasetKey(datasetID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len) // Length of prefix + datasetID + DatasetChunks
	k[0] = datasetPrefix                           // datasetPrefix is a constant representing the dataset category
	copy(k[1:], datasetID[:])                      // Copy the datasetID
	binary.BigEndian.PutUint16(k[1+ids.IDLen:], DatasetChunks)
	return
}

// Used to serve RPC queries
func GetDatasetFromState(
	ctx context.Context,
	f ReadState,
	datasetID ids.ID,
) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	values, errs := f(ctx, [][]byte{DatasetKey(datasetID)})
	return innerGetDataset(values[0], errs[0])
}

func GetDataset(
	ctx context.Context,
	im state.Immutable,
	datasetID ids.ID,
) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	k := DatasetKey(datasetID)
	return innerGetDataset(im.GetValue(ctx, k))
}

func innerGetDataset(v []byte, err error) (bool, []byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, bool, ids.ID, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
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

func SetDataset(ctx context.Context, mu state.Mutable, datasetID ids.ID, name []byte, description []byte, categories []byte, licenseName []byte, licenseSymbol []byte, licenseURL []byte, metadata []byte, isCommunityDataset bool, onSale bool, baseAsset ids.ID, basePrice uint64, revenueModelDataShare uint8, revenueModelMetadataShare uint8, revenueModeldataOwnerCut uint8, revenueModelMetadataOwnerCut uint8, owner codec.Address) error {
	k := DatasetKey(datasetID)
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

func DeleteDataset(ctx context.Context, mu state.Mutable, datasetID ids.ID) error {
	k := DatasetKey(datasetID)
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
