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
	pageCount := decodeNFTCountInCollection(pageCountValue)

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
	currentPageCount := decodeNFTCountInCollection(currentPageCountBytes)

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

func decodeNFTCountInCollection(countValue []byte) uint64 {
	if len(countValue) == 0 {
		return 0 // No NFTs in the collection.
	}
	return binary.BigEndian.Uint64(countValue)
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
	pageCount := decodeNFTCountInCollection(pageCountValue)
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
