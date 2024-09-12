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
