// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
)

const (
	DatasetInfoChunks             uint16 = 94
	DatasetContributionInfoChunks uint16 = 10
)

const ( 
	MaxDatasetTextSize = 256
	MaxDatasetMetadataSize = 5120
)

func DatasetInfoKey(datasetAddress codec.Address) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+consts.Uint16Len)                 // Length of prefix + datasetAddress + DatasetInfoChunks
	k[0] = datasetInfoPrefix                                              // datasetInfoPrefix is a constant representing the dataset category
	copy(k[1:1+codec.AddressLen], datasetAddress[:])                      // Copy the datasetAddress
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], DatasetInfoChunks) // Adding DatasetInfoChunks
	return
}

func SetDatasetInfo(ctx context.Context, mu state.Mutable, datasetAddress codec.Address, name []byte, description []byte, categories []byte, licenseName []byte, licenseSymbol []byte, licenseURL []byte, metadata []byte, isCommunityDataset bool, marketplaceAssetAddress codec.Address, baseAssetAddress codec.Address, basePrice uint64, revenueModelDataShare uint8, revenueModelMetadataShare uint8, revenueModeldataOwnerCut uint8, revenueModelMetadataOwnerCut uint8, owner codec.Address) error {
	// Setup
	k := DatasetInfoKey(datasetAddress)
	nameLen := len(name)
	descriptionLen := len(description)
	categoriesLen := len(categories)
	licenseNameLen := len(licenseName)
	licenseSymbolLen := len(licenseSymbol)
	licenseURLLen := len(licenseURL)
	metadataLen := len(metadata)
	datasetInfoSize := consts.Uint16Len + nameLen + consts.Uint16Len + descriptionLen + consts.Uint16Len + categoriesLen + consts.Uint16Len + licenseNameLen + consts.Uint16Len + licenseSymbolLen + consts.Uint16Len + licenseURLLen + consts.Uint16Len + metadataLen + consts.BoolLen + codec.AddressLen + codec.AddressLen + consts.Uint64Len + consts.Uint8Len + consts.Uint8Len + consts.Uint8Len + consts.Uint8Len + codec.AddressLen
	v := make([]byte, datasetInfoSize)

	// Populate
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

	copy(v[offset:], marketplaceAssetAddress[:])
	offset += codec.AddressLen
	copy(v[offset:], baseAssetAddress[:])
	offset += codec.AddressLen

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

// Used to serve RPC queries
func GetDatasetInfoFromState(
	ctx context.Context,
	f ReadState,
	dataset codec.Address,
) ([]byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, codec.Address, codec.Address, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	values, errs := f(ctx, [][]byte{DatasetInfoKey(dataset)})
	if errs[0] != nil {
		return nil, nil, nil, nil, nil, nil, nil, false, codec.EmptyAddress, codec.EmptyAddress, 0, 0, 0, 0, 0, codec.EmptyAddress, errs[0]
	}
	return innerGetDatasetInfo(values[0])
}

func GetDatasetInfoNoController(
	ctx context.Context,
	im state.Immutable,
	dataset codec.Address,
) ([]byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, codec.Address, codec.Address, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	k := DatasetInfoKey(dataset)
	v, err := im.GetValue(ctx, k)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, false, codec.EmptyAddress, codec.EmptyAddress, 0, 0, 0, 0, 0, codec.EmptyAddress, err
	}
	return innerGetDatasetInfo(v)
}

func innerGetDatasetInfo(v []byte) ([]byte, []byte, []byte, []byte, []byte, []byte, []byte, bool, codec.Address, codec.Address, uint64, uint8, uint8, uint8, uint8, codec.Address, error) {
	// Extract
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

	var marketplaceAssetAddress, baseAssetAddress codec.Address
	copy(marketplaceAssetAddress[:], v[offset:])
	offset += codec.AddressLen
	copy(baseAssetAddress[:], v[offset:])
	offset += codec.AddressLen

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

	return name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, marketplaceAssetAddress, baseAssetAddress, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, nil
}

func DeleteDatasetInfo(ctx context.Context, mu state.Mutable, datasetAddress codec.Address) error {
	k := DatasetInfoKey(datasetAddress)
	return mu.Remove(ctx, k)
}

func DatasetExists(
	ctx context.Context,
	mu state.Immutable,
	datasetAddress codec.Address,
) bool {
	v, err := mu.GetValue(ctx, DatasetInfoKey(datasetAddress))
	return v != nil && err == nil
}

