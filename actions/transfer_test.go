// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"
)

func TestTransferAction(t *testing.T) {
	actor := codectest.NewRandomAddress()
	assetAddress := codectest.NewRandomAddress()
	nftAddress := codectest.NewRandomAddress()

	tests := []chaintest.ActionTest{
		{
			Name:  "ZeroTransfer",
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        0,
			},
			ExpectedErr: ErrValueZero,
		},
		{
			Name:  "NotEnoughBalance",
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        1,
			},
			State: func() state.Mutable {
				s := chaintest.NewInMemoryStore()
				_, err := storage.MintAsset(
					context.Background(),
					s,
					storage.NAIAddress,
					actor,
					0,
				)
				require.NoError(t, err)
				return s
			}(),
			ExpectedErr: storage.ErrInvalidBalance,
		},
		{
			Name:  "OverflowBalance",
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        math.MaxUint64,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, storage.NAIAddress, actor, 1))
				return store
			}(),
			ExpectedErr: storage.ErrInvalidBalance,
		},
		{
			Name:  "MemoSizeExceeded",
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        1,
				Memo:         strings.Repeat("a", storage.MaxAssetMetadataSize+1),
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrMemoTooLarge,
		},
		{
			Name:  "SelfTransfer",
			Actor: actor,
			Action: &Transfer{
				To:           actor,
				AssetAddress: storage.NAIAddress,
				Value:        1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, storage.NAIAddress, actor, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				balance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
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
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, storage.NAIAddress, actor, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, codec.EmptyAddress)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "ValidFTTransfer",
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: assetAddress,
				Value:        1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, assetAddress, consts.AssetFungibleTokenID, []byte("Asset1"), []byte("ASSET1"), 0, []byte("metadata"), []byte("uri"), 1, 0, actor, actor, actor, actor, actor))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, assetAddress, actor, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, codec.EmptyAddress)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "ValidNFTTransfer",
			Actor: actor,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: nftAddress,
				Value:        1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, nftAddress, consts.AssetNonFungibleTokenID, []byte("Asset1"), []byte("ASSET1"), 0, []byte("metadata"), []byte("uri"), 1, 0, actor, actor, actor, actor, actor))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, nftAddress, actor, 1))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, codec.EmptyAddress)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, actor)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "InvalidInsufficientTokenBalance",
			Actor: actor, // Someone other than the actual owner
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: nftAddress,
				Value:        1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				require.NoError(t, storage.SetAssetInfo(context.Background(), store, nftAddress, consts.AssetNonFungibleTokenID, []byte("Asset1"), []byte("ASSET1"), 0, []byte("metadata"), []byte("uri"), 1, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
				require.NoError(t, storage.SetAssetAccountBalance(context.Background(), store, nftAddress, codec.EmptyAddress, 1))
				return store
			}(),
			ExpectedErr: ErrInsufficientAssetBalance,
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkSimpleTransfer(b *testing.B) {
	require := require.New(b)
	to := codectest.NewRandomAddress()
	actor := codectest.NewRandomAddress()

	transferActionTest := &chaintest.ActionBenchmark{
		Name:  "SimpleTransferBenchmark",
		Actor: actor,
		Action: &Transfer{
			To:           to,
			AssetAddress: storage.NAIAddress,
			Value:        1,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			require.NoError(storage.SetAssetInfo(context.Background(), store, storage.NAIAddress, consts.AssetFungibleTokenID, []byte("Asset1"), []byte("ASSET1"), 0, []byte("metadata"), []byte("uri"), 1, 0, actor, actor, actor, actor, actor))
			require.NoError(storage.SetAssetAccountBalance(context.Background(), store, storage.NAIAddress, actor, 1))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, codec.EmptyAddress)
			require.NoError(err)
			require.Equal(receiverBalance, uint64(1))
			senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor)
			require.NoError(err)
			require.Equal(senderBalance, uint64(0))
		},
	}

	ctx := context.Background()
	transferActionTest.Run(ctx, b)
}
