// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/codec/codectest"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/state/tstate"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

func TestTransferAction(t *testing.T) {
	req := require.New(t)
	ts := tstate.New(1)

	actor1 := codectest.NewRandomAddress()
	actor2 := codectest.NewRandomAddress()

	assetAddress := codectest.NewRandomAddress()
	nftAddress := codectest.NewRandomAddress()

	parentState := ts.NewView(
		state.Keys{
			string(storage.AssetInfoKey(storage.NAIAddress)):                   state.All,
			string(storage.AssetInfoKey(assetAddress)):                         state.All,
			string(storage.AssetInfoKey(nftAddress)):                           state.All,
			string(storage.AssetAccountBalanceKey(storage.NAIAddress, actor1)): state.All,
			string(storage.AssetAccountBalanceKey(assetAddress, actor1)):       state.All,
			string(storage.AssetAccountBalanceKey(nftAddress, actor1)):         state.All,
			string(storage.AssetAccountBalanceKey(storage.NAIAddress, actor2)): state.All,
			string(storage.AssetAccountBalanceKey(assetAddress, actor2)):       state.All,
			string(storage.AssetAccountBalanceKey(nftAddress, actor2)):         state.All,
		},
		chaintest.NewInMemoryStore().Storage,
	)
	req.NoError(storage.SetAssetInfo(context.Background(), parentState, storage.NAIAddress, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 1, 0, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))

	tests := []chaintest.ActionTest{
		{
			Name:  "Can only transfer existing tokens",
			Actor: actor1,
			Action: &Transfer{
				To:           actor2,
				AssetAddress: codectest.NewRandomAddress(),
				Value:        1,
			},
			ExpectedOutputs: nil,
			ExpectedErr:     ErrAssetDoesNotExist,
			State:           parentState,
		},
		{
			Name:  "Transfer value must be greater than 0",
			Actor: actor1,
			Action: &Transfer{
				To:           actor2,
				AssetAddress: storage.NAIAddress,
				Value:        0,
			},
			ExpectedOutputs: nil,
			ExpectedErr:     ErrValueZero,
			State:           parentState,
		},
		{
			Name:  "NotEnoughBalance",
			Actor: actor1,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        1,
			},
			State:       parentState,
			ExpectedErr: storage.ErrInsufficientAssetBalance,
		},
		{
			Name:  "OverflowBalance",
			Actor: actor1,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        math.MaxUint64,
			},
			State:       parentState,
			ExpectedErr: storage.ErrInsufficientAssetBalance,
		},
		{
			Name:  "MemoSizeExceeded",
			Actor: actor1,
			Action: &Transfer{
				To:           codec.EmptyAddress,
				AssetAddress: storage.NAIAddress,
				Value:        1,
				Memo:         strings.Repeat("a", storage.MaxAssetMetadataSize+1),
			},
			State:       parentState,
			ExpectedErr: ErrMemoTooLarge,
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}

	req.NoError(storage.SetAssetAccountBalance(context.Background(), parentState, storage.NAIAddress, actor1, 1))
	req.NoError(storage.SetAssetInfo(context.Background(), parentState, assetAddress, nconsts.AssetFungibleTokenID, []byte("My Token"), []byte("MYT"), 9, []byte("Metadata"), []byte("uri"), 1, 0, actor1, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
	req.NoError(storage.SetAssetAccountBalance(context.Background(), parentState, assetAddress, actor1, 1))
	req.NoError(storage.SetAssetInfo(context.Background(), parentState, nftAddress, nconsts.AssetNonFungibleTokenID, []byte("My Token"), []byte("MYT"), 0, []byte("Metadata"), assetAddress[:], 1, 0, actor1, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress))
	req.NoError(storage.SetAssetAccountBalance(context.Background(), parentState, nftAddress, actor1, 1))

	tests = []chaintest.ActionTest{
		{
			Name:  "InvalidInsufficientAssetBalance",
			Actor: actor1,
			Action: &Transfer{
				To:           actor2,
				AssetAddress: storage.NAIAddress,
				Value:        5,
			},
			State:       parentState,
			ExpectedErr: storage.ErrInsufficientAssetBalance,
		},
		{
			Name:  "SelfTransferShouldNotBePossible",
			Actor: actor1,
			Action: &Transfer{
				To:           actor1,
				AssetAddress: storage.NAIAddress,
				Value:        1,
			},
			State:       parentState,
			ExpectedErr: ErrTransferToSelf,
		},
		{
			Name:  "SimpleTransfer",
			Actor: actor1,
			Action: &Transfer{
				To:           actor2,
				AssetAddress: storage.NAIAddress,
				Value:        1,
			},
			State: parentState,
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor2)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, storage.NAIAddress, actor1)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				Actor:           actor1.String(),
				Receiver:        actor2.String(),
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "ValidFTTransfer",
			Actor: actor1,
			Action: &Transfer{
				To:           actor2,
				AssetAddress: assetAddress,
				Value:        1,
			},
			State: parentState,
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor2)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor1)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				Actor:           actor1.String(),
				Receiver:        actor2.String(),
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
		{
			Name:  "ValidNFTTransfer",
			Actor: actor1,
			Action: &Transfer{
				To:           actor2,
				AssetAddress: nftAddress,
				Value:        1,
			},
			State: parentState,
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check NFT balances
				receiverBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, actor2)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err := storage.GetAssetAccountBalanceNoController(ctx, store, nftAddress, actor1)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
				// Check collectionAsset balances
				receiverBalance, err = storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor2)
				require.NoError(t, err)
				require.Equal(t, receiverBalance, uint64(1))
				senderBalance, err = storage.GetAssetAccountBalanceNoController(ctx, store, assetAddress, actor1)
				require.NoError(t, err)
				require.Equal(t, senderBalance, uint64(0))
			},
			ExpectedOutputs: &TransferResult{
				Actor:           actor1.String(),
				Receiver:        actor2.String(),
				SenderBalance:   0,
				ReceiverBalance: 1,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}
