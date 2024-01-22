// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"
	"fmt"
	"net/http"

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
	"github.com/ava-labs/hypersdk/state"
	hstorage "github.com/ava-labs/hypersdk/storage"
	"github.com/ava-labs/hypersdk/vm"
	"go.uber.org/zap"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	"github.com/nuklai/nuklaivm/config"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/rpc"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/version"
)

var _ vm.Controller = (*Controller)(nil)

type Controller struct {
	inner *vm.VM

	snowCtx      *snow.Context
	genesis      *genesis.Genesis
	config       *config.Config
	stateManager *storage.StateManager

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
	c.stateManager = &storage.StateManager{}

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
		consts.Name,
		rpc.NewJSONRPCServer(c),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	apis[rpc.JSONRPCEndpoint] = jsonRPCHandler

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
		gossip, err = gossiper.NewProposer(inner, gcfg)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	currentValidators := make(map[ids.NodeID]*validators.GetValidatorOutput)
	if !c.config.TestMode {
		// We only get the validators in non-test mode
		currentValidators, _ = inner.CurrentValidators(context.TODO())
	}
	// Initialize emission
	c.emission = emission.New(c, c.genesis.MaxSupply, c.genesis.RewardsPerBlock, currentValidators)

	return c.config, c.genesis, build, gossip, blockDB, stateDB, apis, consts.ActionRegistry, consts.AuthRegistry, auth.Engines(), nil
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

	// Retrieve the vm state
	stateDB, err := c.inner.State()
	if err != nil {
		return err
	}
	// Retrieve the state.Mutable to write to
	mu := state.NewSimpleMutable(stateDB)

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
		currentHeight := c.inner.LastAcceptedBlock().Height()
		if result.Success {
			switch action := tx.Action.(type) {
			case *actions.Transfer:
				c.metrics.transfer.Inc()
			case *actions.StakeValidator:
				c.metrics.stake.Inc()
				currentValidators, _ := c.inner.CurrentValidators(ctx)
				// Check to make sure the stake is valid
				if action.EndLockUp > currentHeight {
					if err := c.emission.StakeToValidator(tx.ID(), tx.Auth.Actor(), currentValidators, currentHeight, action); err != nil {
						c.inner.Logger().Error("failed to stake to validator", zap.Error(err))
						break
					}
				} else {
					c.inner.Logger().Error("failed to stake to validator", zap.Error(fmt.Errorf("start lockup %d is greater than end lockup %d", currentHeight, action.EndLockUp)))
				}
			case *actions.UnstakeValidator:
				c.metrics.unstake.Inc()
				// Check to make sure the unstake is valid
				_, _, stakedAmount, endLockUp, owner, _ := storage.GetStake(ctx, mu, action.Stake)
				if currentHeight > endLockUp {
					if err := c.emission.UnstakeFromValidator(owner, action); err != nil {
						c.inner.Logger().Error("failed to unstake from validator", zap.Error(err))
						// We exit early if it's an error that must never happen
						// Otherwise, we move on because while the stake may be  removed from Emission Balancer,
						// it may not have been removed from the blockchain state yet
						if err == emission.ErrInvalidNodeID {
							break
						}
					}
					// We exit early if the stake cannot be deleted from the state
					if err := storage.DeleteStake(ctx, mu, action.Stake); err != nil {
						c.inner.Logger().Error("failed to delete stake from blockchain state", zap.Error(err))
						break
					}
					// We exit early if the staked amount cannot be added to the user balance
					if err := storage.AddBalance(ctx, mu, owner, stakedAmount, true); err != nil {
						c.inner.Logger().Error("failed to add the staked amount to the user balance", zap.Error(err))
						break
					}
				}
			}
		}
	}

	// Calculate and distribute fees
	feeManager := blk.FeeManager()
	unitsConsumed := feeManager.UnitsConsumed()
	unitPrices := feeManager.UnitPrices()
	totalFee := uint64(0)
	for i := 0; i < len(unitsConsumed); i++ {
		totalFee += unitsConsumed[i] * unitPrices[i]
	}
	emissionAddr, err := codec.ParseAddressBech32(consts.HRP, c.genesis.EmissionAddress)
	if err != nil {
		return err
	}
	if err := c.emission.DistributeFees(ctx, mu, totalFee, emissionAddr); err != nil {
		return err
	}
	c.inner.Logger().Info("distributed fees to Emission and Validators", zap.Uint64("current block height", c.inner.LastAcceptedBlock().Height()), zap.Uint64("total fee", totalFee), zap.Uint64("total supply", c.emission.GetTotalSupply()), zap.Uint64("max supply", c.emission.GetMaxSupply()))

	// Mint new NAI if needed
	mintNewNAI, err := c.emission.MintNewNAI(ctx, mu, emissionAddr)
	if err != nil {
		return err
	}
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
