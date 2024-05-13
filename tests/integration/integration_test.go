// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package integration_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/fatih/color"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/pubsub"
	hrpc "github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"
	"github.com/ava-labs/hypersdk/vm"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/controller"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

var (
	logFactory logging.Factory
	log        logging.Logger
)

func init() {
	logFactory = logging.NewFactory(logging.Config{
		DisplayLevel: logging.Debug,
	})
	l, err := logFactory.Make("main")
	if err != nil {
		panic(err)
	}
	log = l
}

func TestIntegration(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "nuklaivm integration test suites")
}

var (
	requestTimeout time.Duration
	vms            int
)

func init() {
	flag.DurationVar(
		&requestTimeout,
		"request-timeout",
		120*time.Second,
		"timeout for transaction issuance and confirmation",
	)
	flag.IntVar(
		&vms,
		"vms",
		5,
		"number of VMs to create",
	)
}

var (
	priv    ed25519.PrivateKey
	factory *auth.ED25519Factory
	rsender codec.Address
	sender  string

	priv2    ed25519.PrivateKey
	factory2 *auth.ED25519Factory
	rsender2 codec.Address
	sender2  string

	asset1         []byte
	asset1Symbol   []byte
	asset1Decimals uint8
	asset1ID       ids.ID
	asset2         []byte
	asset2Symbol   []byte
	asset2Decimals uint8
	asset2ID       ids.ID
	asset3         []byte
	asset3Symbol   []byte
	asset3Decimals uint8
	asset3ID       ids.ID

	// when used with embedded VMs
	genesisBytes []byte
	instances    []instance
	blocks       []snowman.Block

	networkID uint32
	gen       *genesis.Genesis
	app       *appSender

	withdraw0       string
	delegate        string
	rwithdraw0      codec.Address
	rdelegate       codec.Address
	delegateFactory *auth.ED25519Factory
	nodesFactories  []*auth.BLSFactory
	nodesAddresses  []codec.Address
	emissions       []emission.Tracker
	nodesPubKeys    []*bls.PublicKey
	height          int
)

type instance struct {
	chainID             ids.ID
	nodeID              ids.NodeID
	vm                  *vm.VM
	toEngine            chan common.Message
	JSONRPCServer       *httptest.Server
	NuklaiJSONRPCServer *httptest.Server
	WebSocketServer     *httptest.Server
	hcli                *hrpc.JSONRPCClient // clients for embedded VMs
	ncli                *nrpc.JSONRPCClient
}

var _ = ginkgo.BeforeSuite(func() {
	log.Info("VMID", zap.Stringer("id", nconsts.ID))
	gomega.Ω(vms).Should(gomega.BeNumerically(">", 1))

	var err error
	priv, err = ed25519.GeneratePrivateKey()
	gomega.Ω(err).Should(gomega.BeNil())
	factory = auth.NewED25519Factory(priv)
	rsender = auth.NewED25519Address(priv.PublicKey())
	sender = codec.MustAddressBech32(nconsts.HRP, rsender)
	log.Debug(
		"generated key",
		zap.String("addr", sender),
		zap.String("pk", hex.EncodeToString(priv[:])),
	)

	priv2, err = ed25519.GeneratePrivateKey()
	gomega.Ω(err).Should(gomega.BeNil())
	factory2 = auth.NewED25519Factory(priv2)
	rsender2 = auth.NewED25519Address(priv2.PublicKey())
	sender2 = codec.MustAddressBech32(nconsts.HRP, rsender2)
	log.Debug(
		"generated key",
		zap.String("addr", sender2),
		zap.String("pk", hex.EncodeToString(priv2[:])),
	)

	asset1 = []byte("1")
	asset1Symbol = []byte("s1")
	asset1Decimals = uint8(1)
	asset2 = []byte("2")
	asset2Symbol = []byte("s2")
	asset2Decimals = uint8(2)
	asset3 = []byte("3")
	asset3Symbol = []byte("s3")
	asset3Decimals = uint8(3)

	// create embedded VMs
	instances = make([]instance, vms)
	nodesFactories = make([]*auth.BLSFactory, vms)
	nodesAddresses = make([]codec.Address, vms)
	emissions = make([]emission.Tracker, vms)
	nodesPubKeys = make([]*bls.PublicKey, vms)

	gen = genesis.Default()
	gen.MinUnitPrice = chain.Dimensions{1, 1, 1, 1, 1}
	gen.MinBlockGap = 0
	gen.CustomAllocation = []*genesis.CustomAllocation{
		{
			Address: sender,
			Balance: 10_000_000_000_000_000,
		},
	}
	gen.EmissionBalancer = genesis.EmissionBalancer{
		TotalSupply:     10_000_000,
		MaxSupply:       10_000_000_000,
		EmissionAddress: sender,
	}
	genesisBytes, err = json.Marshal(gen)
	gomega.Ω(err).Should(gomega.BeNil())

	networkID = uint32(1)
	subnetID := ids.GenerateTestID()
	chainID := ids.GenerateTestID()

	app = &appSender{}
	for i := range instances {
		nodeID := ids.GenerateTestNodeID()
		sk, err := bls.NewSecretKey()
		gomega.Ω(err).Should(gomega.BeNil())
		l, err := logFactory.Make(nodeID.String())
		gomega.Ω(err).Should(gomega.BeNil())
		dname, err := os.MkdirTemp("", fmt.Sprintf("%s-chainData", nodeID.String()))
		gomega.Ω(err).Should(gomega.BeNil())
		snowCtx := &snow.Context{
			NetworkID:      networkID,
			SubnetID:       subnetID,
			ChainID:        chainID,
			NodeID:         nodeID,
			Log:            l,
			ChainDataDir:   dname,
			Metrics:        metrics.NewOptionalGatherer(),
			PublicKey:      bls.PublicFromSecretKey(sk),
			WarpSigner:     warp.NewSigner(sk, networkID, chainID),
			ValidatorState: &validators.TestState{},
		}
		nodesFactories[i] = auth.NewBLSFactory(sk)
		nodesAddresses[i] = auth.NewBLSAddress(snowCtx.PublicKey)
		nodesPubKeys[i] = snowCtx.PublicKey

		toEngine := make(chan common.Message, 1)
		db := memdb.New()

		v := controller.New()
		err = v.Initialize(
			context.TODO(),
			snowCtx,
			db,
			genesisBytes,
			nil,
			[]byte(
				`{"parallelism":3, "testMode":true, "logLevel":"debug", "trackedPairs":["*"]}`,
			),
			toEngine,
			nil,
			app,
		)
		gomega.Ω(err).Should(gomega.BeNil())

		var hd map[string]http.Handler
		hd, err = v.CreateHandlers(context.TODO())
		gomega.Ω(err).Should(gomega.BeNil())

		emissions[i] = emission.GetEmission()

		hjsonRPCServer := httptest.NewServer(hd[hrpc.JSONRPCEndpoint])
		njsonRPCServer := httptest.NewServer(hd[nrpc.JSONRPCEndpoint])
		webSocketServer := httptest.NewServer(hd[hrpc.WebSocketEndpoint])
		instances[i] = instance{
			chainID:             snowCtx.ChainID,
			nodeID:              snowCtx.NodeID,
			vm:                  v,
			toEngine:            toEngine,
			JSONRPCServer:       hjsonRPCServer,
			NuklaiJSONRPCServer: njsonRPCServer,
			WebSocketServer:     webSocketServer,
			hcli:                hrpc.NewJSONRPCClient(hjsonRPCServer.URL),
			ncli:                nrpc.NewJSONRPCClient(njsonRPCServer.URL, snowCtx.NetworkID, snowCtx.ChainID),
		}

		// Force sync ready (to mimic bootstrapping from genesis)
		v.ForceReady()
	}

	// Verify genesis allocates loaded correctly (do here otherwise test may
	// check during and it will be inaccurate)
	for _, inst := range instances {
		ncli := inst.ncli
		g, err := ncli.Genesis(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())

		csupply := uint64(0)
		for _, alloc := range g.CustomAllocation {
			balance, err := ncli.Balance(context.Background(), alloc.Address, ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(alloc.Balance))
			csupply += alloc.Balance
		}
		exists, symbol, decimals, metadata, supply, owner, warp, err := ncli.Asset(context.Background(), ids.Empty, false)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(exists).Should(gomega.BeTrue())
		gomega.Ω(string(symbol)).Should(gomega.Equal(nconsts.Symbol))
		gomega.Ω(decimals).Should(gomega.Equal(uint8(nconsts.Decimals)))
		gomega.Ω(string(metadata)).Should(gomega.Equal(nconsts.Name))
		gomega.Ω(supply).Should(gomega.Equal(csupply))
		gomega.Ω(owner).Should(gomega.Equal(codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress)))
		gomega.Ω(warp).Should(gomega.BeFalse())
	}
	blocks = []snowman.Block{}

	setEmissionValidators()

	app.instances = instances
	color.Blue("created %d VMs", vms)
})

