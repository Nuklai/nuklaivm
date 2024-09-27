// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"encoding/json"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/vm"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/marketplace"

	safemath "github.com/ava-labs/avalanchego/utils/math"
)

const Namespace = "controller"

type Config struct {
	Enabled bool `json:"enabled"`
}

func NewDefaultConfig() Config {
	return Config{
		Enabled: true,
	}
}

func With() vm.Option {
	return vm.NewOption(Namespace, NewDefaultConfig(), func(v *vm.VM, config Config) error {
		if !config.Enabled {
			return nil
		}
		vm.WithVMAPIs(jsonRPCServerFactory{})(v)
		return nil
	})
}

func WithRuntime() vm.Option {
	return vm.NewOption(Namespace+"runtime", *runtime.NewConfig(), func(v *vm.VM, cfg runtime.Config) error {
		wasmRuntime = runtime.NewRuntime(&cfg, v.Logger())
		return nil
	})
}

func WithEmissionBalancer() vm.Option {
	return vm.NewOption(Namespace+"emissionBalancer", NewDefaultConfig(), func(v *vm.VM, config Config) error {
		if !config.Enabled {
			return nil
		}
		var ngenesis genesis.Genesis
		if err := json.Unmarshal(v.GenesisBytes, &ngenesis); err != nil {
			return err
		}
		// Get totalSupply
		totalSupply := uint64(0)
		for _, alloc := range ngenesis.CustomAllocation {
			supply, err := safemath.Add(totalSupply, alloc.Balance)
			if err != nil {
				return err
			}
			totalSupply = supply
		}
		emissionAddress, err := codec.StringToAddress(ngenesis.EmissionBalancer.EmissionAddress)
		if err != nil {
			return err
		}

		emissionBalancer = emission.NewEmission(v.Logger(), v, totalSupply, ngenesis.EmissionBalancer.MaxSupply, emissionAddress)
		return nil
	})
}

func WithNuklaiMarketplace() vm.Option {
	return vm.NewOption(Namespace+"marketplace", NewDefaultConfig(), func(v *vm.VM, config Config) error {
		if !config.Enabled {
			return nil
		}
		nuklaiMarketplace = marketplace.NewMarketplace(v.Logger(), v)
		return nil
	})
}
