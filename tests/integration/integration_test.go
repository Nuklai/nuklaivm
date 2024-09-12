// Copyright (C) 2024, Nuklai. All rights reserved.
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
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/fatih/color"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"
	"github.com/ava-labs/hypersdk/vm"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/controller"
	"github.com/nuklai/nuklaivm/genesis"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

var (
	logFactory logging.Factory
	log        logging.Logger

	requestTimeout time.Duration
	vms            int

	priv    ed25519.PrivateKey
	factory *auth.ED25519Factory
	rsender codec.Address
	sender  string

	priv2    ed25519.PrivateKey
	factory2 *auth.ED25519Factory
	rsender2 codec.Address
	sender2  string

	priv3    ed25519.PrivateKey
	factory3 *auth.ED25519Factory
	rsender3 codec.Address
	sender3  string

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
	asset4         []byte
	asset4Symbol   []byte
	asset4Decimals uint8

	// when used with embedded VMs
	genesisBytes []byte
	instances    []instance
	blocks       []snowman.Block

	networkID uint32
	gen       *genesis.Genesis
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
	ginkgo.RunSpecs(t, "tokenvm integration test suites")
}

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
		4,
		"number of VMs to create",
	)
}

type instance struct {
	chainID             ids.ID
	nodeID              ids.NodeID
	vm                  *vm.VM
	toEngine            chan common.Message
	JSONRPCServer       *httptest.Server
	NuklaiJSONRPCServer *httptest.Server
	WebSocketServer     *httptest.Server
	cli                 *rpc.JSONRPCClient // clients for embedded VMs
	ncli                *nrpc.JSONRPCClient
}

var _ = ginkgo.BeforeSuite(func() {
	require := require.New(ginkgo.GinkgoT())

	require.Greater(vms, 1)

	var err error
	priv, err = ed25519.GeneratePrivateKey()
	require.NoError(err)
	factory = auth.NewED25519Factory(priv)
	rsender = auth.NewED25519Address(priv.PublicKey())
	sender = codec.MustAddressBech32(nconsts.HRP, rsender)
	log.Debug(
		"generated key",
		zap.String("addr", sender),
		zap.String("pk", hex.EncodeToString(priv[:])),
	)

	priv2, err = ed25519.GeneratePrivateKey()
	require.NoError(err)
	factory2 = auth.NewED25519Factory(priv2)
	rsender2 = auth.NewED25519Address(priv2.PublicKey())
	sender2 = codec.MustAddressBech32(nconsts.HRP, rsender2)
	log.Debug(
		"generated key",
		zap.String("addr", sender2),
		zap.String("pk", hex.EncodeToString(priv2[:])),
	)

	priv3, err = ed25519.GeneratePrivateKey()
	require.NoError(err)
	factory3 = auth.NewED25519Factory(priv3)
	rsender3 = auth.NewED25519Address(priv3.PublicKey())
	sender3 = codec.MustAddressBech32(nconsts.HRP, rsender3)
	log.Debug(
		"generated key",
		zap.String("addr", sender3),
		zap.String("pk", hex.EncodeToString(priv3[:])),
	)

	asset1 = []byte("as1")
	asset1Symbol = []byte("as1")
	asset1Decimals = uint8(1)
	asset2 = []byte("as2")
	asset2Symbol = []byte("ass2")
	asset2Decimals = uint8(2)
	asset3 = []byte("as3")
	asset3Symbol = []byte("as3")
	asset3Decimals = uint8(3)
	asset4 = []byte("as4")
	asset4Symbol = []byte("as4")
	asset4Decimals = uint8(0)

	// create embedded VMs
	instances = make([]instance, vms)

	gen = genesis.Default()
	gen.MinUnitPrice = fees.Dimensions{1, 1, 1, 1, 1}
	gen.MinBlockGap = 0
	gen.CustomAllocation = []*genesis.CustomAllocation{
		{
			Address: sender,
			Balance: 10_000_000,
		},
	}
	genesisBytes, err = json.Marshal(gen)
	require.NoError(err)

	networkID = uint32(1)
	subnetID := ids.GenerateTestID()
	chainID := ids.GenerateTestID()

	app := &appSender{}
	for i := range instances {
		nodeID := ids.GenerateTestNodeID()
		sk, err := bls.NewSecretKey()
		require.NoError(err)
		l, err := logFactory.Make(nodeID.String())
		require.NoError(err)
		dname, err := os.MkdirTemp("", fmt.Sprintf("%s-chainData", nodeID.String()))
		require.NoError(err)
		snowCtx := &snow.Context{
			NetworkID:      networkID,
			SubnetID:       subnetID,
			ChainID:        chainID,
			NodeID:         nodeID,
			Log:            l,
			ChainDataDir:   dname,
			Metrics:        metrics.NewOptionalGatherer(),
			PublicKey:      bls.PublicFromSecretKey(sk),
			ValidatorState: &validators.TestState{},
		}

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
		require.NoError(err)

		var hd map[string]http.Handler
		hd, err = v.CreateHandlers(context.TODO())
		require.NoError(err)

		jsonRPCServer := httptest.NewServer(hd[rpc.JSONRPCEndpoint])
		njsonRPCServer := httptest.NewServer(hd[nrpc.JSONRPCEndpoint])
		webSocketServer := httptest.NewServer(hd[rpc.WebSocketEndpoint])
		instances[i] = instance{
			chainID:             snowCtx.ChainID,
			nodeID:              snowCtx.NodeID,
			vm:                  v,
			toEngine:            toEngine,
			JSONRPCServer:       jsonRPCServer,
			NuklaiJSONRPCServer: njsonRPCServer,
			WebSocketServer:     webSocketServer,
			cli:                 rpc.NewJSONRPCClient(jsonRPCServer.URL),
			ncli:                nrpc.NewJSONRPCClient(njsonRPCServer.URL, snowCtx.NetworkID, snowCtx.ChainID),
		}

		// Force sync ready (to mimic bootstrapping from genesis)
		v.ForceReady()
	}

	// Verify genesis allocates loaded correctly (do here otherwise test may
	// check during and it will be inaccurate)
	for _, inst := range instances {
		cli := inst.ncli
		g, err := cli.Genesis(context.Background())
		require.NoError(err)

		csupply := uint64(0)
		for _, alloc := range g.CustomAllocation {
			balance, err := cli.Balance(context.Background(), alloc.Address, nconsts.Symbol)
			require.NoError(err)
			require.Equal(balance, alloc.Balance)
			csupply += alloc.Balance
		}
		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := cli.Asset(context.Background(), nconsts.Symbol, false)

		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal(name, nconsts.Name)
		require.Equal(symbol, nconsts.Symbol)
		require.Equal(decimals, uint8(nconsts.Decimals))
		require.Equal(metadata, nconsts.Name)
		require.Equal(uri, nconsts.Name)
		require.Equal(totalSupply, csupply)
		require.Equal(maxSupply, g.EmissionBalancer.MaxSupply)
		require.Equal(admin, codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress))
		require.Equal(mintActor, codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress))
		require.Equal(pauseUnpauseActor, codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress))
		require.Equal(freezeUnfreezeActor, codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress))
		require.Equal(enableDisableKYCAccountActor, codec.MustAddressBech32(nconsts.HRP, codec.EmptyAddress))
	}
	blocks = []snowman.Block{}

	app.instances = instances
	color.Blue("created %d VMs", vms)
})

