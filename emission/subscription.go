// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/event"
	"go.uber.org/zap"
)

const (
	Name      = "emissionBalancer"
	Namespace = "emissionBalancer"
)

var (
	_ event.SubscriptionFactory[*chain.StatefulBlock] = (*EmissionSubscriptionFactory)(nil)
)

type EmissionSubscriptionFactory struct {
	log      logging.Logger
	emission Tracker
}

func (e *EmissionSubscriptionFactory) New() (event.Subscription[*chain.StatefulBlock], error) {
	return e, nil
}

func (e *EmissionSubscriptionFactory) Accept(blk *chain.StatefulBlock) error {
	totalFee := uint64(0)
	results := blk.Results()
	for j := range blk.Txs {
		result := results[j]
		totalFee += result.Fee
	}

	emissionAccount, totalSupply, maxSupply, totalStaked, _ := e.emission.GetInfo()
	// Distribute fees
	if totalFee > 0 {
		e.emission.DistributeFees(totalFee)
		e.log.Info("distributed fees to Emission and Validators", zap.Uint64("current block height", e.emission.GetLastAcceptedBlockHeight()), zap.Uint64("total fee", totalFee), zap.Uint64("total supply", totalSupply), zap.Uint64("max supply", maxSupply), zap.Uint64("total staked", totalStaked), zap.Uint64("rewards per epock", e.emission.GetRewardsPerEpoch()), zap.String("emission address", emissionAccount.Address.String()), zap.Uint64("emission address accumulated reward", emissionAccount.AccumulatedReward))
	}

	// Mint new NAI if needed
	mintNewNAI := e.emission.MintNewNAI()
	if mintNewNAI > 0 {
		e.emission.AddToTotalSupply(mintNewNAI)
		e.log.Info("minted new NAI", zap.Uint64("current block height", e.emission.GetLastAcceptedBlockHeight()), zap.Uint64("newly minted NAI", mintNewNAI), zap.Uint64("total supply", totalSupply+mintNewNAI), zap.Uint64("max supply", maxSupply))
	}

	return nil
}

func (e *EmissionSubscriptionFactory) Close() error {
	return nil
}

func NewEmissionSubscriptionFactory(log logging.Logger, emission Tracker) event.SubscriptionFactory[*chain.StatefulBlock] {
	return &EmissionSubscriptionFactory{
		log:      log,
		emission: emission,
	}
}
