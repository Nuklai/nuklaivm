// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"github.com/nuklai/nuklaivm/config"
	"github.com/nuklai/nuklaivm/emission"

	"github.com/ava-labs/hypersdk/api/indexer"
	"github.com/ava-labs/hypersdk/extension/externalsubscriber"
	"github.com/ava-labs/hypersdk/vm"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"
)

const (
	Namespace      = "controller"
	configFilePath = "config.json" // Path to JSON config file
)

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

func WithIndexer(cfg config.Config) vm.Option {
	return vm.NewOption(indexer.Namespace, indexer.Config{
		Enabled:     true,
		BlockWindow: uint64(cfg.IndexerBlockWindow),
	}, indexer.OptionFunc)
}

func WithExternalSubscriber(cfg config.Config) vm.Option {
	if cfg.ExternalSubscriberAddr != "" {
		return vm.NewOption(externalsubscriber.Namespace, externalsubscriber.Config{
			Enabled:       true,
			ServerAddress: cfg.ExternalSubscriberAddr,
		}, externalsubscriber.OptionFunc)
	}
	return vm.NewOption(externalsubscriber.Namespace, externalsubscriber.Config{}, externalsubscriber.OptionFunc)
}

func WithRuntime() vm.Option {
	return vm.NewOption(Namespace+"runtime", *runtime.NewConfig(), func(v *vm.VM, cfg runtime.Config) error {
		wasmRuntime = runtime.NewRuntime(&cfg, v.Logger())
		return nil
	})
}

func WithEmissionBalancer() vm.Option {
	return vm.NewOption(Namespace+emission.Namespace, NewDefaultConfig(), func(v *vm.VM, config Config) error {
		if !config.Enabled {
			return nil
		}
		tracker, err := emission.NewEmission(v.Logger(), v)
		if err != nil {
			return err
		}
		emissionFactory := emission.NewEmissionSubscriptionFactory(v.Logger(), tracker)
		vm.WithBlockSubscriptions(emissionFactory)(v)
		emissionTracker = tracker
		return nil
	})
}
