// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
)

type Controller interface {
	Genesis() *genesis.Genesis
	Tracer() trace.Tracer
	GetTransaction(context.Context, ids.ID) (bool, int64, bool, fees.Dimensions, uint64, codec.Address, error)
	GetAssetFromState(context.Context, ids.ID) (bool, uint8, []byte, []byte, uint8, []byte, []byte, uint64, uint64, codec.Address, codec.Address, codec.Address, codec.Address, codec.Address, error)
	GetAssetNFTFromState(context.Context, ids.ID) (bool, ids.ID, uint64, []byte, []byte, codec.Address, error)

	GetBalanceFromState(context.Context, codec.Address, ids.ID) (uint64, error)

	GetEmissionInfo() (uint64, uint64, uint64, uint64, uint64, emission.EmissionAccount, emission.EpochTracker, error)
	GetValidators(ctx context.Context, staked bool) ([]*emission.Validator, error)
	GetStakedValidatorInfo(nodeID ids.NodeID) (*emission.Validator, error)
	GetValidatorStakeFromState(ctx context.Context, nodeID ids.NodeID) (
		bool, // exists
		uint64, // StakeStartBlock
		uint64, // StakeEndBlock
		uint64, // StakedAmount
		uint64, // DelegationFeeRate
		codec.Address, // RewardAddress
		codec.Address, // OwnerAddress
		error,
	)
	GetDelegatedUserStakeFromState(ctx context.Context, owner codec.Address, nodeID ids.NodeID) (
		bool, // exists
		uint64, // StakeStartBlock
		uint64, // StakeEndBlock
		uint64, // StakedAmount
		codec.Address, // RewardAddress
		codec.Address, // OwnerAddress
		error,
	)

	GetDatasetFromState(ctx context.Context, datasetID ids.ID) (
		bool, // exists
		[]byte, // name
		[]byte, // description
		[]byte, // categories
		[]byte, // licenseName
		[]byte, // licenseSymbol
		[]byte, // licenseURL
		[]byte, // metadata
		bool, // isCommunityDataset
		bool, // onSale
		ids.ID, // baseAsset
		uint64, // basePrice
		uint8, // revenueModelDataShare
		uint8, // revenueModelMetadataShare
		uint8, // revenueModelDataOwnerCut
		uint8, // revenueModelMetadataOwnerCut
		codec.Address, // Owner
		error,
	)
}