func DatasetContributionID(datasetAddress codec.Address, dataLocation, dataIdentifer []byte, contributor codec.Address) ids.ID {
	v := make([]byte, codec.AddressLen+len(dataLocation)+len(dataIdentifer)+codec.AddressLen)
	offset := 0
	copy(v[offset:], datasetAddress[:])
	offset += codec.AddressLen
	copy(v[offset:], dataLocation)
	offset += len(dataLocation)
	copy(v[offset:], dataIdentifer)
	offset += len(dataIdentifer)
	copy(v[offset:], contributor[:])
	return utils.ToID(v)
}

func DatasetContributionInfoKey(contributionID ids.ID) (k []byte) {
	k = make([]byte, 1+ids.IDLen+consts.Uint16Len)                                    // Length of prefix + contributionID + MarketplaceContributionInfoChunks
	k[0] = marketplaceContributionPrefix                                              // marketplaceContributionPrefix is a constant representing the contribution category
	copy(k[1:1+ids.IDLen], contributionID[:])                                         // Copy the contributionID
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], DatasetContributionInfoChunks) // Adding MarketplaceContributionInfoChunks
	return
}

func SetDatasetContributionInfo(ctx context.Context, mu state.Mutable, contributionID ids.ID, datasetAddress codec.Address, dataLocation, dataIdentifier []byte, contributor codec.Address, active bool) error {
	// Setup
	k := DatasetContributionInfoKey(contributionID)
	dataLocationLen := len(dataLocation)
	dataIdentifierLen := len(dataIdentifier)
	contributionInfoSize := codec.AddressLen + consts.Uint16Len + dataLocationLen + consts.Uint16Len + dataIdentifierLen + codec.AddressLen + consts.BoolLen
	v := make([]byte, contributionInfoSize)

	// Populate
	offset := 0
	copy(v[offset:], datasetAddress[:])
	offset += codec.AddressLen
	binary.BigEndian.PutUint16(v[offset:], uint16(dataLocationLen))
	offset += consts.Uint16Len
	copy(v[offset:], dataLocation)
	offset += dataLocationLen
	binary.BigEndian.PutUint16(v[offset:], uint16(dataIdentifierLen))
	offset += consts.Uint16Len
	copy(v[offset:], dataIdentifier)
	offset += dataIdentifierLen
	copy(v[offset:], contributor[:])
	offset += codec.AddressLen
	if active {
		v[offset] = successByte
	} else {
		v[offset] = failureByte
	}

	return mu.Insert(ctx, k, v)
}

// Used to serve RPC queries
func GetDatasetContributionInfoFromState(ctx context.Context, f ReadState, contributionID ids.ID) (codec.Address, []byte, []byte, codec.Address, bool, error) {
	values, errs := f(ctx, [][]byte{DatasetContributionInfoKey(contributionID)})
	if errs[0] != nil {
		return codec.EmptyAddress, nil, nil, codec.EmptyAddress, false, errs[0]
	}
	return innerGetDatasetContributionInfo(values[0])
}

func GetDatasetContributionInfoNoController(
	ctx context.Context,
	im state.Immutable,
	contributionID ids.ID) (codec.Address, []byte, []byte, codec.Address, bool, error) {
	k := DatasetContributionInfoKey(contributionID)
	v, err := im.GetValue(ctx, k)
	if err != nil {
		return codec.EmptyAddress, nil, nil, codec.EmptyAddress, false, err
	}
	return innerGetDatasetContributionInfo(v)
}

func innerGetDatasetContributionInfo(v []byte) (codec.Address, []byte, []byte, codec.Address, bool, error) {
	// Extract
	offset := uint16(0)
	var datasetAddress codec.Address
	copy(datasetAddress[:], v[offset:])
	offset += codec.AddressLen
	dataLocationLen := binary.BigEndian.Uint16(v[offset:])
	offset += consts.Uint16Len
	dataLocation := v[offset : offset+dataLocationLen]
	offset += dataLocationLen
	dataIdentifierLen := binary.BigEndian.Uint16(v[offset:])
	offset += consts.Uint16Len
	dataIdentifier := v[offset : offset+dataIdentifierLen]
	offset += dataIdentifierLen
	var contributor codec.Address
	copy(contributor[:], v[offset:])
	offset += codec.AddressLen
	active := v[offset] == successByte

	return datasetAddress, dataLocation, dataIdentifier, contributor, active, nil
}

func DeleteDatasetContributionInfo(ctx context.Context, mu state.Mutable, contributionID ids.ID) error {
	k := DatasetContributionInfoKey(contributionID)
	return mu.Remove(ctx, k)
}

func DatasetContributionExists(
	ctx context.Context,
	mu state.Immutable,
	contributionID ids.ID,
) bool {
	v, err := mu.GetValue(ctx, DatasetContributionInfoKey(contributionID))
	return v != nil && err == nil
}