var _ = ginkgo.AfterSuite(func() {
	require := require.New(ginkgo.GinkgoT())

	for _, iv := range instances {
		iv.JSONRPCServer.Close()
		iv.NuklaiJSONRPCServer.Close()
		iv.WebSocketServer.Close()
		err := iv.vm.Shutdown(context.TODO())
		require.NoError(err)
	}
})

var _ = ginkgo.Describe("[Ping]", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("can ping", func() {
		for _, inst := range instances {
			cli := inst.cli
			ok, err := cli.Ping(context.Background())
			require.NoError(err)
			require.True(ok)
		}
	})
})

var _ = ginkgo.Describe("[Network]", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("can get network", func() {
		for _, inst := range instances {
			cli := inst.cli
			networkID, subnetID, chainID, err := cli.Network(context.Background())
			require.NoError(err)
			require.Equal(networkID, uint32(1))
			require.NotEqual(subnetID, ids.Empty)
			require.NotEqual(chainID, ids.Empty)
		}
	})
})

var _ = ginkgo.Describe("[Tx Processing]", func() {
	require := require.New(ginkgo.GinkgoT())

	// Unit explanation
	//
	// bandwidth: tx size
	// compute: 5 for signature, 1 for base, 1 for transfer
	// read: 2 keys reads
	// allocate: 1 key created with 1 chunk
	// write: 2 keys modified
	transferTxUnits := fees.Dimensions{224, 7, 14, 50, 26}
	transferTxFee := uint64(321)

	ginkgo.It("get currently accepted block ID", func() {
		for _, inst := range instances {
			cli := inst.cli
			_, _, _, err := cli.Accepted(context.Background())
			require.NoError(err)
		}
	})

	var transferTxRoot *chain.Transaction
	ginkgo.It("Gossip TransferTx to a different node", func() {
		ginkgo.By("issue TransferTx", func() {
			parser, err := instances[0].ncli.Parser(context.Background())
			require.NoError(err)
			submit, transferTx, _, err := instances[0].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.Transfer{
					To:    rsender2,
					Value: 100_000, // must be more than StateLockup
				}},
				factory,
			)
			transferTxRoot = transferTx
			require.NoError(err)
			require.NoError(submit(context.Background()))
			require.Equal(instances[0].vm.Mempool().Len(context.Background()), 1)
		})

		ginkgo.By("skip duplicate", func() {
			_, err := instances[0].cli.SubmitTx(
				context.Background(),
				transferTxRoot.Bytes(),
			)
			require.Error(err)
		})

		ginkgo.By("send gossip from node 0 to 1", func() {
			err := instances[0].vm.Gossiper().Force(context.TODO())
			require.NoError(err)
		})

		ginkgo.By("skip invalid time", func() {
			tx := chain.NewTx(
				&chain.Base{
					ChainID:   instances[0].chainID,
					Timestamp: 0,
					MaxFee:    1000,
				},
				[]chain.Action{&actions.Transfer{
					To:    rsender2,
					Value: 110,
				}},
			)
			// Must do manual construction to avoid `tx.Sign` error (would fail with
			// 0 timestamp)
			msg, err := tx.Digest()
			require.NoError(err)
			auth, err := factory.Sign(msg)
			require.NoError(err)
			tx.Auth = auth
			p := codec.NewWriter(0, consts.MaxInt) // test codec growth
			require.NoError(tx.Marshal(p))
			require.NoError(p.Err())
			_, err = instances[0].cli.SubmitTx(
				context.Background(),
				p.Bytes(),
			)
			require.Error(err)
		})

		ginkgo.By("skip duplicate (after gossip, which shouldn't clear)", func() {
			_, err := instances[0].cli.SubmitTx(
				context.Background(),
				transferTxRoot.Bytes(),
			)
			require.Error(err)
		})

		ginkgo.By("receive gossip in the node 1, and signal block build", func() {
			require.NoError(instances[1].vm.Builder().Force(context.TODO()))
			<-instances[1].toEngine
		})

		ginkgo.By("build block in the node 1", func() {
			ctx := context.TODO()
			blk, err := instances[1].vm.BuildBlock(ctx)
			require.NoError(err)

			require.NoError(blk.Verify(ctx))
			require.Equal(blk.Status(), choices.Processing)

			err = instances[1].vm.SetPreference(ctx, blk.ID())
			require.NoError(err)

			require.NoError(blk.Accept(ctx))
			require.Equal(blk.Status(), choices.Accepted)
			blocks = append(blocks, blk)

			lastAccepted, err := instances[1].vm.LastAccepted(ctx)
			require.NoError(err)
			require.Equal(lastAccepted, blk.ID())

			results := blk.(*chain.StatelessBlock).Results()
			require.Len(results, 1)
			require.True(results[0].Success)
			require.Equal(results[0].Units, transferTxUnits)
			require.Equal(results[0].Fee, transferTxFee)
		})

		ginkgo.By("ensure balance is updated", func() {
			balance, err := instances[1].ncli.Balance(context.Background(), sender, nconsts.Symbol)
			require.NoError(err)
			require.Equal(balance, uint64(9_899_679))
			balance2, err := instances[1].ncli.Balance(context.Background(), sender2, nconsts.Symbol)
			require.NoError(err)
			require.Equal(balance2, uint64(100_000))
		})
	})

	ginkgo.It("ensure multiple txs work ", func() {
		ginkgo.By("transfer funds again", func() {
			parser, err := instances[1].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[1].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.Transfer{
					To:    rsender2,
					Value: 101,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))
			time.Sleep(2 * time.Second) // for replay test
			accept := expectBlk(instances[1])
			results := accept(true)
			require.Len(results, 1)
			require.True(results[0].Success)

			balance2, err := instances[1].ncli.Balance(context.Background(), sender2, nconsts.Symbol)
			require.NoError(err)
			require.Equal(balance2, uint64(100_101))
		})
	})

	ginkgo.It("Test processing block handling", func() {
		var accept, accept2 func(bool) []*chain.Result

		ginkgo.By("create processing tip", func() {
			parser, err := instances[1].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[1].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.Transfer{
					To:    rsender2,
					Value: 200,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))
			time.Sleep(2 * time.Second) // for replay test
			accept = expectBlk(instances[1])

			submit, _, _, err = instances[1].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.Transfer{
					To:    rsender2,
					Value: 201,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))
			time.Sleep(2 * time.Second) // for replay test
			accept2 = expectBlk(instances[1])
		})

		ginkgo.By("clear processing tip", func() {
			results := accept(true)
			require.Len(results, 1)
			require.True(results[0].Success)
			results = accept2(true)
			require.Len(results, 1)
			require.True(results[0].Success)
		})
	})

	ginkgo.It("ensure mempool works", func() {
		ginkgo.By("fail Gossip TransferTx to a stale node when missing previous blocks", func() {
			parser, err := instances[1].ncli.Parser(context.Background())
			require.NoError(err)
			submit, _, _, err := instances[1].cli.GenerateTransaction(
				context.Background(),
				parser,
				[]chain.Action{&actions.Transfer{
					To:    rsender2,
					Value: 203,
				}},
				factory,
			)
			require.NoError(err)
			require.NoError(submit(context.Background()))

			err = instances[1].vm.Gossiper().Force(context.TODO())
			require.NoError(err)

			// mempool in 0 should be 1 (old amount), since gossip/submit failed
			require.Equal(instances[0].vm.Mempool().Len(context.TODO()), 1)
		})
	})

	ginkgo.It("ensure unprocessed tip and replay protection works", func() {
		ginkgo.By("import accepted blocks to instance 2", func() {
			ctx := context.TODO()

			require.Equal(blocks[0].Height(), uint64(1))

			n := instances[2]
			blk1, err := n.vm.ParseBlock(ctx, blocks[0].Bytes())
			require.NoError(err)
			err = blk1.Verify(ctx)
			require.NoError(err)

			// Parse tip
			blk2, err := n.vm.ParseBlock(ctx, blocks[1].Bytes())
			require.NoError(err)
			blk3, err := n.vm.ParseBlock(ctx, blocks[2].Bytes())
			require.NoError(err)

			// Verify tip
			err = blk2.Verify(ctx)
			require.NoError(err)
			err = blk3.Verify(ctx)
			require.NoError(err)

			// Check if tx from old block would be considered a repeat on processing tip
			tx := blk2.(*chain.StatelessBlock).Txs[0]
			sblk3 := blk3.(*chain.StatelessBlock)
			sblk3t := sblk3.Timestamp().UnixMilli()
			ok, err := sblk3.IsRepeat(ctx, sblk3t-n.vm.Rules(sblk3t).GetValidityWindow(), []*chain.Transaction{tx}, set.NewBits(), false)
			require.NoError(err)
			require.Equal(ok.Len(), 1)

			// Accept tip
			err = blk1.Accept(ctx)
			require.NoError(err)
			err = blk2.Accept(ctx)
			require.NoError(err)
			err = blk3.Accept(ctx)
			require.NoError(err)

			// Parse another
			blk4, err := n.vm.ParseBlock(ctx, blocks[3].Bytes())
			require.NoError(err)
			err = blk4.Verify(ctx)
			require.NoError(err)
			err = blk4.Accept(ctx)
			require.NoError(err)

			// Check if tx from old block would be considered a repeat on accepted tip
			time.Sleep(2 * time.Second)
			require.Equal(n.vm.IsRepeat(ctx, []*chain.Transaction{tx}, set.NewBits(), false).Len(), 1)
		})
	})

	ginkgo.It("processes valid index transactions (w/block listening)", func() {
		// Clear previous txs on instance 0
		accept := expectBlk(instances[0])
		accept(false) // don't care about results

		// Subscribe to blocks
		cli, err := rpc.NewWebSocketClient(instances[0].WebSocketServer.URL, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
		require.NoError(err)
		require.NoError(cli.RegisterBlocks())

		// Wait for message to be sent
		time.Sleep(2 * pubsub.MaxMessageWait)

		// Fetch balances
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)

		// Send tx
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		transfer := []chain.Action{&actions.Transfer{
			To:    auth.NewED25519Address(other.PublicKey()),
			Value: 1,
		}}

		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			transfer,
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Read item from connection
		blk, lresults, prices, err := cli.ListenBlock(context.TODO(), parser)
		require.NoError(err)
		require.Len(blk.Txs, 1)
		tx := blk.Txs[0].Actions[0].(*actions.Transfer)
		require.Equal(tx.Asset, ids.Empty)
		require.Equal(tx.Value, uint64(1))
		require.Equal(lresults, results)
		require.Equal(prices, fees.Dimensions{1, 1, 1, 1, 1})

		// Check balance modifications are correct
		balancea, err := instances[0].ncli.Balance(context.TODO(), sender, nconsts.Symbol)
		require.NoError(err)
		require.Equal(balance, balancea+lresults[0].Fee+1)

		// Close connection when done
		require.NoError(cli.Close())
	})

	ginkgo.It("processes valid index transactions (w/streaming verification)", func() {
		// Create streaming client
		cli, err := rpc.NewWebSocketClient(instances[0].WebSocketServer.URL, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
		require.NoError(err)

		// Create tx
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		transfer := []chain.Action{&actions.Transfer{
			To:    auth.NewED25519Address(other.PublicKey()),
			Value: 1,
		}}
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		_, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			transfer,
			factory,
		)
		require.NoError(err)

		// Submit tx and accept block
		require.NoError(cli.RegisterTx(tx))

		// Wait for message to be sent
		time.Sleep(2 * pubsub.MaxMessageWait)

		for instances[0].vm.Mempool().Len(context.TODO()) == 0 {
			// We need to wait for mempool to be populated because issuance will
			// return as soon as bytes are on the channel.
			hutils.Outf("{{yellow}}waiting for mempool to return non-zero txs{{/}}\n")
			time.Sleep(500 * time.Millisecond)
		}
		require.NoError(err)
		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Read decision from connection
		txID, dErr, result, err := cli.ListenTx(context.TODO())
		require.NoError(err)
		require.Equal(txID, tx.ID())
		require.Nil(dErr)
		require.True(result.Success)
		require.Equal(result, results[0])

		// Close connection when done
		require.NoError(cli.Close())
	})

	ginkgo.It("transfer an asset with a memo", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.Transfer{
				To:    auth.NewED25519Address(other.PublicKey()),
				Value: 10,
				Memo:  []byte("hello"),
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.True(result.Success)
	})

	ginkgo.It("transfer an asset with large memo", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1001,
			},
			[]chain.Action{&actions.Transfer{
				To:    auth.NewED25519Address(other.PublicKey()),
				Value: 10,
				Memo:  make([]byte, 1000),
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// too large)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "size is larger than limit")
	})

	ginkgo.It("mint a fungible asset that doesn't exist", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		assetID := ids.GenerateTestID()
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetFT{
				To:    auth.NewED25519Address(other.PublicKey()),
				Asset: assetID,
				Value: 10,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "asset missing")

		exists, _, _, _, _, _, _, _, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), assetID.String(), false)
		require.NoError(err)
		require.False(exists)
	})

	ginkgo.It("mint a non-fungible asset that doesn't exist", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		assetID := ids.GenerateTestID()
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetNFT{
				To:       auth.NewED25519Address(other.PublicKey()),
				Asset:    assetID,
				UniqueID: 0,
				URI:      []byte("uri"),
				Metadata: []byte("metadata"),
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "asset missing")

		exists, _, _, _, _, _, _, _, _, _, _, _, _, _, err := instances[0].ncli.Asset(context.TODO(), assetID.String(), false)
		require.NoError(err)
		require.False(exists)
	})

	ginkgo.It("create a new asset (no metadata)", func() {
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1001,
			},
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetFungibleTokenID,
				Name:                         []byte("n00"),
				Symbol:                       []byte("s00"),
				Decimals:                     0,
				Metadata:                     nil,
				URI:                          []byte("uri"),
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// too large)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "Bytes field is not populated")
	})

	ginkgo.It("create a new asset (no symbol)", func() {
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1001,
			},
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetFungibleTokenID,
				Name:                         []byte("n00"),
				Symbol:                       nil,
				Decimals:                     0,
				Metadata:                     []byte("m00"),
				URI:                          []byte("u00"),
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// too large)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "Bytes field is not populated")
	})

	ginkgo.It("create asset with too long of metadata", func() {
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1000,
			},
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetFungibleTokenID,
				Name:                         []byte("n00"),
				Symbol:                       []byte("s00"),
				Decimals:                     0,
				Metadata:                     make([]byte, actions.MaxMetadataSize*2),
				URI:                          []byte("u00"),
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// too large)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "size is larger than limit")
	})

	ginkgo.It("create a new non-fungible asset (decimals is greater than 0)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetNonFungibleTokenID,
				Name:                         []byte("n00"),
				Symbol:                       []byte("s00"),
				Decimals:                     1,
				Metadata:                     []byte("m00"),
				URI:                          []byte("u00"),
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "decimal is invalid")
	})

	ginkgo.It("create a new fungible asset (simple metadata)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetFungibleTokenID,
				Name:                         asset1,
				Symbol:                       asset1Symbol,
				Decimals:                     asset1Decimals,
				Metadata:                     asset1,
				URI:                          asset1,
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset1ID = chain.CreateActionID(tx.ID(), 0)
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)

		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1Symbol)
		require.Equal(decimals, asset1Decimals)
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Zero(totalSupply)
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("create a new non-fungible asset (simple metadata)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetNonFungibleTokenID,
				Name:                         asset2,
				Symbol:                       asset2Symbol,
				Decimals:                     0,
				Metadata:                     asset2,
				URI:                          asset2,
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset2ID = chain.CreateActionID(tx.ID(), 0)
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)

		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
		require.Equal([]byte(name), asset2)
		require.Equal([]byte(symbol), asset2Symbol)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset2)
		require.Equal([]byte(uri), asset2)
		require.Zero(totalSupply)
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("create a new dataset asset (simple metadata)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetDatasetTokenID,
				Name:                         asset3,
				Symbol:                       asset3Symbol,
				Decimals:                     0,
				Metadata:                     asset3,
				URI:                          asset3,
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)

		require.True(results[0].Success)

		asset3ID = chain.CreateActionID(tx.ID(), 0)
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset3ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))
		nftID := nchain.GenerateID(asset3ID, 0)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))
		balance, err = instances[0].ncli.Balance(context.TODO(), sender2, asset3ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset3ID.String(), false)

		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
		require.Equal([]byte(name), asset3)
		require.Equal([]byte(symbol), asset3Symbol)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset3)
		require.Equal([]byte(uri), asset3)
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)

		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, asset3ID.String())
		require.Equal(uniqueID, uint64(0))
		require.Equal([]byte(uri), asset3)
		require.Equal([]byte(metadata), asset3)
		require.Equal(owner, sender)

		// TODO: Get storage.GetAssetNFTsByCollection
	})

	ginkgo.It("update an asset that doesn't exist", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.UpdateAsset{
				Asset:    ids.GenerateTestID(),
				Name:     asset4,
				Symbol:   asset4Symbol,
				Metadata: asset4,
				URI:      asset4,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "asset not found")
	})

	ginkgo.It("update an asset(no field is updated)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.UpdateAsset{
				Asset: asset1ID,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "must update at least one field")
	})

	ginkgo.It("update an existing asset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.UpdateAsset{
				Asset:     asset3ID,
				MaxSupply: uint64(100000),
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset3ID.String(), false)

		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
		require.Equal([]byte(name), asset3)
		require.Equal([]byte(symbol), asset3Symbol)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset3)
		require.Equal([]byte(uri), asset3)
		require.Equal(totalSupply, uint64(1))
		require.Equal(maxSupply, uint64(100000))
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("mint a new fungible asset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetFT{
				To:    rsender2,
				Asset: asset1ID,
				Value: 15,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(15))
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1Symbol)
		require.Equal(decimals, asset1Decimals)
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Equal(totalSupply, uint64(15))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("mint a new non-fungible asset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetNFT{
				To:       rsender2,
				Asset:    asset2ID,
				UniqueID: 0,
				URI:      []byte("uri"),
				Metadata: []byte("metadata"),
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset2ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))
		nftID := nchain.GenerateID(asset2ID, 0)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender2, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
		require.Equal([]byte(name), asset2)
		require.Equal([]byte(symbol), asset2Symbol)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset2)
		require.Equal([]byte(uri), asset2)
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)

		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, asset2ID.String())
		require.Equal(uniqueID, uint64(0))
		require.Equal([]byte(uri), []byte("uri"))
		require.Equal([]byte(metadata), []byte("metadata"))
		require.Equal(owner, sender2)
	})

	ginkgo.It("mint fungible asset from wrong owner", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetFT{
				To:    auth.NewED25519Address(other.PublicKey()),
				Asset: asset1ID,
				Value: 10,
			}},
			factory2,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]

		require.False(result.Success)
		require.Contains(string(result.Error), "wrong mint actor")

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1Symbol)
		require.Equal(decimals, asset1Decimals)
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Equal(totalSupply, uint64(15))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("mint non-fungible asset from wrong owner", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetNFT{
				To:       auth.NewED25519Address(other.PublicKey()),
				Asset:    asset2ID,
				UniqueID: 1,
				URI:      []byte("uri"),
				Metadata: []byte("metadata"),
			}},
			factory2,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]

		require.False(result.Success)
		require.Contains(string(result.Error), "wrong mint actor")

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
		require.Equal([]byte(name), asset2)
		require.Equal([]byte(symbol), asset2Symbol)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset2)
		require.Equal([]byte(uri), asset2)
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("burn new fungible asset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.BurnAssetFT{
				Asset: asset1ID,
				Value: 5,
			}},
			factory2,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(10))
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1Symbol)
		require.Equal(decimals, asset1Decimals)
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Equal(totalSupply, uint64(10))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("burn new non-fungible asset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		nftID := nchain.GenerateID(asset2ID, 0)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.BurnAssetNFT{
				Asset: asset2ID,
				NftID: nftID,
			}},
			factory2,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset2ID.String())
		require.NoError(err)
		require.Zero(balance)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset2ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetNonFungibleTokenDesc)
		require.Equal([]byte(name), asset2)
		require.Equal([]byte(symbol), asset2Symbol)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset2)
		require.Equal([]byte(uri), asset2)
		require.Equal(totalSupply, uint64(0))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("burn missing asset", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.BurnAssetFT{
				Asset: asset1ID,
				Value: 10,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "invalid balance")

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1Symbol)
		require.Equal(decimals, asset1Decimals)
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Equal(totalSupply, uint64(10))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("rejects empty mint", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1000,
			},
			[]chain.Action{&actions.MintAssetFT{
				To:    auth.NewED25519Address(other.PublicKey()),
				Asset: asset1ID,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// bad codec)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "Uint64 field is not populated")
	})

	ginkgo.It("reject max mint", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetFT{
				To:    rsender2,
				Asset: asset1ID,
				Value: consts.MaxUint64,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "overflow")

		balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(10))
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Zero(balance)

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetFungibleTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1Symbol)
		require.Equal(decimals, asset1Decimals)
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Equal(totalSupply, uint64(10))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)
	})

	ginkgo.It("rejects mint of native token", func() {
		other, err := ed25519.GeneratePrivateKey()
		require.NoError(err)
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1000,
			},
			[]chain.Action{&actions.MintAssetFT{
				To:    auth.NewED25519Address(other.PublicKey()),
				Value: 10,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// bad codec)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "ID field is not populated")
	})

	ginkgo.It("mints another new asset (to self)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetFungibleTokenID,
				Name:                         asset2,
				Symbol:                       asset2Symbol,
				Decimals:                     asset2Decimals,
				Metadata:                     asset2,
				URI:                          asset2,
				MaxSupply:                    0,
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)
		asset2ID = chain.CreateActionID(tx.ID(), 0)

		submit, _, _, err = instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetFT{
				To:    rsender,
				Asset: asset2ID,
				Value: 10,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[0])
		results = accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset2ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(10))
	})

	ginkgo.It("mints another new asset (to self) on another account", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetFungibleTokenID,
				Name:                         asset3,
				Symbol:                       asset3Symbol,
				Decimals:                     asset3Decimals,
				Metadata:                     asset3,
				URI:                          asset3,
				MaxSupply:                    0,
				MintActor:                    rsender2,
				PauseUnpauseActor:            rsender2,
				FreezeUnfreezeActor:          rsender2,
				EnableDisableKYCAccountActor: rsender2,
			}},
			factory2,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)
		asset3ID = chain.CreateActionID(tx.ID(), 0)

		submit, _, _, err = instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.MintAssetFT{
				To:    rsender2,
				Asset: asset3ID,
				Value: 10,
			}},
			factory2,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[0])
		results = accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance, err := instances[0].ncli.Balance(context.TODO(), sender2, asset3ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(10))
	})

	// Use new instance to make balance checks easier (note, instances are in different
	// states and would never agree)
	ginkgo.It("transfer to multiple accounts in a single tx", func() {
		parser, err := instances[3].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[3].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{
				&actions.Transfer{
					To:    rsender2,
					Value: 10000,
				},
				&actions.Transfer{
					To:    rsender3,
					Value: 5000,
				},
			},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		time.Sleep(2 * time.Second) // for replay test
		accept := expectBlk(instances[3])
		results := accept(true)
		require.Len(results, 1)
		require.True(results[0].Success)

		balance2, err := instances[3].ncli.Balance(context.Background(), sender2, nconsts.Symbol)
		require.NoError(err)
		require.Equal(balance2, uint64(10000))

		balance3, err := instances[3].ncli.Balance(context.Background(), sender3, nconsts.Symbol)
		require.NoError(err)
		require.Equal(balance3, uint64(5000))
	})

	ginkgo.It("create and mint multiple of fungible assets in a single tx", func() {
		// Create asset
		parser, err := instances[3].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[3].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{
				&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         asset1,
					Symbol:                       asset1Symbol,
					Decimals:                     asset1Decimals,
					Metadata:                     asset1,
					URI:                          asset1,
					MaxSupply:                    0,
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
				},
				&actions.CreateAsset{
					AssetType:                    nconsts.AssetFungibleTokenID,
					Name:                         asset2,
					Symbol:                       asset2Symbol,
					Decimals:                     asset2Decimals,
					Metadata:                     asset2,
					URI:                          asset2,
					MaxSupply:                    0,
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
				},
			},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[3])
		results := accept(true)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset1ID = chain.CreateActionID(tx.ID(), 0)
		asset2ID = chain.CreateActionID(tx.ID(), 1)

		// Mint multiple
		submit, _, _, err = instances[3].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{
				&actions.MintAssetFT{
					To:    rsender2,
					Asset: asset1ID,
					Value: 10,
				},
				&actions.MintAssetFT{
					To:    rsender2,
					Asset: asset2ID,
					Value: 10,
				},
			},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[3])
		results = accept(true)
		require.Len(results, 1)
		require.True(results[0].Success)

		// check sender2 assets
		balance1, err := instances[3].ncli.Balance(context.TODO(), sender2, asset1ID.String())
		require.NoError(err)
		require.Equal(balance1, uint64(10))

		balance2, err := instances[3].ncli.Balance(context.TODO(), sender2, asset2ID.String())
		require.NoError(err)
		require.Equal(balance2, uint64(10))
	})

	ginkgo.It("create and mint multiple of non-fungible assets in a single tx", func() {
		// Create asset
		parser, err := instances[3].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[3].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{
				&actions.CreateAsset{
					AssetType:                    nconsts.AssetNonFungibleTokenID,
					Name:                         asset1,
					Symbol:                       asset1Symbol,
					Decimals:                     0,
					Metadata:                     asset1,
					URI:                          asset1,
					MaxSupply:                    0,
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
				},
				&actions.CreateAsset{
					AssetType:                    nconsts.AssetNonFungibleTokenID,
					Name:                         asset2,
					Symbol:                       asset2Symbol,
					Decimals:                     0,
					Metadata:                     asset2,
					URI:                          asset2,
					MaxSupply:                    0,
					MintActor:                    rsender,
					PauseUnpauseActor:            rsender,
					FreezeUnfreezeActor:          rsender,
					EnableDisableKYCAccountActor: rsender,
				},
			},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[3])
		results := accept(true)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset1ID = chain.CreateActionID(tx.ID(), 0)
		asset2ID = chain.CreateActionID(tx.ID(), 1)

		// Mint multiple
		submit, _, _, err = instances[3].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{
				&actions.MintAssetNFT{
					To:       rsender2,
					Asset:    asset1ID,
					UniqueID: 0,
					URI:      []byte("uri1"),
					Metadata: []byte("metadata1"),
				},
				&actions.MintAssetNFT{
					To:       rsender2,
					Asset:    asset2ID,
					UniqueID: 1,
					URI:      []byte("uri2"),
					Metadata: []byte("metadata2"),
				},
			},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[3])
		results = accept(true)
		require.Len(results, 1)
		require.True(results[0].Success)

		// check sender2 assets
		balance1, err := instances[3].ncli.Balance(context.TODO(), sender2, asset1ID.String())
		require.NoError(err)
		require.Equal(balance1, uint64(1))
		/* 		nftID := nchain.GenerateID(asset2ID, 0)
		balance1, err = instances[3].ncli.Balance(context.TODO(), sender2, nftID.String())
		require.NoError(err)
		require.Equal(balance1, uint64(1)) */

		balance2, err := instances[3].ncli.Balance(context.TODO(), sender2, asset2ID.String())
		require.NoError(err)
		require.Equal(balance2, uint64(1))
		/* 	nftID = nchain.GenerateID(asset2ID, 1)
		balance2, err = instances[0].ncli.Balance(context.TODO(), sender2, nftID.String())
		require.NoError(err)
		require.Equal(balance2, uint64(1)) */
	})

	ginkgo.It("create a new dataset (no metadata)", func() {
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1001,
			},
			[]chain.Action{&actions.CreateDataset{
				AssetID:            ids.Empty,
				Name:               []byte("n00"),
				Description:        []byte("s00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				Metadata:           nil,
				IsCommunityDataset: false,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// too large)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "Bytes field is not populated")
	})

	ginkgo.It("create dataset with too long of metadata", func() {
		tx := chain.NewTx(
			&chain.Base{
				ChainID:   instances[0].chainID,
				Timestamp: hutils.UnixRMilli(-1, 5*consts.MillisecondsPerSecond),
				MaxFee:    1000,
			},
			[]chain.Action{&actions.CreateDataset{
				AssetID:            ids.Empty,
				Name:               []byte("n00"),
				Description:        []byte("d00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				Metadata:           make([]byte, actions.MaxDatasetMetadataSize*2),
				IsCommunityDataset: false,
			}},
		)
		// Must do manual construction to avoid `tx.Sign` error (would fail with
		// too large)
		msg, err := tx.Digest()
		require.NoError(err)
		auth, err := factory.Sign(msg)
		require.NoError(err)
		tx.Auth = auth
		p := codec.NewWriter(0, consts.MaxInt) // test codec growth
		require.NoError(tx.Marshal(p))
		require.NoError(p.Err())
		_, err = instances[0].cli.SubmitTx(
			context.Background(),
			p.Bytes(),
		)
		require.ErrorContains(err, "size is larger than limit")
	})

	ginkgo.It("create a new dataset (solo contributor dataset)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateDataset{
				AssetID:            ids.Empty,
				Name:               asset1,
				Description:        []byte("d00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				Metadata:           asset1,
				IsCommunityDataset: false,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset1ID = chain.CreateActionID(tx.ID(), 0)

		// Check dataset info
		exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(description), []byte("d00"))
		require.Equal([]byte(categories), []byte("c00"))
		require.Equal([]byte(licenseName), []byte("l00"))
		require.Equal([]byte(licenseSymbol), []byte("ls00"))
		require.Equal([]byte(licenseURL), []byte("lu00"))
		require.Equal([]byte(metadata), asset1)
		require.False(isCommunityDataset)
		require.False(onSale)
		require.Equal(baseAsset, ids.Empty.String())
		require.Zero(basePrice)
		require.Equal(revenueModelDataShare, uint8(100))
		require.Equal(revenueModelMetadataShare, uint8(0))
		require.Equal(revenueModelDataOwnerCut, uint8(100))
		require.Equal(revenueModelMetadataOwnerCut, uint8(0))
		require.Equal(owner, sender)

		// Check asset info
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), []byte("d00"))
		require.Equal([]byte(uri), []byte("d00"))
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)

		// Check NFT info
		nftID := nchain.GenerateID(asset1ID, 0)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, asset1ID.String())
		require.Equal(uniqueID, uint64(0))
		require.Equal([]byte(uri), []byte("d00"))
		require.Equal([]byte(metadata), []byte("d00"))
		require.Equal(owner, sender)
	})

	ginkgo.It("create a new dataset (community contributor dataset)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateDataset{
				AssetID:            ids.Empty,
				Name:               asset1,
				Description:        []byte("d00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				Metadata:           asset1,
				IsCommunityDataset: true,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset1ID = chain.CreateActionID(tx.ID(), 0)

		// Check dataset info
		exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(description), []byte("d00"))
		require.Equal([]byte(categories), []byte("c00"))
		require.Equal([]byte(licenseName), []byte("l00"))
		require.Equal([]byte(licenseSymbol), []byte("ls00"))
		require.Equal([]byte(licenseURL), []byte("lu00"))
		require.Equal([]byte(metadata), asset1)
		require.True(isCommunityDataset)
		require.False(onSale)
		require.Equal(baseAsset, ids.Empty.String())
		require.Zero(basePrice)
		require.Equal(revenueModelDataShare, uint8(100))
		require.Equal(revenueModelMetadataShare, uint8(0))
		require.Equal(revenueModelDataOwnerCut, uint8(10))
		require.Equal(revenueModelMetadataOwnerCut, uint8(0))
		require.Equal(owner, sender)

		// Check asset info
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), []byte("d00"))
		require.Equal([]byte(uri), []byte("d00"))
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)

		// Check NFT info
		nftID := nchain.GenerateID(asset1ID, 0)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, asset1ID.String())
		require.Equal(uniqueID, uint64(0))
		require.Equal([]byte(uri), []byte("d00"))
		require.Equal([]byte(metadata), []byte("d00"))
		require.Equal(owner, sender)
	})

	ginkgo.It("create a new dataset (after an asset is already created)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, tx, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateAsset{
				AssetType:                    nconsts.AssetDatasetTokenID,
				Name:                         asset1,
				Symbol:                       asset1Symbol,
				Decimals:                     0,
				Metadata:                     asset1,
				URI:                          asset1,
				MaxSupply:                    uint64(0),
				MintActor:                    rsender,
				PauseUnpauseActor:            rsender,
				FreezeUnfreezeActor:          rsender,
				EnableDisableKYCAccountActor: rsender,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		asset1ID = chain.CreateActionID(tx.ID(), 0)
		parser, err = instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err = instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.CreateDataset{
				AssetID:            asset1ID,
				Name:               asset1,
				Description:        []byte("d00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				Metadata:           asset1,
				IsCommunityDataset: false,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept = expectBlk(instances[0])
		results = accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Check dataset info
		exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, onSale, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadataOwnerCut, owner, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(description), []byte("d00"))
		require.Equal([]byte(categories), []byte("c00"))
		require.Equal([]byte(licenseName), []byte("l00"))
		require.Equal([]byte(licenseSymbol), []byte("ls00"))
		require.Equal([]byte(licenseURL), []byte("lu00"))
		require.Equal([]byte(metadata), asset1)
		require.False(isCommunityDataset)
		require.False(onSale)
		require.Equal(baseAsset, ids.Empty.String())
		require.Zero(basePrice)
		require.Equal(revenueModelDataShare, uint8(100))
		require.Equal(revenueModelMetadataShare, uint8(0))
		require.Equal(revenueModelDataOwnerCut, uint8(100))
		require.Equal(revenueModelMetadataOwnerCut, uint8(0))
		require.Equal(owner, sender)

		// Check asset info
		balance, err := instances[0].ncli.Balance(context.TODO(), sender, asset1ID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := instances[0].ncli.Asset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(assetType, nconsts.AssetDatasetTokenDesc)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(symbol), asset1)
		require.Equal(decimals, uint8(0))
		require.Equal([]byte(metadata), asset1)
		require.Equal([]byte(uri), asset1)
		require.Equal(totalSupply, uint64(1))
		require.Zero(maxSupply)
		require.Equal(admin, sender)
		require.Equal(mintActor, sender)
		require.Equal(pauseUnpauseActor, sender)
		require.Equal(freezeUnfreezeActor, sender)
		require.Equal(enableDisableKYCAccountActor, sender)

		// Check NFT info
		nftID := nchain.GenerateID(asset1ID, 0)
		balance, err = instances[0].ncli.Balance(context.TODO(), sender, nftID.String())
		require.NoError(err)
		require.Equal(balance, uint64(1))

		exists, collectionID, uniqueID, uri, metadata, owner, err := instances[0].ncli.AssetNFT(context.TODO(), nftID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal(collectionID, asset1ID.String())
		require.Equal(uniqueID, uint64(0))
		require.Equal([]byte(uri), asset1)
		require.Equal([]byte(metadata), asset1)
		require.Equal(owner, sender)
	})

	ginkgo.It("update a dataset that doesn't exist", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.UpdateDataset{
				Dataset:            ids.GenerateTestID(),
				Name:               []byte("n00"),
				Description:        []byte("d00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				IsCommunityDataset: false,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "dataset not found")
	})

	ginkgo.It("update a dataset(no field is updated)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.UpdateDataset{
				Dataset:            asset1ID,
				Name:               asset1,
				Description:        []byte("d00"),
				Categories:         []byte("c00"),
				LicenseName:        []byte("l00"),
				LicenseSymbol:      []byte("ls00"),
				LicenseURL:         []byte("lu00"),
				IsCommunityDataset: false,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		result := results[0]
		require.False(result.Success)
		require.Contains(string(result.Error), "must update at least one field")
	})

	ginkgo.It("update an existing dataset(convert solo contributor to community dataset)", func() {
		parser, err := instances[0].ncli.Parser(context.Background())
		require.NoError(err)
		submit, _, _, err := instances[0].cli.GenerateTransaction(
			context.Background(),
			parser,
			[]chain.Action{&actions.UpdateDataset{
				Dataset:            asset1ID,
				Name:               asset1,
				Description:        []byte("d00-updated"),
				Categories:         []byte("c00-updated"),
				LicenseName:        []byte("l00-updated"),
				LicenseSymbol:      []byte("ls00up"),
				LicenseURL:         []byte("lu00-updated"),
				IsCommunityDataset: true,
			}},
			factory,
		)
		require.NoError(err)
		require.NoError(submit(context.Background()))

		accept := expectBlk(instances[0])
		results := accept(false)
		require.Len(results, 1)
		require.True(results[0].Success)

		// Check dataset info
		exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, _, _, _, _, _, revenueModelDataOwnerCut, _, _, err := instances[0].ncli.Dataset(context.TODO(), asset1ID.String(), false)
		require.NoError(err)
		require.True(exists)
		require.Equal([]byte(name), asset1)
		require.Equal([]byte(description), []byte("d00-updated"))
		require.Equal([]byte(categories), []byte("c00-updated"))
		require.Equal([]byte(licenseName), []byte("l00-updated"))
		require.Equal([]byte(licenseSymbol), []byte("ls00up"))
		require.Equal([]byte(licenseURL), []byte("lu00-updated"))
		require.Equal([]byte(metadata), asset1)
		require.True(isCommunityDataset)
		require.Equal(revenueModelDataOwnerCut, uint8(10))
	})
})

