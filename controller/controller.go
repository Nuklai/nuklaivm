// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/validators"
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

	emission *emission.Emission
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
	currentValidators := make(map[ids.NodeID]*validators.GetValidatorOutput)
	if !c.config.TestMode {
		// We only get the validators in non-test mode
		currentValidators, _ = inner.CurrentValidators(context.TODO())
	}
	emissionAddr, err := codec.ParseAddressBech32(nconsts.HRP, c.genesis.EmissionBalancer.EmissionAddress)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	c.emission = emission.New(c, c.genesis.EmissionBalancer.TotalSupply, c.genesis.EmissionBalancer.MaxSupply, c.genesis.EmissionBalancer.RewardsPerBlock, emissionAddr, currentValidators)

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
		currentHeight := c.inner.LastAcceptedBlock().Height()
		currentValidators, _ := c.inner.CurrentValidators(ctx)
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

				// Check if the tx actor has signing permission for this NodeID
				isValidatorOwner := false
				for _, validator := range currentValidators {
					signer := auth.NewBLSAddress(validator.PublicKey)
					if codec.MustAddressBech32(nconsts.HRP, tx.Auth.Actor()) == codec.MustAddressBech32(nconsts.HRP, signer) {
						isValidatorOwner = true
						break
					}
				}
				if !isValidatorOwner {
					c.inner.Logger().Error("failed to register validator stake", zap.Error(fmt.Errorf("actor %s is not the owner of the validator", codec.MustAddressBech32(nconsts.HRP, tx.Auth.Actor()))))
					continue
				}

				// Check to make sure the stake is valid
				currentTime := c.inner.LastAcceptedBlock().Timestamp().UTC()
				stakeStartTime := time.Unix(int64(stakeInfo.StakeStartTime), 0).UTC()
				if stakeStartTime.Before(currentTime) {
					// This should never happen as we check for this in register_validator_stake action
					c.inner.Logger().Error("failed to register validator stake", zap.Error(fmt.Errorf("stake start time %d is before current time %d", stakeInfo.StakeStartTime, currentTime.Unix())))
					return fmt.Errorf("stake start time %d is before current time %d", stakeInfo.StakeStartTime, currentTime.Unix())
				}
				// TODO: Register validator stake

				c.metrics.stakeAmount.Add(float64(stakeInfo.StakedAmount))
				c.metrics.registerValidatorStake.Inc()
			case *actions.StakeValidator:
				c.metrics.stake.Inc()
				// Check to make sure the stake is valid
				if action.EndLockUp > currentHeight {
					if err := c.emission.StakeToValidator(tx.ID(), tx.Auth.Actor(), currentValidators, currentHeight, action.NodeID, action.StakedAmount); err != nil {
						c.inner.Logger().Error("failed to stake to validator", zap.Error(err))
					}
				} else {
					c.inner.Logger().Error("failed to stake to validator", zap.Error(fmt.Errorf("start lockup %d is greater than end lockup %d", currentHeight, action.EndLockUp)))
				}
			case *actions.UnstakeValidator:
				c.metrics.unstake.Inc()
				stakeResult, err := actions.UnmarshalStakeResult(result.Output)
				if err != nil {
					// This should never happen
					return err
				}

				// Check to make sure the unstake is valid
				if currentHeight > stakeResult.EndLockUp {
					if err := c.emission.UnstakeFromValidator(tx.Auth.Actor(), action.NodeID, action.Stake); err != nil {
						c.inner.Logger().Error("failed to unstake from validator", zap.Error(err))
					}
				}
			}
		}
	}

	// Distribute fees
	if totalFee > 0 {
		c.emission.DistributeFees(totalFee)
		emissionAddress, err := codec.AddressBech32(nconsts.HRP, c.emission.GetEmissionAccount().Address)
		if err != nil {
			return err // This should never happen
		}
		c.inner.Logger().Info("distributed fees to Emission and Validators", zap.Uint64("current block height", c.inner.LastAcceptedBlock().Height()), zap.Uint64("total fee", totalFee), zap.Uint64("total supply", c.emission.GetTotalSupply()), zap.Uint64("max supply", c.emission.GetMaxSupply()), zap.Uint64("rewards per block", c.emission.GetRewardsPerBlock()), zap.String("emission address", emissionAddress), zap.Uint64("emission address balance", c.emission.GetEmissionAccount().Balance))
	}

	// Mint new NAI if needed
	mintNewNAI := c.emission.MintNewNAI()
	if mintNewNAI > 0 {
		c.emission.AddToTotalSupply(mintNewNAI)
		c.inner.Logger().Info("minted new NAI", zap.Uint64("current block height", c.inner.LastAcceptedBlock().Height()), zap.Uint64("newly minted NAI", mintNewNAI), zap.Uint64("total supply", c.emission.GetTotalSupply()), zap.Uint64("max supply", c.emission.GetMaxSupply()))
		c.metrics.mintNAI.Add(float64(mintNewNAI))
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
