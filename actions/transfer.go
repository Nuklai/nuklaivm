// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	TransferComputeUnits = 1
)

var (
	ErrAssetDoesNotExist              = errors.New("asset does not exist")
	ErrValueZero                      = errors.New("value is zero")
	ErrNFTValueMustBeOne              = errors.New("NFT value must be one")
	ErrMemoTooLarge                   = errors.New("memo is too large")
	ErrTransferToSelf                 = errors.New("cannot transfer to self")
	_                    chain.Action = (*Transfer)(nil)
)

type Transfer struct {
	// To is the recipient of the [Value].
	To codec.Address `serialize:"true" json:"to"`

	// AssetAddress to transfer.
	AssetAddress codec.Address `serialize:"true" json:"asset_address"`

	// Amount are transferred to [To].
	Value uint64 `serialize:"true" json:"value"`

	// Optional message to accompany transaction.
	Memo string `serialize:"true" json:"memo"`
}

func (*Transfer) GetTypeID() uint8 {
	return nconsts.TransferID
}

func (t *Transfer) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(t.AssetAddress)):                  state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(t.AssetAddress, actor)): state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(t.AssetAddress, t.To)):  state.All,
	}
}

func (t *Transfer) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Ensure that the user is not transferring to self
	if actor == t.To {
		return nil, ErrTransferToSelf
	}
	// Check that asset exists
	assetType, _, _, _, _, _, _, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, t.AssetAddress)
	if err != nil {
		return nil, ErrAssetDoesNotExist
	}

	// Check the invariants
	if assetType == nconsts.AssetNonFungibleTokenID && t.Value != 1 {
		return nil, ErrNFTValueMustBeOne
	} else if t.Value == 0 {
		return nil, ErrValueZero
	}
	if len(t.Memo) > storage.MaxTextSize {
		return nil, ErrMemoTooLarge
	}

	// Check that balance is sufficient
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, t.AssetAddress, actor)
	if err != nil {
		return nil, err
	}
	if balance < t.Value {
		return nil, storage.ErrInsufficientAssetBalance
	}

	senderBalance, receiverBalance, err := storage.TransferAsset(ctx, mu, t.AssetAddress, actor, t.To, t.Value)
	if err != nil {
		return nil, err
	}

	return &TransferResult{
		SenderBalance:   senderBalance,
		ReceiverBalance: receiverBalance,
	}, nil
}

func (*Transfer) ComputeUnits(chain.Rules) uint64 {
	return TransferComputeUnits
}

func (*Transfer) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalTransfer(p *codec.Packer) (chain.Action, error) {
	var transfer Transfer
	p.UnpackAddress(&transfer.To)
	p.UnpackAddress(&transfer.AssetAddress)
	transfer.Value = p.UnpackUint64(true)
	transfer.Memo = p.UnpackString(false)
	return &transfer, p.Err()
}

var (
	_ codec.Typed     = (*TransferResult)(nil)
	_ chain.Marshaler = (*TransferResult)(nil)
)

type TransferResult struct {
	SenderBalance   uint64 `serialize:"true" json:"sender_balance"`
	ReceiverBalance uint64 `serialize:"true" json:"receiver_balance"`
}

func (*TransferResult) GetTypeID() uint8 {
	return nconsts.TransferID
}

func (*TransferResult) Size() int {
	return consts.Uint64Len * 2
}

func (r *TransferResult) Marshal(p *codec.Packer) {
	p.PackLong(r.SenderBalance)
	p.PackLong(r.ReceiverBalance)
}

func UnmarshalTransferResult(p *codec.Packer) (codec.Typed, error) {
	var result TransferResult
	result.SenderBalance = p.UnpackUint64(false)
	result.ReceiverBalance = p.UnpackUint64(false)
	return &result, p.Err()
}