var _ = ginkgo.AfterSuite(func() {
	for _, iv := range instances {
		iv.JSONRPCServer.Close()
		iv.NuklaiJSONRPCServer.Close()
		iv.WebSocketServer.Close()
		err := iv.vm.Shutdown(context.TODO())
		gomega.Ω(err).Should(gomega.BeNil())
	}
})

var _ = ginkgo.Describe("[Ping]", func() {
	ginkgo.It("can ping", func() {
		for _, inst := range instances {
			hcli := inst.hcli
			ok, err := hcli.Ping(context.Background())
			gomega.Ω(ok).Should(gomega.BeTrue())
			gomega.Ω(err).Should(gomega.BeNil())
		}
	})
})

var _ = ginkgo.Describe("[Network]", func() {
	ginkgo.It("can get network", func() {
		for _, inst := range instances {
			hcli := inst.hcli
			networkID, subnetID, chainID, err := hcli.Network(context.Background())
			gomega.Ω(networkID).Should(gomega.Equal(uint32(1)))
			gomega.Ω(subnetID).ShouldNot(gomega.Equal(ids.Empty))
			gomega.Ω(chainID).ShouldNot(gomega.Equal(ids.Empty))
			gomega.Ω(err).Should(gomega.BeNil())
		}
	})
})

