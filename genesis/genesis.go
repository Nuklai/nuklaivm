// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"context"
	"encoding/json"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/x/merkledb"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/genesis"
	"github.com/ava-labs/hypersdk/state"

	safemath "github.com/ava-labs/avalanchego/utils/math"
)

var (
	_ genesis.Genesis               = (*Genesis)(nil)
	_ genesis.GenesisAndRuleFactory = (*GenesisFactory)(nil)
)

type EmissionBalancer struct {
	MaxSupply       uint64 `json:"maxSupply"`       // Max supply of NAI
	EmissionAddress string `json:"emissionAddress"` // Emission address
}

type Genesis struct {
	*genesis.DefaultGenesis
	EmissionBalancer *EmissionBalancer `json:"emissionBalancer"`
}

func NewGenesis(customAllocations []*genesis.CustomAllocation, emissionBalancer EmissionBalancer) *Genesis {
	// Initialize the DefaultGenesis part using the NewDefaultGenesis function
	defaultGenesis := genesis.NewDefaultGenesis(customAllocations)

	// Return a new Genesis object, including EmissionBalancer initialization
	return &Genesis{
		DefaultGenesis:   defaultGenesis,
		EmissionBalancer: &emissionBalancer,
	}
}

func (g *Genesis) InitializeState(ctx context.Context, tracer trace.Tracer, mu state.Mutable, balanceHandler chain.BalanceHandler) error {
	_, span := tracer.Start(ctx, "Nuklai Genesis.InitializeState")
	defer span.End()

	// Initialize state from the DefaultGenesis first
	if err := g.DefaultGenesis.InitializeState(ctx, tracer, mu, balanceHandler); err != nil {
		return err
	}

	// Get totalSupply
	var (
		totalSupply uint64
		err         error
	)
	for _, alloc := range g.CustomAllocation {
		totalSupply, err = safemath.Add(totalSupply, alloc.Balance)
		if err != nil {
			return err
		}
	}

	// Set the asset info for NAI using storage.SetAsset
	return storage.SetAssetInfo(
		ctx,
		mu,
		storage.NAIAddress,           // Asset Address
		consts.AssetFungibleTokenID,  // Asset type ID
		[]byte(consts.Name),          // Name
		[]byte(consts.Symbol),        // Symbol
		consts.Decimals,              // Decimals
		[]byte(consts.Metadata),      // Metadata
		[]byte(consts.Metadata),      // Description
		totalSupply,                  // Initial supply
		g.EmissionBalancer.MaxSupply, // Max supply
		codec.EmptyAddress,           // Owner address
		codec.EmptyAddress,           // MintAdmin address
		codec.EmptyAddress,           // PauseUnpauseAdmin address
		codec.EmptyAddress,           // FreezeUnfreezeAdmin address
		codec.EmptyAddress,           // EnableDisableKYCAccountAdmin address
	)
}

func (g *Genesis) GetStateBranchFactor() merkledb.BranchFactor {
	return g.DefaultGenesis.GetStateBranchFactor()
}

type GenesisFactory struct {
	*genesis.DefaultGenesisFactory
}

// Update the Load function to return the proper type
func (GenesisFactory) Load(genesisBytes []byte, _ []byte, networkID uint32, chainID ids.ID) (genesis.Genesis, genesis.RuleFactory, error) {
	ngenesis := &Genesis{} // This is Nuklai's custom Genesis
	if err := json.Unmarshal(genesisBytes, ngenesis); err != nil {
		return nil, nil, err
	}
	ngenesis.Rules.NetworkID = networkID
	ngenesis.Rules.ChainID = chainID

	// Return the correct types: *Genesis and *ImmutableRuleFactory
	return ngenesis, &genesis.ImmutableRuleFactory{Rules: ngenesis.Rules}, nil
}
