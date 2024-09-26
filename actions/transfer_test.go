// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"math"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/chain"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"
)

func TestTransferAction(t *testing.T) {
	addr := codectest.NewRandomAddress()
	assetID := ids.GenerateTestID()
	nftID := chain.GenerateIDWithIndex(assetID, 0)

	tests := []chaintest.ActionTest{
		{
			Name:  "ZeroTransfer",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    codec.EmptyAddress,
				Value: 0,
			},
			ExpectedErr: ErrOutputValueZero,
		},
		{
			Name:  "InvalidAddress",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    codec.EmptyAddress,
				Value: 1,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: storage.ErrInvalidAddress,
		},
		{
			Name:  "NotEnoughBalance",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    codec.EmptyAddress,
				Value: 1,
			},
			State: func() state.Mutable {
				s := chaintest.NewInMemoryStore()
				_, err := storage.AddBalance(
					context.Background(),
					s,
					codec.EmptyAddress,
					ids.Empty,
					0,
					true,
				)
				require.NoError(t, err)
				return s
			}(),
			ExpectedErr: storage.ErrInvalidBalance,
		},
		{
			Name:  "OverflowBalance",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    codec.EmptyAddress,
				Value: math.MaxUint64,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetBalance(context.Background(), store, codec.EmptyAddress, ids.Empty, 1))
				return store
			}(),
			ExpectedErr: storage.ErrInvalidBalance,
		},
		{
			Name:  "MemoSizeExceeded",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    addr,
				Value: 1,
				Memo:  make([]byte, MaxMemoSize+1),
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrOutputMemoTooLarge,
		},
		{
			Name:  "SelfTransfer",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    codec.EmptyAddress,
				Value: 1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetBalance(context.Background(), store, codec.EmptyAddress, ids.Empty, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				balance, err := storage.GetBalance(ctx, store, codec.EmptyAddress, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, balance, uint64(1))
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "SimpleTransfer",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:    addr,
				Value: 1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetBalance(context.Background(), store, codec.EmptyAddress, ids.Empty, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetBalance(ctx, store, codec.EmptyAddress, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "EmptyAssetIDTransfer",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:      addr,
				Value:   10,
				AssetID: ids.Empty, // Transferring native asset
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetBalance(context.Background(), store, codec.EmptyAddress, ids.Empty, 10))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, uint64(10), receiverBalance)
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 10,
			},
		},
		{
			Name:  "ValidFTTransfer",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:      addr,
				Value:   1,
				AssetID: assetID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetBalance(context.Background(), store, codec.EmptyAddress, assetID, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check asset balance for addr
				balance, err := storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "ValidNFTTransfer",
			Actor: codec.EmptyAddress,
			Action: &Transfer{
				To:      addr,
				Value:   1,
				AssetID: nftID,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set initial ownership of the NFT to the actor
				collectionID, uniqueID, uri, metadata, owner := assetID, uint64(0), "uri", "metadata", codec.EmptyAddress
				require.NoError(t, storage.SetAssetNFT(context.Background(), store, collectionID, uniqueID, nftID, []byte(uri), []byte(metadata), owner))
				require.NoError(t, storage.SetBalance(context.Background(), store, owner, nftID, 1))
				require.NoError(t, storage.SetBalance(context.Background(), store, owner, collectionID, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check NFT balance for addr
				balance, err := storage.GetBalance(ctx, store, addr, nftID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
				// Check collectionID balance for addr
				balance, err = storage.GetBalance(ctx, store, addr, assetID)
				require.NoError(t, err)
				require.Equal(t, uint64(1), balance)
				// Check if the NFT has been transferred correctly
				exists, _, _, _, _, owner, _ := storage.GetAssetNFT(ctx, store, nftID)
				require.True(t, exists)
				require.Equal(t, addr.String(), owner.String())
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "InvalidOwnerForNFTTransfer",
			Actor: addr, // Someone other than the actual owner
			Action: &Transfer{
				To:      codec.EmptyAddress,
				Value:   1,
				AssetID: nftID, // Assume this is an NFT asset
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set initial ownership of the NFT to another address
				collectionID, uniqueID, uri, metadata, owner := assetID, uint64(0), "uri", "metadata", codec.EmptyAddress
				require.NoError(t, storage.SetAssetNFT(context.Background(), store, collectionID, uniqueID, nftID, []byte(uri), []byte(metadata), owner))
				return store
			}(),
			ExpectedErr: ErrOutputWrongOwner,
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkSimpleTransfer(b *testing.B) {
	require := require.New(b)
	to := codec.CreateAddress(0, ids.GenerateTestID())
	from := codec.CreateAddress(0, ids.GenerateTestID())

	transferActionTest := &chaintest.ActionBenchmark{
		Name:  "SimpleTransferBenchmark",
		Actor: from,
		Action: &Transfer{
			To:    to,
			Value: 1,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			err := storage.SetBalance(context.Background(), store, from, ids.Empty, 1)
			require.NoError(err)
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			toBalance, err := storage.GetBalance(ctx, store, to, ids.Empty)
			require.NoError(err)
			require.Equal(uint64(1), toBalance)

			fromBalance, err := storage.GetBalance(ctx, store, from, ids.Empty)
			require.NoError(err)
			require.Equal(uint64(0), fromBalance)
		},
	}

	ctx := context.Background()
	transferActionTest.Run(ctx, b)
}
