// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package externalsubscriber

import (
	"context"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/event"
	es "github.com/ava-labs/hypersdk/extension/externalsubscriber"
	"github.com/ava-labs/hypersdk/vm"
)

func OptionFunc(v *vm.VM, config es.Config) error {
	if !config.Enabled {
		return nil
	}
	server, err := NewExternalSubscriberClient(
		context.TODO(),
		v.Logger(),
		config.ServerAddress,
		v.GenesisBytes,
	)
	if err != nil {
		return err
	}

	blockSubscription := event.SubscriptionFuncFactory[*chain.ExecutedBlock]{
		AcceptF: func(blk *chain.ExecutedBlock) error {
			return server.Accept(blk)
		},
	}

	vm.WithBlockSubscriptions(blockSubscription)(v)
	return nil
}