var _ = ginkgo.Describe("[Nuklai staking mechanism]", func() {
	ginkgo.It("Setup and get initial staked validators", func() {
		height = 0
		withdraw0Priv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		rwithdraw0 = auth.NewED25519Address(withdraw0Priv.PublicKey())
		withdraw0 = codec.MustAddressBech32(nconsts.HRP, rwithdraw0)

		delegatePriv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		rdelegate = auth.NewED25519Address(delegatePriv.PublicKey())
		delegate = codec.MustAddressBech32(nconsts.HRP, rdelegate)
		delegateFactory = auth.NewED25519Factory(delegatePriv)

		validators, err := instances[3].ncli.StakedValidators(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(len(validators)).Should(gomega.Equal(0))
	})

	ginkgo.It("Funding node 3", func() {
		parser, err := instances[3].ncli.Parser(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		submit, _, _, err := instances[3].hcli.GenerateTransaction(
			context.Background(),
			parser,
			nil,
			&actions.Transfer{
				To:    nodesAddresses[3],
				Asset: ids.Empty,
				Value: 200_000_000_000,
			},
			factory,
		)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

		accept := expectBlk(instances[3])
		results := accept(true)
		gomega.Ω(results).Should(gomega.HaveLen(1))
		gomega.Ω(results[0].Success).Should(gomega.BeTrue())

		gomega.Ω(len(blocks)).Should(gomega.Equal(1))

		blk := blocks[height]
		ImportBlockToInstance(instances[0].vm, blk)
		ImportBlockToInstance(instances[4].vm, blk)
		ImportBlockToInstance(instances[2].vm, blk)
		ImportBlockToInstance(instances[1].vm, blk)
		height++

		balance, err := instances[3].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(balance).Should(gomega.Equal(uint64(200_000_000_000)))

		// check if gossip/ new state happens
		balanceOther, err := instances[4].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(balanceOther).Should(gomega.Equal(uint64(200_000_000_000)))

		balance, err = instances[3].ncli.Balance(context.TODO(), sender, ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(balance).Should(gomega.Equal(uint64(9_999_799_999_999_703)))
	})
	ginkgo.It("Register validator stake node 3", func() {
		parser, err := instances[3].ncli.Parser(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		currentBlockHeight := instances[3].vm.LastAcceptedBlock().Height()
		stakeStartBlock := currentBlockHeight + 2
		stakeEndBlock := currentBlockHeight + 100
		delegationFeeRate := 50

		stakeInfo := &actions.ValidatorStakeInfo{
			NodeID:            instances[3].nodeID.Bytes(),
			StakeStartBlock:   stakeStartBlock,
			StakeEndBlock:     stakeEndBlock,
			StakedAmount:      100_000_000_000,
			DelegationFeeRate: uint64(delegationFeeRate),
			RewardAddress:     rwithdraw0,
		}

		stakeInfoBytes, err := stakeInfo.Marshal()
		gomega.Ω(err).Should(gomega.BeNil())
		signature, err := nodesFactories[3].Sign(stakeInfoBytes)
		gomega.Ω(err).Should(gomega.BeNil())
		signaturePacker := codec.NewWriter(signature.Size(), signature.Size())
		signature.Marshal(signaturePacker)
		authSignature := signaturePacker.Bytes()
		submit, _, _, err := instances[3].hcli.GenerateTransaction(
			context.Background(),
			parser,
			nil,
			&actions.RegisterValidatorStake{
				StakeInfo:     stakeInfoBytes,
				AuthSignature: authSignature,
			},
			nodesFactories[3],
		)
		gomega.Ω(err).Should(gomega.BeNil())

		_, err = instances[3].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())

		gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

		accept := expectBlk(instances[3])
		results := accept(true)
		gomega.Ω(results).Should(gomega.HaveLen(1))
		gomega.Ω(results[0].Success).Should(gomega.BeTrue())

		gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

		blk := blocks[height]
		ImportBlockToInstance(instances[4].vm, blk)
		ImportBlockToInstance(instances[0].vm, blk)
		ImportBlockToInstance(instances[2].vm, blk)
		ImportBlockToInstance(instances[1].vm, blk)
		height++

		_, err = instances[3].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())

		// check if gossip/ new state happens
		_, err = instances[4].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())

		emissionInstance := emissions[3]
		currentValidators := emissionInstance.GetAllValidators(context.TODO())
		gomega.Ω(len(currentValidators)).To(gomega.Equal(5))
		stakedValidator := emissionInstance.GetStakedValidator(instances[3].nodeID)
		gomega.Ω(len(stakedValidator)).To(gomega.Equal(1))

		validator, exists := emissions[3].GetEmissionValidators()[instances[3].nodeID]
		gomega.Ω(exists).To(gomega.Equal(true))

		// check when it becomes active ?
		gomega.Ω(validator.IsActive).To(gomega.Equal(false))

		// test same block is accepted
		lastAcceptedBlock3, err := instances[3].vm.LastAccepted(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		lastAcceptedBlock4, err := instances[4].vm.LastAccepted(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(lastAcceptedBlock3).To(gomega.Equal(lastAcceptedBlock4))

		validators, err := instances[4].ncli.StakedValidators(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(len(validators)).Should(gomega.Equal(1))
	})

	ginkgo.It("Get validator staked amount after node 3 validator staking", func() {
		_, _, stakedAmount, _, _, _, err := instances[3].ncli.ValidatorStake(context.Background(), instances[3].nodeID)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(stakedAmount).Should(gomega.Equal(uint64(100_000_000_000)))
	})

	ginkgo.It("Get validator staked amount after staking using node 0 cli", func() {
		_, _, stakedAmount, _, _, _, err := instances[0].ncli.ValidatorStake(context.Background(), instances[3].nodeID)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(stakedAmount).Should(gomega.Equal(uint64(100_000_000_000)))
	})

	ginkgo.It("Get staked validators", func() {
		validators, err := instances[4].ncli.StakedValidators(context.TODO())
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(len(validators)).Should(gomega.Equal(1))
	})

	ginkgo.It("Transfer NAI to delegate user", func() {
		parser, err := instances[3].ncli.Parser(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		submit, _, _, err := instances[3].hcli.GenerateTransaction(
			context.Background(),
			parser,
			nil,
			&actions.Transfer{
				To:    rdelegate,
				Asset: ids.Empty,
				Value: 100_000_000_000,
			},
			factory,
		)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

		accept := expectBlk(instances[3])
		results := accept(true)
		gomega.Ω(results).Should(gomega.HaveLen(1))
		gomega.Ω(results[0].Success).Should(gomega.BeTrue())

		gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

		blk := blocks[height]
		ImportBlockToInstance(instances[0].vm, blk)
		ImportBlockToInstance(instances[4].vm, blk)
		ImportBlockToInstance(instances[2].vm, blk)
		ImportBlockToInstance(instances[1].vm, blk)
		height++

		balance, err := instances[0].ncli.Balance(context.TODO(), delegate, ids.Empty)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(balance).Should(gomega.Equal(uint64(100_000_000_000)))

		// test same block is accepted
		lastAcceptedBlock3, err := instances[3].vm.LastAccepted(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		lastAcceptedBlock4, err := instances[4].vm.LastAccepted(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(lastAcceptedBlock3).To(gomega.Equal(lastAcceptedBlock4))
	})

	ginkgo.It("Delegate user stake to node 3", func() {
		parser, err := instances[3].ncli.Parser(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		submit, _, _, err := instances[3].hcli.GenerateTransaction(
			context.Background(),
			parser,
			nil,
			&actions.DelegateUserStake{
				NodeID:        instances[3].nodeID.Bytes(),
				StakedAmount:  30_000_000_000,
				RewardAddress: rdelegate,
			},
			delegateFactory,
		)
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

		accept := expectBlk(instances[3])
		results := accept(true)
		gomega.Ω(results).Should(gomega.HaveLen(1))
		gomega.Ω(results[0].Success).Should(gomega.BeTrue())

		gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

		fmt.Printf("delegate stake to node 3 %d", height)

		blk := blocks[height]
		ImportBlockToInstance(instances[4].vm, blk)
		ImportBlockToInstance(instances[2].vm, blk)
		ImportBlockToInstance(instances[1].vm, blk)
		ImportBlockToInstance(instances[0].vm, blk)
		height++

		// test same block is accepted
		lastAcceptedBlock3, err := instances[3].vm.LastAccepted(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		lastAcceptedBlock4, err := instances[4].vm.LastAccepted(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		gomega.Ω(lastAcceptedBlock3).To(gomega.Equal(lastAcceptedBlock4))
	})

	// TODO: GetUserStakeFromState is returning an empty value
	// TODO: transactions are played twice because of Verify and Accept (block is already processed)
	/*
			ginkgo.FIt("Get user stake before claim", func() {
				for _, inst := range instances {
					color.Blue("checking %q", inst.nodeID)

					// Ensure all blocks processed
					for {
						_, h, _, err := inst.hcli.Accepted(context.Background())
						gomega.Ω(err).Should(gomega.BeNil())
						if h > 0 {
							break
						}
						time.Sleep(1 * time.Second)
					}
					_, stakedAmount, _, _, err := inst.ncli.UserStake(context.Background(), rdelegate, instances[3].nodeID)
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(stakedAmount).Should(gomega.Equal(uint64(30_000_000_000)))

				}
			})

				ginkgo.It("Claim delegation stake rewards from node 3", func() {
					parser, err := instances[3].ncli.Parser(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					submit, _, _, err := instances[3].hcli.GenerateTransaction(
						context.Background(),
						parser,
						nil,
						&actions.ClaimDelegationStakeRewards{
							NodeID:           instances[3].nodeID.Bytes(),
							UserStakeAddress: rdelegate,
						},
						delegateFactory,
					)
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
					gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

					accept := expectBlk(instances[3])
					results := accept(true)
					gomega.Ω(results).Should(gomega.HaveLen(1))
					gomega.Ω(results[0].Success).Should(gomega.BeTrue())

					gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

					blk := blocks[height]
					fmt.Println(blk.ID())
					ImportBlockToInstance(instances[4].vm, blk)
					ImportBlockToInstance(instances[0].vm, blk)
					ImportBlockToInstance(instances[2].vm, blk)
					ImportBlockToInstance(instances[1].vm, blk)
					height++

					// test same block is accepted
					lastAcceptedBlock3, err := instances[3].vm.LastAccepted(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					lastAcceptedBlock4, err := instances[4].vm.LastAccepted(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(lastAcceptedBlock3).To(gomega.Equal(lastAcceptedBlock4))
				})

				ginkgo.It("Get user stake after claim", func() {
					_, stakedAmount, _, _, err := instances[3].ncli.UserStake(context.Background(), rdelegate, instances[0].nodeID)
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(stakedAmount).Should(gomega.Equal(0))
				})

				ginkgo.It("Undelegate user stake from node 3", func() {
					parser, err := instances[3].ncli.Parser(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					submit, _, _, err := instances[3].hcli.GenerateTransaction(
						context.Background(),
						parser,
						nil,
						&actions.UndelegateUserStake{
							NodeID:        instances[3].nodeID.Bytes(),
							RewardAddress: rdelegate,
						},
						delegateFactory,
					)
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
					gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

					accept := expectBlk(instances[3])
					results := accept(true)
					gomega.Ω(results).Should(gomega.HaveLen(1))
					gomega.Ω(results[0].Success).Should(gomega.BeTrue())

					gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

					blk := blocks[height]
					fmt.Println(blk.ID())
					ImportBlockToInstance(instances[4].vm, blk)
					ImportBlockToInstance(instances[0].vm, blk)
					ImportBlockToInstance(instances[2].vm, blk)
					ImportBlockToInstance(instances[1].vm, blk)
					height++
				})

				ginkgo.It("Claim validator node 0 stake reward", func() {
					parser, err := instances[3].ncli.Parser(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					submit, _, _, err := instances[3].hcli.GenerateTransaction(
						context.Background(),
						parser,
						nil,
						&actions.ClaimValidatorStakeRewards{
							NodeID: instances[3].nodeID.Bytes(),
						},
						factory,
					)
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
					gomega.Ω(instances[3].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))

					accept := expectBlk(instances[3])
					results := accept(true)
					gomega.Ω(results).Should(gomega.HaveLen(1))
					gomega.Ω(results[0].Success).Should(gomega.BeTrue())

					gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

					blk := blocks[height]
					fmt.Println(blk.ID())
					ImportBlockToInstance(instances[4].vm, blk)
					ImportBlockToInstance(instances[0].vm, blk)
					ImportBlockToInstance(instances[2].vm, blk)
					ImportBlockToInstance(instances[1].vm, blk)
					height++

					gomega.Ω(instances[3].ncli.Balance(context.Background(), withdraw0, ids.Empty)).Should(gomega.BeNumerically(">", 0))
				})

				ginkgo.It("Withdraw validator node 0 stake", func() {
					parser, err := instances[3].ncli.Parser(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					submit, _, _, err := instances[3].hcli.GenerateTransaction(
						context.Background(),
						parser,
						nil,
						&actions.WithdrawValidatorStake{
							NodeID:        instances[3].nodeID.Bytes(),
							RewardAddress: rwithdraw0,
						},
						factory,
					)
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

					accept := expectBlk(instances[3])
					results := accept(true)
					gomega.Ω(results).Should(gomega.HaveLen(1))
					gomega.Ω(results[0].Success).Should(gomega.BeTrue())

					gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

					blk := blocks[height]
					fmt.Println(blk.ID())
					ImportBlockToInstance(instances[4].vm, blk)
					ImportBlockToInstance(instances[0].vm, blk)
					ImportBlockToInstance(instances[2].vm, blk)
					ImportBlockToInstance(instances[1].vm, blk)
					height++

				})

				ginkgo.It("Get staked validators after staking withdraw ", func() {
					validators, err := instances[0].ncli.StakedValidators(context.Background())
					gomega.Ω(err).Should(gomega.BeNil())
					gomega.Ω(len(validators)).Should(gomega.Equal(0))
				})

	}) */

	_ = ginkgo.Describe("[Tx Processing]", func() {
		ginkgo.It("get currently accepted block ID", func() {
			for _, inst := range instances {
				hcli := inst.hcli
				_, _, _, err := hcli.Accepted(context.Background())
				gomega.Ω(err).Should(gomega.BeNil())
			}
		})

		var transferTxRoot *chain.Transaction
		ginkgo.It("Gossip TransferTx to a different node", func() {
			ginkgo.By("issue TransferTx", func() {
				parser, err := instances[0].ncli.Parser(context.Background())
				gomega.Ω(err).Should(gomega.BeNil())
				submit, transferTx, _, err := instances[0].hcli.GenerateTransaction(
					context.Background(),
					parser,
					nil,
					&actions.Transfer{
						To:    rsender2,
						Value: 100_000, // must be more than StateLockup
					},
					factory,
				)
				transferTxRoot = transferTx
				gomega.Ω(err).Should(gomega.BeNil())
				gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
				gomega.Ω(instances[0].vm.Mempool().Len(context.Background())).Should(gomega.Equal(1))
			})

			ginkgo.By("skip duplicate", func() {
				_, err := instances[0].hcli.SubmitTx(
					context.Background(),
					transferTxRoot.Bytes(),
				)
				gomega.Ω(err).To(gomega.Not(gomega.BeNil()))
			})

			ginkgo.By("send gossip from node 0 to 1", func() {
				err := instances[0].vm.Gossiper().Force(context.TODO())
				gomega.Ω(err).Should(gomega.BeNil())
			})

			ginkgo.By("skip invalid time", func() {
				tx := chain.NewTx(
					&chain.Base{
						ChainID:   instances[0].chainID,
						Timestamp: 0,
						MaxFee:    1000,
					},
					nil,
					&actions.Transfer{
						To:    rsender2,
						Value: 110,
					},
				)
				// Must do manual construction to avoid `tx.Sign` error (would fail with
				// 0 timestamp)
				msg, err := tx.Digest()
				gomega.Ω(err).To(gomega.BeNil())
				auth, err := factory.Sign(msg)
				gomega.Ω(err).To(gomega.BeNil())
				tx.Auth = auth
				p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
				gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
				gomega.Ω(p.Err()).To(gomega.BeNil())
				_, err = instances[0].hcli.SubmitTx(
					context.Background(),
					p.Bytes(),
				)
				gomega.Ω(err).To(gomega.Not(gomega.BeNil()))
			})

			ginkgo.By("skip duplicate (after gossip, which shouldn't clear)", func() {
				_, err := instances[0].hcli.SubmitTx(
					context.Background(),
					transferTxRoot.Bytes(),
				)
				gomega.Ω(err).To(gomega.Not(gomega.BeNil()))
			})

			ginkgo.By("receive gossip in the node 1, and signal block build", func() {
				gomega.Ω(instances[1].vm.Builder().Force(context.TODO())).To(gomega.BeNil())
				<-instances[1].toEngine
			})

			ginkgo.By("build block in the node 1", func() {
				ctx := context.TODO()
				blk, err := instances[1].vm.BuildBlock(ctx)
				gomega.Ω(err).To(gomega.BeNil())

				gomega.Ω(blk.Verify(ctx)).To(gomega.BeNil())
				gomega.Ω(blk.Status()).To(gomega.Equal(choices.Processing))

				err = instances[1].vm.SetPreference(ctx, blk.ID())
				gomega.Ω(err).To(gomega.BeNil())

				gomega.Ω(blk.Accept(ctx)).To(gomega.BeNil())
				gomega.Ω(blk.Status()).To(gomega.Equal(choices.Accepted))
				blocks = append(blocks, blk)

				lastAccepted, err := instances[1].vm.LastAccepted(ctx)
				gomega.Ω(err).To(gomega.BeNil())
				gomega.Ω(lastAccepted).To(gomega.Equal(blk.ID()))

				results := blk.(*chain.StatelessBlock).Results()
				gomega.Ω(results).Should(gomega.HaveLen(1))
				gomega.Ω(results[0].Success).Should(gomega.BeTrue())
				gomega.Ω(results[0].Output).Should(gomega.BeNil())

				// Unit explanation
				//
				// bandwidth: tx size
				// compute: 5 for signature, 1 for base, 1 for transfer
				// read: 2 keys reads, 1 had 0 chunks
				// allocate: 1 key created
				// write: 1 key modified, 1 key new
				transferTxConsumed := chain.Dimensions{227, 7, 12, 25, 26}
				gomega.Ω(results[0].Consumed).Should(gomega.Equal(transferTxConsumed))

				// Fee explanation
				//
				// Multiply all unit consumption by 1 and sum
				gomega.Ω(results[0].Fee).Should(gomega.Equal(uint64(297)))
			})

			ginkgo.By("ensure balance is updated", func() {
				balance, err := instances[1].ncli.Balance(context.Background(), sender, ids.Empty)
				gomega.Ω(err).To(gomega.BeNil())
				gomega.Ω(balance).To(gomega.Equal(uint64(9999699999899109)))
				balance2, err := instances[1].ncli.Balance(context.Background(), sender2, ids.Empty)
				gomega.Ω(err).To(gomega.BeNil())
				gomega.Ω(balance2).To(gomega.Equal(uint64(100000)))
			})
		})

		ginkgo.It("ensure multiple txs work ", func() {
			ginkgo.By("transfer funds again", func() {
				parser, err := instances[1].ncli.Parser(context.Background())
				gomega.Ω(err).Should(gomega.BeNil())
				submit, _, _, err := instances[1].hcli.GenerateTransaction(
					context.Background(),
					parser,
					nil,
					&actions.Transfer{
						To:    rsender2,
						Value: 101,
					},
					factory,
				)
				gomega.Ω(err).Should(gomega.BeNil())
				gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
				time.Sleep(2 * time.Second) // for replay test
				accept := expectBlk(instances[1])
				results := accept(true)
				gomega.Ω(results).Should(gomega.HaveLen(1))
				gomega.Ω(results[0].Success).Should(gomega.BeTrue())

				balance2, err := instances[1].ncli.Balance(context.Background(), sender2, ids.Empty)
				gomega.Ω(err).To(gomega.BeNil())
				gomega.Ω(balance2).To(gomega.Equal(uint64(100101)))
			})
		})

		ginkgo.It("Test processing block handling", func() {
			var accept, accept2 func(bool) []*chain.Result

			ginkgo.By("create processing tip", func() {
				parser, err := instances[1].ncli.Parser(context.Background())
				gomega.Ω(err).Should(gomega.BeNil())
				submit, _, _, err := instances[1].hcli.GenerateTransaction(
					context.Background(),
					parser,
					nil,
					&actions.Transfer{
						To:    rsender2,
						Value: 200,
					},
					factory,
				)
				gomega.Ω(err).Should(gomega.BeNil())
				gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
				time.Sleep(2 * time.Second) // for replay test
				accept = expectBlk(instances[1])

				submit, _, _, err = instances[1].hcli.GenerateTransaction(
					context.Background(),
					parser,
					nil,
					&actions.Transfer{
						To:    rsender2,
						Value: 201,
					},
					factory,
				)
				gomega.Ω(err).Should(gomega.BeNil())
				gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
				time.Sleep(2 * time.Second) // for replay test
				accept2 = expectBlk(instances[1])
			})

			ginkgo.By("clear processing tip", func() {
				results := accept(true)
				gomega.Ω(results).Should(gomega.HaveLen(1))
				gomega.Ω(results[0].Success).Should(gomega.BeTrue())
				results = accept2(true)
				gomega.Ω(results).Should(gomega.HaveLen(1))
				gomega.Ω(results[0].Success).Should(gomega.BeTrue())
			})
		})

		ginkgo.It("ensure mempool works", func() {
			ginkgo.By("fail Gossip TransferTx to a stale node when missing previous blocks", func() {
				parser, err := instances[1].ncli.Parser(context.Background())
				gomega.Ω(err).Should(gomega.BeNil())
				submit, _, _, err := instances[1].hcli.GenerateTransaction(
					context.Background(),
					parser,
					nil,
					&actions.Transfer{
						To:    rsender2,
						Value: 203,
					},
					factory,
				)
				gomega.Ω(err).Should(gomega.BeNil())
				gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

				err = instances[1].vm.Gossiper().Force(context.TODO())
				gomega.Ω(err).Should(gomega.BeNil())

				// mempool in 0 should be 1 (old amount), since gossip/submit failed
				gomega.Ω(instances[0].vm.Mempool().Len(context.TODO())).Should(gomega.Equal(1))
			})
		})

		ginkgo.It("ensure unprocessed tip and replay protection works", func() {
			ginkgo.By("import accepted blocks to instance 2", func() {
				ctx := context.TODO()

				gomega.Ω(blocks[0].Height()).Should(gomega.Equal(uint64(1)))

				n := instances[2]
				blk1, err := n.vm.ParseBlock(ctx, blocks[4].Bytes())
				gomega.Ω(err).Should(gomega.BeNil())
				err = blk1.Verify(ctx)
				gomega.Ω(err).Should(gomega.BeNil())

				// Parse tip
				blk2, err := n.vm.ParseBlock(ctx, blocks[5].Bytes())
				gomega.Ω(err).Should(gomega.BeNil())
				blk3, err := n.vm.ParseBlock(ctx, blocks[6].Bytes())
				gomega.Ω(err).Should(gomega.BeNil())

				// Verify tip
				err = blk2.Verify(ctx)
				gomega.Ω(err).Should(gomega.BeNil())
				err = blk3.Verify(ctx)
				gomega.Ω(err).Should(gomega.BeNil())

				// Check if tx from old block would be considered a repeat on processing tip
				tx := blk2.(*chain.StatelessBlock).Txs[0]
				sblk3 := blk3.(*chain.StatelessBlock)
				sblk3t := sblk3.Timestamp().UnixMilli()
				ok, err := sblk3.IsRepeat(ctx, sblk3t-n.vm.Rules(sblk3t).GetValidityWindow(), []*chain.Transaction{tx}, set.NewBits(), false)
				gomega.Ω(err).Should(gomega.BeNil())
				gomega.Ω(ok.Len()).Should(gomega.Equal(1))

				// Accept tip
				err = blk1.Accept(ctx)
				gomega.Ω(err).Should(gomega.BeNil())
				err = blk2.Accept(ctx)
				gomega.Ω(err).Should(gomega.BeNil())
				err = blk3.Accept(ctx)
				gomega.Ω(err).Should(gomega.BeNil())

				// Parse another
				blk4, err := n.vm.ParseBlock(ctx, blocks[7].Bytes())
				gomega.Ω(err).Should(gomega.BeNil())
				err = blk4.Verify(ctx)
				gomega.Ω(err).Should(gomega.BeNil())
				err = blk4.Accept(ctx)
				gomega.Ω(err).Should(gomega.BeNil())

				// Check if tx from old block would be considered a repeat on accepted tip
				time.Sleep(2 * time.Second)
				gomega.Ω(n.vm.IsRepeat(ctx, []*chain.Transaction{tx}, set.NewBits(), false).Len()).Should(gomega.Equal(1))
			})
		})

		ginkgo.It("processes valid index transactions (w/block listening)", func() {
			// Clear previous txs on instance 0
			accept := expectBlk(instances[0])
			accept(false) // don't care about results

			// Subscribe to blocks
			hcli, err := hrpc.NewWebSocketClient(instances[0].WebSocketServer.URL, hrpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(hcli.RegisterBlocks()).Should(gomega.BeNil())

			// Wait for message to be sent
			time.Sleep(2 * pubsub.MaxMessageWait)

			// Fetch balances
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())

			// Send tx
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			transfer := &actions.Transfer{
				To:    auth.NewED25519Address(other.PublicKey()),
				Value: 1,
			}

			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				transfer,
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

			gomega.Ω(err).Should(gomega.BeNil())
			accept = expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			// Read item from connection
			blk, lresults, prices, err := hcli.ListenBlock(context.TODO(), parser)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(len(blk.Txs)).Should(gomega.Equal(1))
			tx := blk.Txs[0].Action.(*actions.Transfer)
			gomega.Ω(tx.Asset).To(gomega.Equal(ids.Empty))
			gomega.Ω(tx.Value).To(gomega.Equal(uint64(1)))
			gomega.Ω(lresults).Should(gomega.Equal(results))
			gomega.Ω(prices).Should(gomega.Equal(chain.Dimensions{1, 1, 1, 1, 1}))

			// Check balance modifications are correct
			balancea, err := instances[0].ncli.Balance(context.TODO(), sender, ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(balancea + lresults[0].Fee + 1))

			// Close connection when done
			gomega.Ω(hcli.Close()).Should(gomega.BeNil())
		})

		ginkgo.It("processes valid index transactions (w/streaming verification)", func() {
			// Create streaming client
			hcli, err := hrpc.NewWebSocketClient(instances[0].WebSocketServer.URL, hrpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
			gomega.Ω(err).Should(gomega.BeNil())

			// Create tx
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			transfer := &actions.Transfer{
				To:    auth.NewED25519Address(other.PublicKey()),
				Value: 1,
			}
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			_, tx, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				transfer,
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())

			// Submit tx and accept block
			gomega.Ω(hcli.RegisterTx(tx)).Should(gomega.BeNil())

			// Wait for message to be sent
			time.Sleep(2 * pubsub.MaxMessageWait)

			for instances[0].vm.Mempool().Len(context.TODO()) == 0 {
				// We need to wait for mempool to be populated because issuance will
				// return as soon as bytes are on the channel.
				hutils.Outf("{{yellow}}waiting for mempool to return non-zero txs{{/}}\n")
				time.Sleep(500 * time.Millisecond)
			}
			gomega.Ω(err).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			// Read decision from connection
			txID, dErr, result, err := hcli.ListenTx(context.TODO())
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(txID).Should(gomega.Equal(tx.ID()))
			gomega.Ω(dErr).Should(gomega.BeNil())
			gomega.Ω(result.Success).Should(gomega.BeTrue())
			gomega.Ω(result).Should(gomega.Equal(results[0]))

			// Close connection when done
			gomega.Ω(hcli.Close()).Should(gomega.BeNil())
		})

		ginkgo.It("transfer an asset with a memo", func() {
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.Transfer{
					To:    auth.NewED25519Address(other.PublicKey()),
					Value: 10,
					Memo:  []byte("hello"),
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeTrue())
		})

		ginkgo.It("transfer an asset with large memo", func() {
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1001,
				},
				nil,
				&actions.Transfer{
					To:    auth.NewED25519Address(other.PublicKey()),
					Value: 10,
					Memo:  make([]byte, 1000),
				},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// too large)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("size is larger than limit"))
		})

		ginkgo.It("mint an asset that doesn't exist", func() {
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			assetID := ids.GenerateTestID()
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.MintAsset{
					To:    auth.NewED25519Address(other.PublicKey()),
					Asset: assetID,
					Value: 10,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeFalse())
			gomega.Ω(string(result.Output)).
				Should(gomega.ContainSubstring("asset missing"))

			exists, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), assetID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeFalse())
		})

		ginkgo.It("create a new asset (no metadata)", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1001,
				},
				nil,
				&actions.CreateAsset{
					Symbol:   []byte("s0"),
					Decimals: 0,
					Metadata: nil,
				},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// too large)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("Bytes field is not populated"))
		})

		ginkgo.It("create a new asset (no symbol)", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1001,
				},
				nil,
				&actions.CreateAsset{
					Symbol:   nil,
					Decimals: 0,
					Metadata: []byte("m"),
				},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// too large)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("Bytes field is not populated"))
		})

		ginkgo.It("create asset with too long of metadata", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				nil,
				&actions.CreateAsset{
					Symbol:   []byte("s0"),
					Decimals: 0,
					Metadata: make([]byte, actions.MaxMetadataSize*2),
				},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// too large)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("size is larger than limit"))
		})

		ginkgo.It("create a new asset (simple metadata)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, tx, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.CreateAsset{
					Symbol:   asset1Symbol,
					Decimals: asset1Decimals,
					Metadata: asset1,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			asset1ID = tx.ID()
			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(0)))

			exists, symbol, decimals, metadata, supply, owner, warp, err := instances[0].ncli.Asset(context.TODO(), asset1ID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeTrue())
			gomega.Ω(symbol).Should(gomega.Equal(asset1Symbol))
			gomega.Ω(decimals).Should(gomega.Equal(asset1Decimals))
			gomega.Ω(metadata).Should(gomega.Equal(asset1))
			gomega.Ω(supply).Should(gomega.Equal(uint64(0)))
			gomega.Ω(owner).Should(gomega.Equal(sender))
			gomega.Ω(warp).Should(gomega.BeFalse())
		})

		ginkgo.It("mint a new asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.MintAsset{
					To:    rsender2,
					Asset: asset1ID,
					Value: 15,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(15)))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(0)))

			exists, symbol, decimals, metadata, supply, owner, warp, err := instances[0].ncli.Asset(context.TODO(), asset1ID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeTrue())
			gomega.Ω(symbol).Should(gomega.Equal(asset1Symbol))
			gomega.Ω(decimals).Should(gomega.Equal(asset1Decimals))
			gomega.Ω(metadata).Should(gomega.Equal(asset1))
			gomega.Ω(supply).Should(gomega.Equal(uint64(15)))
			gomega.Ω(owner).Should(gomega.Equal(sender))
			gomega.Ω(warp).Should(gomega.BeFalse())
		})

		ginkgo.It("mint asset from wrong owner", func() {
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.MintAsset{
					To:    auth.NewED25519Address(other.PublicKey()),
					Asset: asset1ID,
					Value: 10,
				},
				factory2,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeFalse())
			gomega.Ω(string(result.Output)).
				Should(gomega.ContainSubstring("wrong owner"))

			exists, symbol, decimals, metadata, supply, owner, warp, err := instances[0].ncli.Asset(context.TODO(), asset1ID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeTrue())
			gomega.Ω(symbol).Should(gomega.Equal(asset1Symbol))
			gomega.Ω(decimals).Should(gomega.Equal(asset1Decimals))
			gomega.Ω(metadata).Should(gomega.Equal(asset1))
			gomega.Ω(supply).Should(gomega.Equal(uint64(15)))
			gomega.Ω(owner).Should(gomega.Equal(sender))
			gomega.Ω(warp).Should(gomega.BeFalse())
		})

		ginkgo.It("burn new asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.BurnAsset{
					Asset: asset1ID,
					Value: 5,
				},
				factory2,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(10)))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(0)))

			exists, symbol, decimals, metadata, supply, owner, warp, err := instances[0].ncli.Asset(context.TODO(), asset1ID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeTrue())
			gomega.Ω(symbol).Should(gomega.Equal(asset1Symbol))
			gomega.Ω(decimals).Should(gomega.Equal(asset1Decimals))
			gomega.Ω(metadata).Should(gomega.Equal(asset1))
			gomega.Ω(supply).Should(gomega.Equal(uint64(10)))
			gomega.Ω(owner).Should(gomega.Equal(sender))
			gomega.Ω(warp).Should(gomega.BeFalse())
		})

		ginkgo.It("burn missing asset", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.BurnAsset{
					Asset: asset1ID,
					Value: 10,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeFalse())
			gomega.Ω(string(result.Output)).
				Should(gomega.ContainSubstring("invalid balance"))

			exists, symbol, decimals, metadata, supply, owner, warp, err := instances[0].ncli.Asset(context.TODO(), asset1ID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeTrue())
			gomega.Ω(symbol).Should(gomega.Equal(asset1Symbol))
			gomega.Ω(decimals).Should(gomega.Equal(asset1Decimals))
			gomega.Ω(metadata).Should(gomega.Equal(asset1))
			gomega.Ω(supply).Should(gomega.Equal(uint64(10)))
			gomega.Ω(owner).Should(gomega.Equal(sender))
			gomega.Ω(warp).Should(gomega.BeFalse())
		})

		ginkgo.It("rejects empty mint", func() {
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				nil,
				&actions.MintAsset{
					To:    auth.NewED25519Address(other.PublicKey()),
					Asset: asset1ID,
				},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// bad codec)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("Uint64 field is not populated"))
		})

		ginkgo.It("reject max mint", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.MintAsset{
					To:    rsender2,
					Asset: asset1ID,
					Value: hconsts.MaxUint64,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeFalse())
			gomega.Ω(string(result.Output)).
				Should(gomega.ContainSubstring("overflow"))

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(10)))
			balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(0)))

			exists, symbol, decimals, metadata, supply, owner, warp, err := instances[0].ncli.Asset(context.TODO(), asset1ID, false)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(exists).Should(gomega.BeTrue())
			gomega.Ω(symbol).Should(gomega.Equal(asset1Symbol))
			gomega.Ω(decimals).Should(gomega.Equal(asset1Decimals))
			gomega.Ω(metadata).Should(gomega.Equal(asset1))
			gomega.Ω(supply).Should(gomega.Equal(uint64(10)))
			gomega.Ω(owner).Should(gomega.Equal(sender))
			gomega.Ω(warp).Should(gomega.BeFalse())
		})

		ginkgo.It("rejects mint of native token", func() {
			other, err := ed25519.GeneratePrivateKey()
			gomega.Ω(err).Should(gomega.BeNil())
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				nil,
				&actions.MintAsset{
					To:    auth.NewED25519Address(other.PublicKey()),
					Value: 10,
				},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// bad codec)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("ID field is not populated"))
		})

		ginkgo.It("mints another new asset (to self)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, tx, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.CreateAsset{
					Symbol:   asset2Symbol,
					Decimals: asset2Decimals,
					Metadata: asset2,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())
			asset2ID = tx.ID()

			submit, _, _, err = instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.MintAsset{
					To:    rsender,
					Asset: asset2ID,
					Value: 10,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept = expectBlk(instances[0])
			results = accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset2ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(10)))
		})

		ginkgo.It("mints another new asset (to self) on another account", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, tx, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.CreateAsset{
					Symbol:   asset3Symbol,
					Decimals: asset3Decimals,
					Metadata: asset3,
				},
				factory2,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())
			asset3ID = tx.ID()

			submit, _, _, err = instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.MintAsset{
					To:    rsender2,
					Asset: asset3ID,
					Value: 10,
				},
				factory2,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept = expectBlk(instances[0])
			results = accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset3ID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(10)))
		})

		ginkgo.It("import warp message with nil when expected", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				nil,
				&actions.ImportAsset{},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// empty warp)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("expected warp message"))
		})

		ginkgo.It("import warp message empty", func() {
			wm, err := warp.NewMessage(&warp.UnsignedMessage{}, &warp.BitSetSignature{})
			gomega.Ω(err).Should(gomega.BeNil())
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				wm,
				&actions.ImportAsset{},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// empty warp)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("empty warp payload"))
		})

		ginkgo.It("import with wrong payload", func() {
			uwm, err := warp.NewUnsignedMessage(networkID, ids.Empty, []byte("hello"))
			gomega.Ω(err).Should(gomega.BeNil())
			wm, err := warp.NewMessage(uwm, &warp.BitSetSignature{})
			gomega.Ω(err).Should(gomega.BeNil())
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				wm,
				&actions.ImportAsset{},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// invalid object)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("insufficient length for input"))
		})

		ginkgo.It("import with invalid payload", func() {
			wt := &actions.WarpTransfer{}
			wtb, err := wt.Marshal()
			gomega.Ω(err).Should(gomega.BeNil())
			uwm, err := warp.NewUnsignedMessage(networkID, ids.Empty, wtb)
			gomega.Ω(err).Should(gomega.BeNil())
			wm, err := warp.NewMessage(uwm, &warp.BitSetSignature{})
			gomega.Ω(err).Should(gomega.BeNil())
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: hutils.UnixRMilli(-1, 5*hconsts.MillisecondsPerSecond),
					MaxFee:    1000,
				},
				wm,
				&actions.ImportAsset{},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// invalid object)
			msg, err := tx.Digest()
			gomega.Ω(err).To(gomega.BeNil())
			auth, err := factory.Sign(msg)
			gomega.Ω(err).To(gomega.BeNil())
			tx.Auth = auth
			p := codec.NewWriter(0, hconsts.MaxInt) // test codec growth
			gomega.Ω(tx.Marshal(p)).To(gomega.BeNil())
			gomega.Ω(p.Err()).To(gomega.BeNil())
			_, err = instances[0].hcli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			gomega.Ω(err.Error()).Should(gomega.ContainSubstring("field is not populated"))
		})

		ginkgo.It("import with wrong destination", func() {
			wt := &actions.WarpTransfer{
				To:                 rsender,
				Symbol:             []byte("s"),
				Decimals:           2,
				Asset:              ids.GenerateTestID(),
				Value:              100,
				Return:             false,
				Reward:             100,
				TxID:               ids.GenerateTestID(),
				DestinationChainID: ids.GenerateTestID(),
			}
			wtb, err := wt.Marshal()
			gomega.Ω(err).Should(gomega.BeNil())
			uwm, err := warp.NewUnsignedMessage(networkID, ids.Empty, wtb)
			gomega.Ω(err).Should(gomega.BeNil())
			wm, err := warp.NewMessage(uwm, &warp.BitSetSignature{})
			gomega.Ω(err).Should(gomega.BeNil())
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				wm,
				&actions.ImportAsset{},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

			// Build block with no context (should fail)
			gomega.Ω(instances[0].vm.Builder().Force(context.TODO())).To(gomega.BeNil())
			<-instances[0].toEngine
			blk, err := instances[0].vm.BuildBlock(context.TODO())
			gomega.Ω(err).To(gomega.Not(gomega.BeNil()))
			gomega.Ω(blk).To(gomega.BeNil())

			// Wait for mempool to be size 1 (txs are restored async)
			for {
				if instances[0].vm.Mempool().Len(context.Background()) > 0 {
					break
				}
				log.Info("waiting for txs to be restored")
				time.Sleep(100 * time.Millisecond)
			}

			// Build block with context
			accept := expectBlkWithContext(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeFalse())
			gomega.Ω(string(result.Output)).Should(gomega.ContainSubstring("warp verification failed"))
		})

		ginkgo.It("export native asset", func() {
			dest := ids.GenerateTestID()
			loan, err := instances[0].ncli.Loan(context.TODO(), ids.Empty, dest)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(loan).Should(gomega.Equal(uint64(0)))

			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, tx, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.ExportAsset{
					To:          rsender,
					Asset:       ids.Empty,
					Value:       100,
					Return:      false,
					Reward:      10,
					Destination: dest,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeTrue())
			wt := &actions.WarpTransfer{
				To:                 rsender,
				Symbol:             []byte(nconsts.Symbol),
				Decimals:           nconsts.Decimals,
				Asset:              ids.Empty,
				Value:              100,
				Return:             false,
				Reward:             10,
				TxID:               tx.ID(),
				DestinationChainID: dest,
			}
			wtb, err := wt.Marshal()
			gomega.Ω(err).Should(gomega.BeNil())
			wm, err := warp.NewUnsignedMessage(networkID, instances[0].chainID, wtb)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(result.WarpMessage).Should(gomega.Equal(wm))

			loan, err = instances[0].ncli.Loan(context.TODO(), ids.Empty, dest)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(loan).Should(gomega.Equal(uint64(110)))
		})

		ginkgo.It("export native asset (invalid return)", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.ExportAsset{
					To:          rsender,
					Asset:       ids.Empty,
					Value:       100,
					Return:      true,
					Reward:      10,
					Destination: ids.GenerateTestID(),
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			accept := expectBlk(instances[0])
			results := accept(false)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			result := results[0]
			gomega.Ω(result.Success).Should(gomega.BeFalse())
			gomega.Ω(string(result.Output)).Should(gomega.ContainSubstring("not warp asset"))
		})
	})
})