func expectBlk(i instance) func(bool) []*chain.Result {
	require := require.New(ginkgo.GinkgoT())

	ctx := context.TODO()

	// manually signal ready
	require.NoError(i.vm.Builder().Force(ctx))
	// manually ack ready sig as in engine
	<-i.toEngine

	blk, err := i.vm.BuildBlock(ctx)
	require.NoError(err)
	require.NotNil(blk)

	require.NoError(blk.Verify(ctx))
	require.Equal(blk.Status(), choices.Processing)

	err = i.vm.SetPreference(ctx, blk.ID())
	require.NoError(err)

	return func(add bool) []*chain.Result {
		require.NoError(blk.Accept(ctx))
		require.Equal(blk.Status(), choices.Accepted)

		if add {
			blocks = append(blocks, blk)
		}

		lastAccepted, err := i.vm.LastAccepted(ctx)
		require.NoError(err)
		require.Equal(lastAccepted, blk.ID())
		return blk.(*chain.StatelessBlock).Results()
	}
}

var _ common.AppSender = &appSender{}

type appSender struct {
	next      int
	instances []instance
}

func (app *appSender) SendAppGossip(ctx context.Context, _ common.SendConfig, appGossipBytes []byte) error {
	n := len(app.instances)
	sender := app.instances[app.next].nodeID
	app.next++
	app.next %= n
	return app.instances[app.next].vm.AppGossip(ctx, sender, appGossipBytes)
}

func (*appSender) SendAppRequest(context.Context, set.Set[ids.NodeID], uint32, []byte) error {
	return nil
}

func (*appSender) SendAppError(context.Context, ids.NodeID, uint32, int32, string) error {
	return nil
}

func (*appSender) SendAppResponse(context.Context, ids.NodeID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppRequest(context.Context, ids.ID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppResponse(context.Context, ids.ID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppError(context.Context, ids.ID, uint32, int32, string) error {
	return nil
}
