// tx_processing_test.go
package integration

import (
	"context"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
)

var _ = ginkgo.Describe("tx_processing", func() {
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
			require.Equal(balance, uint64(9_999_999_899_679))
			balance2, err := instances[1].ncli.Balance(context.Background(), sender2, nconsts.Symbol)
			require.NoError(err)
			require.Equal(balance2, uint64(10_000_000_100_000))
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
			require.Equal(balance2, uint64(10_000_000_100_101))
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
		require.Equal(prices, fees.Dimensions{0x1, 0x1, 0x1, 0x1, 0x1})

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

	ginkgo.It("transfer an asset", func() {
		ginkgo.By("transfer an asset with a memo", func() {
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

		ginkgo.By("transfer an asset with large memo", func() {
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

		// Use new instance to make balance checks easier (note, instances are in different
		// states and would never agree)
		ginkgo.By("transfer to multiple accounts in a single tx", func() {
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
			require.Equal(balance2, uint64(10_000_000_010_000))

			balance3, err := instances[3].ncli.Balance(context.Background(), sender3, nconsts.Symbol)
			require.NoError(err)
			require.Equal(balance3, uint64(10_000_000_005_000))
		})
	})

})