func expectBlk(i instance) func(bool) []*chain.Result {
	ctx := context.TODO()

	// manually signal ready
	gomega.Ω(i.vm.Builder().Force(ctx)).To(gomega.BeNil())
	// manually ack ready sig as in engine
	<-i.toEngine

	blk, err := i.vm.BuildBlock(ctx)
	gomega.Ω(err).To(gomega.BeNil())
	gomega.Ω(blk).To(gomega.Not(gomega.BeNil()))

	gomega.Ω(blk.Verify(ctx)).To(gomega.BeNil())
	gomega.Ω(blk.Status()).To(gomega.Equal(choices.Processing))

	err = i.vm.SetPreference(ctx, blk.ID())
	gomega.Ω(err).To(gomega.BeNil())

	return func(add bool) []*chain.Result {
		gomega.Ω(blk.Accept(ctx)).To(gomega.BeNil())
		gomega.Ω(blk.Status()).To(gomega.Equal(choices.Accepted))

		if add {
			blocks = append(blocks, blk)
		}

		lastAccepted, err := i.vm.LastAccepted(ctx)
		gomega.Ω(err).To(gomega.BeNil())
		gomega.Ω(lastAccepted).To(gomega.Equal(blk.ID()))
		return blk.(*chain.StatelessBlock).Results()
	}
}

