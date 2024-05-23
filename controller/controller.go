// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"
	"fmt"
	"net/http"

	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/hypersdk/builder"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/gossiper"
	hrpc "github.com/ava-labs/hypersdk/rpc"
	hstorage "github.com/ava-labs/hypersdk/storage"
	"github.com/ava-labs/hypersdk/vm"
	"go.uber.org/zap"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	"github.com/nuklai/nuklaivm/config"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	nrpc "github.com/nuklai/nuklaivm/rpc"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/version"
)

var _ vm.Controller = (*Controller)(nil)

type Controller struct {
	inner *vm.VM

	snowCtx      *snow.Context
	genesis      *genesis.Genesis
	config       *config.Config
	stateManager *StateManager

	metrics *metrics

	metaDB database.Database

	emission *emission.Emission // Emission Balancer for NuklaiVM
}

func New() *vm.VM {
	return vm.New(&Controller{}, version.Version)
}

func (c *Controller) Initialize(
	inner *vm.VM,
	snowCtx *snow.Context,
	gatherer ametrics.MultiGatherer,
	genesisBytes []byte,
	upgradeBytes []byte, // subnets to allow for AWM
	configBytes []byte,
) (
	vm.Config,
	vm.Genesis,
	builder.Builder,
	gossiper.Gossiper,
	database.Database,
	database.Database,
	vm.Handlers,
	chain.ActionRegistry,
	chain.AuthRegistry,
	map[uint8]vm.AuthEngine,
	error,
) {
	c.inner = inner
	c.snowCtx = snowCtx
	c.stateManager = &StateManager{}

	// Instantiate metrics
	var err error
	c.metrics, err = newMetrics(gatherer)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	// Load config and genesis
	c.config, err = config.New(c.snowCtx.NodeID, configBytes)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	c.snowCtx.Log.SetLevel(c.config.GetLogLevel())
	snowCtx.Log.Info("initialized config", zap.Bool("loaded", c.config.Loaded()), zap.Any("contents", c.config))

	c.genesis, err = genesis.New(genesisBytes, upgradeBytes)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf(
			"unable to read genesis: %w",
			err,
		)
	}
	snowCtx.Log.Info("loaded genesis", zap.Any("genesis", c.genesis))

	// Create DBs
	blockDB, stateDB, metaDB, err := hstorage.New(snowCtx.ChainDataDir, gatherer)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	c.metaDB = metaDB

	// Create handlers
	//
	// hypersdk handler are initiatlized automatically, you just need to
	// initialize custom handlers here.
	apis := map[string]http.Handler{}
	jsonRPCHandler, err := hrpc.NewJSONRPCHandler(
		nconsts.Name,
		nrpc.NewJSONRPCServer(c),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	apis[nrpc.JSONRPCEndpoint] = jsonRPCHandler

	// Create builder and gossiper
	var (
		build  builder.Builder
		gossip gossiper.Gossiper
	)
	if c.config.TestMode {
		c.inner.Logger().Info("running build and gossip in test mode")
		build = builder.NewManual(inner)
		gossip = gossiper.NewManual(inner)
	} else {
		build = builder.NewTime(inner)
		gcfg := gossiper.DefaultProposerConfig()
		gcfg.GossipMaxSize = c.config.GossipMaxSize
		gcfg.GossipProposerDiff = c.config.GossipProposerDiff
		gcfg.GossipProposerDepth = c.config.GossipProposerDepth
		gcfg.NoGossipBuilderDiff = c.config.NoGossipBuilderDiff
		gcfg.VerifyTimeout = c.config.VerifyTimeout
		gossip, err = gossiper.NewProposer(inner, gcfg)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	// Initialize emission balancer
	emissionAddr, err := codec.ParseAddressBech32(nconsts.HRP, c.genesis.EmissionBalancer.EmissionAddress)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	// Get the total supply from the custom allocations
	totalSupply := uint64(0)
	for _, alloc := range c.genesis.CustomAllocation {
		totalSupply += alloc.Balance
	}
	c.emission = emission.New(c, c.inner, totalSupply, c.genesis.EmissionBalancer.MaxSupply, emissionAddr)

	return c.config, c.genesis, build, gossip, blockDB, stateDB, apis, nconsts.ActionRegistry, nconsts.AuthRegistry, auth.Engines(), nil
}

func (c *Controller) Rules(t int64) chain.Rules {
	// TODO: extend with [UpgradeBytes]
	return c.genesis.Rules(t, c.snowCtx.NetworkID, c.snowCtx.ChainID)
}

func (c *Controller) StateManager() chain.StateManager {
	return c.stateManager
}

func (c *Controller) Accepted(ctx context.Context, blk *chain.StatelessBlock) error {
	batch := c.metaDB.NewBatch()
	defer batch.Reset()

	totalFee := uint64(0)
	results := blk.Results()
	for i, tx := range blk.Txs {
		result := results[i]
		if c.config.GetStoreTransactions() {
			err := storage.StoreTransaction(
				ctx,
				batch,
				tx.ID(),
				blk.GetTimestamp(),
				result.Success,
				result.Consumed,
				result.Fee,
			)
			if err != nil {
				return err
			}
		}
		totalFee += result.Fee
		if result.Success {
			switch action := tx.Action.(type) {
			case *actions.Transfer:
				c.metrics.transfer.Inc()
			case *actions.CreateAsset:
				c.metrics.createAsset.Inc()
			case *actions.MintAsset:
				c.metrics.mintAsset.Inc()
			case *actions.BurnAsset:
				c.metrics.burnAsset.Inc()
			case *actions.ImportAsset:
				c.metrics.importAsset.Inc()
			case *actions.ExportAsset:
				c.metrics.exportAsset.Inc()
			case *actions.RegisterValidatorStake:
				stakeInfo, err := actions.UnmarshalValidatorStakeInfo(action.StakeInfo)
				if err != nil {
					// This should never happen
					return err
				}
				c.metrics.validatorStakeAmount.Add(float64(stakeInfo.StakedAmount))
				c.metrics.registerValidatorStake.Inc()
			case *actions.ClaimValidatorStakeRewards:
				rewardResult, err := actions.UnmarshalClaimRewardsResult(result.Output)
				if err != nil {
					// This should never happen
					return err
				}
				c.metrics.mintedNAI.Add(float64(rewardResult.RewardAmount))
				c.metrics.rewardAmount.Add(float64(rewardResult.RewardAmount))
				c.metrics.claimStakingRewards.Inc()
			case *actions.WithdrawValidatorStake:
				stakeResult, err := actions.UnmarshalWithdrawValidatorStakeResult(result.Output)
				if err != nil {
					// This should never happen
					return err
				}
				c.metrics.validatorStakeAmount.Sub(float64(stakeResult.StakedAmount))
				c.metrics.mintedNAI.Add(float64(stakeResult.RewardAmount))
				c.metrics.rewardAmount.Add(float64(stakeResult.RewardAmount))
				c.metrics.claimStakingRewards.Inc()
				c.metrics.withdrawValidatorStake.Inc()
			case *actions.DelegateUserStake:
				c.metrics.delegatorStakeAmount.Add(float64(action.StakedAmount))
				c.metrics.delegateUserStake.Inc()
			case *actions.ClaimDelegationStakeRewards:
				rewardResult, err := actions.UnmarshalClaimRewardsResult(result.Output)
				if err != nil {
					// This should never happen
					return err
				}
				c.metrics.mintedNAI.Add(float64(rewardResult.RewardAmount))
				c.metrics.rewardAmount.Add(float64(rewardResult.RewardAmount))
				c.metrics.claimStakingRewards.Inc()
			case *actions.UndelegateUserStake:
				stakeResult, err := actions.UnmarshalUndelegateUserStakeResult(result.Output)
				if err != nil {
					// This should never happen
					return err
				}
				c.metrics.delegatorStakeAmount.Sub(float64(stakeResult.StakedAmount))
				c.metrics.mintedNAI.Add(float64(stakeResult.RewardAmount))
				c.metrics.rewardAmount.Add(float64(stakeResult.RewardAmount))
				c.metrics.claimStakingRewards.Inc()
				c.metrics.undelegateUserStake.Inc()
			}
		}
	}

	// Distribute fees
	if totalFee > 0 {
		c.emission.DistributeFees(totalFee)
		emissionAddress, err := codec.AddressBech32(nconsts.HRP, c.emission.EmissionAccount.Address)
		if err != nil {
			return err // This should never happen
		}
		c.inner.Logger().Info("distributed fees to Emission and Validators", zap.Uint64("current block height", c.inner.LastAcceptedBlock().Height()), zap.Uint64("total fee", totalFee), zap.Uint64("total supply", c.emission.TotalSupply), zap.Uint64("max supply", c.emission.MaxSupply), zap.Uint64("rewards per epock", c.emission.GetRewardsPerEpoch()), zap.String("emission address", emissionAddress), zap.Uint64("emission address accumulated reward", c.emission.EmissionAccount.AccumulatedReward))
		c.metrics.feesDistributed.Add(float64(totalFee))
	}

	// Mint new NAI if needed
	mintNewNAI := c.emission.MintNewNAI()
	if mintNewNAI > 0 {
		c.emission.AddToTotalSupply(mintNewNAI)
		c.inner.Logger().Info("minted new NAI", zap.Uint64("current block height", c.inner.LastAcceptedBlock().Height()), zap.Uint64("newly minted NAI", mintNewNAI), zap.Uint64("total supply", c.emission.TotalSupply), zap.Uint64("max supply", c.emission.MaxSupply))
	}

	return batch.Write()
}

func (*Controller) Rejected(context.Context, *chain.StatelessBlock) error {
	return nil
}

func (*Controller) Shutdown(context.Context) error {
	// Do not close any databases provided during initialization. The VM will
	// close any databases your provided.
	return nil
}
