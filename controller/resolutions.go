// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"
	"encoding/base64"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/emissionbalancer"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/storage"
)

func (c *Controller) Genesis() *genesis.Genesis {
	return c.genesis
}

func (c *Controller) Logger() logging.Logger {
	return c.inner.Logger()
}

func (c *Controller) Tracer() trace.Tracer {
	return c.inner.Tracer()
}

func (c *Controller) GetTransaction(
	ctx context.Context,
	txID ids.ID,
) (bool, int64, bool, chain.Dimensions, uint64, error) {
	return storage.GetTransaction(ctx, c.metaDB, txID)
}

func (c *Controller) GetBalanceFromState(
	ctx context.Context,
	acct codec.Address,
) (uint64, error) {
	return storage.GetBalanceFromState(ctx, c.inner.ReadState, acct)
}

func (c *Controller) GetEmissionBalancerInfo(ctx context.Context) (uint64, uint64, uint64, map[string]*emissionbalancer.Validator, error) {
	totalSupply := c.emissionBalancer.TotalSupply
	maxSupply := c.emissionBalancer.MaxSupply
	rewardsPerBlock := c.emissionBalancer.RewardsPerBlock

	validators := make(map[string]*emissionbalancer.Validator)
	currentValidators, _ := c.inner.CurrentValidators(ctx)
	for nodeId, validator := range currentValidators {
		nodeIdString := nodeId.String()
		newValidator := &emissionbalancer.Validator{
			NodeID:        nodeIdString,
			NodePublicKey: base64.StdEncoding.EncodeToString(validator.PublicKey.Compress()),
			NodeWeight:    validator.Weight,
			WalletAddress: "",
			StakedAmount:  0,
			StakedReward:  0,
		}
		validators[nodeIdString] = newValidator
	}

	return totalSupply, maxSupply, rewardsPerBlock, validators, nil
}