// TODO: unify with expectBlk
func expectBlkWithContext(i instance) func(bool) []*chain.Result {
	ctx := context.TODO()

	// manually signal ready
	gomega.Ω(i.vm.Builder().Force(ctx)).To(gomega.BeNil())
	// manually ack ready sig as in engine
	<-i.toEngine

	bctx := &block.Context{PChainHeight: 1}
	blk, err := i.vm.BuildBlockWithContext(ctx, bctx)
	gomega.Ω(err).To(gomega.BeNil())
	gomega.Ω(blk).To(gomega.Not(gomega.BeNil()))
	cblk := blk.(block.WithVerifyContext)

	gomega.Ω(cblk.VerifyWithContext(ctx, bctx)).To(gomega.BeNil())
	gomega.Ω(blk.Status()).To(gomega.Equal(choices.Processing))

	err = i.vm.SetPreference(ctx, blk.ID())
	gomega.Ω(err).To(gomega.BeNil())

	return func(add bool) []*chain.Result {
		gomega.Ω(blk.Accept(ctx)).To(gomega.BeNil())
		gomega.Ω(blk.Status()).To(gomega.Equal(choices.Accepted))

		if add {
			blocks = append(blocks, blk)
		}

		lastAccepted, err := i.vm.LastAccepted(ctx)
		gomega.Ω(err).To(gomega.BeNil())
		gomega.Ω(lastAccepted).To(gomega.Equal(blk.ID()))
		return blk.(*chain.StatelessBlock).Results()
	}
}

