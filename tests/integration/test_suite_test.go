// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// integration_test.go
package integration

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
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/vm"

	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/controller"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/marketplace"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

type instance struct {
	chainID             ids.ID
	nodeID              ids.NodeID
	vm                  *vm.VM
	marketplace         marketplace.Hub
	toEngine            chan common.Message
	JSONRPCServer       *httptest.Server
	NuklaiJSONRPCServer *httptest.Server
	WebSocketServer     *httptest.Server
	cli                 *rpc.JSONRPCClient // clients for embedded VMs
	ncli                *nrpc.JSONRPCClient
}

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

	dataset1ID     ids.ID
	marketplace1ID ids.ID

	// when used with embedded VMs
	genesisBytes []byte
	instances    []instance
	blocks       []snowman.Block

	networkID uint32
	gen       *genesis.Genesis
)

type MockController struct {
	logger logging.Logger
}

func (mc *MockController) Logger() logging.Logger {
	return mc.logger
}

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
			Balance: 10_000_000_000_000,
		},
		{
			Address: sender2,
			Balance: 10_000_000_000_000,
		},
		{
			Address: sender3,
			Balance: 10_000_000_000_000,
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
		mockLogger := logging.NewLogger("", logging.NewWrappedCore(logging.Info, logging.Discard, logging.Plain.ConsoleEncoder()))
		mockController := &MockController{logger: mockLogger}
		instances[i] = instance{
			chainID:             snowCtx.ChainID,
			nodeID:              snowCtx.NodeID,
			vm:                  v,
			marketplace:         marketplace.NewMarketplace(mockController, v),
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

func expectBlk(i instance) func(bool) []*chain.Result {
	require := require.New(ginkgo.GinkgoT())

	ctx := context.TODO()

	require.NoError(i.vm.Builder().Force(ctx))
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
