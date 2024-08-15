// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/x/merkledb"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/vm"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
)

var _ vm.Genesis = (*Genesis)(nil)

type CustomAllocation struct {
	Address string `json:"address"` // bech32 address
	Balance uint64 `json:"balance"`
}

type EmissionBalancer struct {
	MaxSupply       uint64 `json:"maxSupply"`       // Max supply of NAI
	EmissionAddress string `json:"emissionAddress"` // Emission address
}

type Genesis struct {
	// State Parameters
	StateBranchFactor merkledb.BranchFactor `json:"stateBranchFactor"`

	// Chain Parameters
	MinBlockGap      int64 `json:"minBlockGap"`      // ms
	MinEmptyBlockGap int64 `json:"minEmptyBlockGap"` // ms

	// Chain Fee Parameters
	MinUnitPrice               fees.Dimensions `json:"minUnitPrice"`
	UnitPriceChangeDenominator fees.Dimensions `json:"unitPriceChangeDenominator"`
	WindowTargetUnits          fees.Dimensions `json:"windowTargetUnits"` // 10s
	MaxBlockUnits              fees.Dimensions `json:"maxBlockUnits"`     // must be possible to reach before block too large

	// Tx Parameters
	ValidityWindow      int64 `json:"validityWindow"` // ms
	MaxActionsPerTx     uint8 `json:"maxActionsPerTx"`
	MaxOutputsPerAction uint8 `json:"maxOutputsPerAction"`

	// Tx Fee Parameters
	BaseComputeUnits          uint64 `json:"baseUnits"`
	StorageKeyReadUnits       uint64 `json:"storageKeyReadUnits"`
	StorageValueReadUnits     uint64 `json:"storageValueReadUnits"` // per chunk
	StorageKeyAllocateUnits   uint64 `json:"storageKeyAllocateUnits"`
	StorageValueAllocateUnits uint64 `json:"storageValueAllocateUnits"` // per chunk
	StorageKeyWriteUnits      uint64 `json:"storageKeyWriteUnits"`
	StorageValueWriteUnits    uint64 `json:"storageValueWriteUnits"` // per chunk

	// Allocates
	CustomAllocation []*CustomAllocation `json:"customAllocation"`

	// Emission Balancer Info
	EmissionBalancer EmissionBalancer `json:"emissionBalancer"`
}

func Default() *Genesis {
	return &Genesis{
		// State Parameters
		StateBranchFactor: merkledb.BranchFactor16,

		// Chain Parameters
		MinBlockGap:      100,
		MinEmptyBlockGap: 2_500,

		// Chain Fee Parameters
		MinUnitPrice:               fees.Dimensions{100, 100, 100, 100, 100},
		UnitPriceChangeDenominator: fees.Dimensions{48, 48, 48, 48, 48},
		WindowTargetUnits:          fees.Dimensions{20_000_000, 1_000, 1_000, 1_000, 1_000},
		MaxBlockUnits:              fees.Dimensions{1_800_000, 2_000, 2_000, 2_000, 2_000},

		// Tx Parameters
		ValidityWindow:      60 * consts.MillisecondsPerSecond, // ms
		MaxActionsPerTx:     16,
		MaxOutputsPerAction: 1,

		// Tx Fee Compute Parameters
		BaseComputeUnits: 1,

		// Tx Fee Storage Parameters
		//
		// TODO: tune this
		StorageKeyReadUnits:       5,
		StorageValueReadUnits:     2,
		StorageKeyAllocateUnits:   20,
		StorageValueAllocateUnits: 5,
		StorageKeyWriteUnits:      10,
		StorageValueWriteUnits:    3,

		EmissionBalancer: EmissionBalancer{
			MaxSupply:       emission.GetStakingConfig().RewardConfig.SupplyCap,       // 10 billion NAI,
			EmissionAddress: emission.GetStakingConfig().RewardConfig.EmissionAddress, // NAI emission address(If you don't pass this address, it will be set to the default address)
		},
	}
}

func New(b []byte, _ []byte /* upgradeBytes */) (*Genesis, error) {
	g := Default()
	if len(b) > 0 {
		if err := json.Unmarshal(b, g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config %s: %w", string(b), err)
		}
	}
	return g, nil
}

func (g *Genesis) Load(ctx context.Context, tracer trace.Tracer, mu state.Mutable) error {
	ctx, span := tracer.Start(ctx, "Genesis.Load")
	defer span.End()

	if err := g.StateBranchFactor.Valid(); err != nil {
		return err
	}

	totalSupply := uint64(0)
	for _, alloc := range g.CustomAllocation {
		pk, err := codec.ParseAddressBech32(nconsts.HRP, alloc.Address)
		if err != nil {
			return err
		}
		totalSupply, err = smath.Add64(totalSupply, alloc.Balance)
		if err != nil {
			return err
		}
		if err := storage.SetBalance(ctx, mu, pk, ids.Empty, alloc.Balance); err != nil {
			return fmt.Errorf("%w: addr=%s, bal=%d", err, alloc.Address, alloc.Balance)
		}
	}
	return storage.SetAsset(
		ctx,
		mu,
		ids.Empty,
		[]byte(nconsts.Name),
		[]byte(nconsts.Symbol),
		nconsts.Decimals,
		[]byte(nconsts.Name),
		totalSupply,
		g.EmissionBalancer.MaxSupply,
		codec.EmptyAddress,
		codec.EmptyAddress,
		codec.EmptyAddress,
		codec.EmptyAddress,
		codec.EmptyAddress,
		codec.EmptyAddress,
	)
}

func (g *Genesis) GetStateBranchFactor() merkledb.BranchFactor {
	return g.StateBranchFactor
}