var _ common.AppSender = &appSender{}

type appSender struct {
	next      int
	instances []instance
}

func (app *appSender) SendAppGossip(ctx context.Context, appGossipBytes []byte) error {
	n := len(app.instances)
	sender := app.instances[app.next].nodeID
	app.next++
	app.next %= n
	return app.instances[app.next].vm.AppGossip(ctx, sender, appGossipBytes)
}

func (*appSender) SendAppRequest(context.Context, set.Set[ids.NodeID], uint32, []byte) error {
	return nil
}

func (*appSender) SendAppResponse(context.Context, ids.NodeID, uint32, []byte) error {
	return nil
}

func (*appSender) SendAppGossipSpecific(context.Context, set.Set[ids.NodeID], []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppRequest(context.Context, ids.ID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppResponse(context.Context, ids.ID, uint32, []byte) error {
	return nil
}

func ImportBlockToInstance(vm *vm.VM, block snowman.Block) {
	blk, err := vm.ParseBlock(context.Background(), block.Bytes())
	gomega.Ω(err).Should(gomega.BeNil())
	err = blk.Verify(context.Background())
	gomega.Ω(err).Should(gomega.BeNil())

	gomega.Ω(blk.Status()).To(gomega.Equal(choices.Processing))
	err = vm.SetPreference(context.Background(), blk.ID())
	gomega.Ω(err).To(gomega.BeNil())

	err = blk.Accept(context.Background())
	gomega.Ω(err).Should(gomega.BeNil())
}

func setEmissionValidators() {
	currentValidators := make([]*emission.Validator, 0, len(instances))
	for i, inst := range instances {
		val := emission.Validator{
			NodeID:    inst.nodeID,
			PublicKey: bls.PublicKeyToBytes(nodesPubKeys[i]),
		}
		currentValidators = append(currentValidators, &val)
	}
	for i := range instances {
		emissions[i].(*emission.Manual).CurrentValidators = currentValidators
	}
}
