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
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	hrpc "github.com/ava-labs/hypersdk/rpc"
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
	app            *appSender
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

	//to be reordered staking vars
	currentTime       time.Time
	stakeStartTime    time.Time
	stakeEndTime      time.Time
	delegationFeeRate int
	withdraw0         string
	withdraw1         string
	delegate          string
	rwithdraw0        codec.Address
	rwithdraw1        codec.Address
	rdelegate         codec.Address
	delegateFactory   *auth.ED25519Factory
	nodesFactories    []*auth.BLSFactory
	nodesAddresses    []codec.Address
	emissions         []emission.Tracker
	nodesPubKeys      []*bls.PublicKey
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

var _ = ginkgo.Describe("[Nuklai staking mechanism]", func() {
	ginkgo.FIt("Setup and get initial staked validators", func() {
		height := 0
		currentTime = time.Now().UTC()
		stakeStartTime = currentTime.Add(2 * time.Second)
		stakeEndTime = currentTime.Add(15 * time.Minute)
		delegationFeeRate = 50

		withdraw0Priv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		rwithdraw0 = auth.NewED25519Address(withdraw0Priv.PublicKey())
		withdraw0 = codec.MustAddressBech32(nconsts.HRP, rwithdraw0)

		withdraw1Priv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		rwithdraw1 = auth.NewED25519Address(withdraw1Priv.PublicKey())
		withdraw1 = codec.MustAddressBech32(nconsts.HRP, rwithdraw1)

		delegatePriv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		rdelegate = auth.NewED25519Address(delegatePriv.PublicKey())
		delegate = codec.MustAddressBech32(nconsts.HRP, rdelegate)
		delegateFactory = auth.NewED25519Factory(delegatePriv)

		ginkgo.By("Initial staked validators", func() {
			validators, err := instances[3].ncli.StakedValidators(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(len(validators)).Should(gomega.Equal(0))
		})

		ginkgo.By("Funding node 3", func() {
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
			fmt.Println(err)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

			// app.SendAppGossip(context.Background(), tx.Bytes())

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
			fmt.Printf("node 3 instances[3] %s balance %d\n", codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), balance)
			gomega.Ω(balance).Should(gomega.Equal(uint64(200_000_000_000)))

			// check if gossip/ new state happens
			balanceOther, err := instances[4].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())
			fmt.Printf("node 3 instances[4] %s balance %d\n", codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), balanceOther)
			gomega.Ω(balanceOther).Should(gomega.Equal(uint64(200_000_000_000)))

			balance, err = instances[3].ncli.Balance(context.TODO(), sender, ids.Empty)
			fmt.Printf("balance factory %s %d\n", sender, balance)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(balance).Should(gomega.Equal(uint64(9_999_799_999_999_703)))
		})

		ginkgo.By("Register validator stake node 3", func() {
			parser, err := instances[3].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			stakeInfo := &actions.ValidatorStakeInfo{
				NodeID:            instances[3].nodeID.Bytes(),
				StakeStartTime:    uint64(stakeStartTime.Unix()),
				StakeEndTime:      uint64(stakeEndTime.Unix()),
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
			fmt.Println(err)
			gomega.Ω(err).Should(gomega.BeNil())

			balance, err := instances[3].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())
			fmt.Printf("node 3 %s balance before staking %d\n", codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), balance)

			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

			fmt.Println(len(blocks))

			fmt.Println(blocks)

			accept := expectBlk(instances[3])
			results := accept(true)
			gomega.Ω(results).Should(gomega.HaveLen(1))
			gomega.Ω(results[0].Success).Should(gomega.BeTrue())

			gomega.Ω(len(blocks)).Should(gomega.Equal(height + 1))

			for _, b := range blocks {
				fmt.Println(b.ID())
				fmt.Println(b.Parent())
			}

			fmt.Println(blocks)

			blk := blocks[1]
			fmt.Println(blk.ID())
			ImportBlockToInstance(instances[4].vm, blk)
			ImportBlockToInstance(instances[0].vm, blk)
			ImportBlockToInstance(instances[2].vm, blk)
			ImportBlockToInstance(instances[1].vm, blk)
			height++

			balance, err = instances[3].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())
			fmt.Printf("node 3 instances[3] %s balance after staking %d\n", codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), balance)

			// check if gossip/ new state happens
			balanceOther, err := instances[4].ncli.Balance(context.TODO(), codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), ids.Empty)
			gomega.Ω(err).Should(gomega.BeNil())
			fmt.Printf("node 3 instances[4] %s balance after staking %d\n", codec.MustAddressBech32(nconsts.HRP, nodesAddresses[3]), balanceOther)

			emissionInstance := emissions[3]
			currentValidators := emissionInstance.GetAllValidators(context.TODO())
			gomega.Ω(len(currentValidators)).To(gomega.Equal(5))
			stakedValidator := emissionInstance.GetStakedValidator(instances[3].nodeID)
			fmt.Println(stakedValidator)
			gomega.Ω(len(stakedValidator)).To(gomega.Equal(1))

		})

		ginkgo.By("Get validator staked amount after node 3 validator staking", func() {
			_, _, stakedAmount, _, _, _, err := instances[3].ncli.ValidatorStake(context.Background(), instances[3].nodeID)
			fmt.Printf("NODE 3 STAKED AMOUNT %d", stakedAmount)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.Equal(uint64(100_000_000_000)))
		})

		ginkgo.By("Get validator staked amount after staking using node 0 cli", func() {
			_, _, stakedAmount, _, _, _, err := instances[0].ncli.ValidatorStake(context.Background(), instances[3].nodeID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.Equal(uint64(100_000_000_000)))
		})

		// ginkgo.By("Get staked validators", func() {
		// 	fmt.Println("GET EMISSION 3 VALIDATORS")
		// 	fmt.Println(emissions[0].GetEmissionValidators())
		// 	fmt.Println(emissions[1].GetEmissionValidators())
		// 	fmt.Println(emissions[2].GetEmissionValidators())
		// 	fmt.Println(emissions[3].GetEmissionValidators())
		// 	fmt.Println(emissions[4].GetEmissionValidators())
		// 	validators, err := instances[3].ncli.StakedValidators(context.TODO())
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(len(validators)).Should(gomega.Equal(2))
		// })

		ginkgo.By("Transfer NAI to delegate user", func() {
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
		})

		ginkgo.By("delegate stake to node 3", func() {
			parser, err := instances[3].ncli.Parser(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			currentTime := time.Now().UTC()
			userStakeStartTime := currentTime.Add(5 * time.Second)
			submit, _, _, err := instances[3].hcli.GenerateTransaction(
				context.Background(),
				parser,
				nil,
				&actions.DelegateUserStake{
					NodeID:         instances[3].nodeID.Bytes(),
					StakeStartTime: uint64(userStakeStartTime.Unix()),
					StakedAmount:   30_000_000_000,
					RewardAddress:  rdelegate,
				},
				delegateFactory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())

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

		})

		// ginkgo.By("Delegate stake to node 1", func() {
		// 	currentTime = time.Now().UTC()
		// 	userStakeStartTime := currentTime.Add(2 * time.Minute)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	submit, _, _, err := instances[0].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.DelegateUserStake{
		// 			NodeID:         instances[0].nodeID.Bytes(),
		// 			StakeStartTime: uint64(userStakeStartTime.Unix()),
		// 			StakedAmount:   50,
		// 			RewardAddress:  rdelegate,
		// 		},
		// 		delegateFactory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })

		// ginkgo.By("Get user stake before claim", func() {
		// 	_, stakedAmount, _, _, err := instances[0].ncli.UserStake(context.Background(), rdelegate, instances[0].nodeID)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(stakedAmount).Should(gomega.BeNumerically("==", 50))
		// })

		// ginkgo.By("Claim delegation stake rewards from node 0", func() {
		// 	submit, _, _, err := instances[0].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.ClaimDelegationStakeRewards{
		// 			NodeID:           instances[0].nodeID.Bytes(),
		// 			UserStakeAddress: rdelegate,
		// 		},
		// 		delegateFactory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })

		// ginkgo.By("Get user stake after claim", func() {
		// 	_, stakedAmount, _, _, err := instances[0].ncli.UserStake(context.Background(), rdelegate, instances[0].nodeID)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(stakedAmount).Should(gomega.BeNumerically("==", 0))
		// })

		// ginkgo.By("Undelegate user stake from node 0", func() {
		// 	submit, _, _, err := instances[0].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.UndelegateUserStake{
		// 			NodeID: instances[0].nodeID.Bytes(),
		// 		},
		// 		delegateFactory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })

		// add more ginko.By where error should be thrown with wrong data input
		// ginkgo.By("Claim validator node 0 stake reward", func() {
		// 	// ClaimValidatorStakeRewards
		// 	// TO DO: test claim with a wrong key
		// 	submit, _, _, err := instances[0].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.ClaimValidatorStakeRewards{
		// 			NodeID: instances[0].nodeID.Bytes(),
		// 		},
		// 		factory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(instances[0].ncli.Balance(context.Background(), withdraw0, ids.Empty)).Should(gomega.BeNumerically(">", 0))
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })

		// ginkgo.By("Withdraw validator node 0 stake", func() {
		// 	// WithdrawValidatorStake
		// 	// TO DO: test claim with a wrong key
		// 	submit, _, _, err := instances[0].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.WithdrawValidatorStake{
		// 			NodeID: instances[0].nodeID.Bytes(),
		// 		},
		// 		factory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })

		// ginkgo.By("Get validator stake after staking withdraw ", func() {
		// 	_, _, stakedAmount, _, _, _, err := instances[0].ncli.ValidatorStake(context.Background(), instances[0].nodeID)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(stakedAmount).Should(gomega.Equal(0))
		// })

		// ginkgo.By("Withdraw validator node 1 stake with wrong key", func() {
		// 	// WithdrawValidatorStake
		// 	// TO DO: test claim with a wrong key
		// 	submit, _, _, err := instances[1].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.WithdrawValidatorStake{
		// 			NodeID: instances[1].nodeID.Bytes(),
		// 		},
		// 		factory2,
		// 	)
		// 	gomega.Ω(err).ShouldNot(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).ShouldNot(gomega.BeNil())
		// })

		// staking withdraw then user delegate withdraw on node 1
		// ginkgo.By("Withdraw validator node 1 stake", func() {
		// 	// WithdrawValidatorStake
		// 	// TO DO: test claim with a wrong key
		// 	submit, _, _, err := instances[1].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.WithdrawValidatorStake{
		// 			NodeID: instances[1].nodeID.Bytes(),
		// 		},
		// 		factory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })

		// ginkgo.By("Undelegate user stake from node 1", func() {
		// 	submit, _, _, err := instances[1].hcli.GenerateTransaction(
		// 		context.Background(),
		// 		parser0,
		// 		nil,
		// 		&actions.UndelegateUserStake{
		// 			NodeID: instances[0].nodeID.Bytes(),
		// 		},
		// 		delegateFactory,
		// 	)
		// 	gomega.Ω(err).Should(gomega.BeNil())
		// 	gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		// })
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
	fmt.Println(len(currentValidators))
	for i := range instances {
		fmt.Println(emissions[i])
		emissions[i].(*emission.Manual).CurrentValidators = currentValidators
	}
}
