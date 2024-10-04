// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

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
	ErrOutputValueZero                      = errors.New("value is zero")
	ErrOutputMemoTooLarge                   = errors.New("memo is too large")
	ErrOutputWrongOwner                     = errors.New("wrong owner")
	ErrOutputNFTValueMustBeOne              = errors.New("NFT value must be one")
	_                          chain.Action = (*Transfer)(nil)
)

type Transfer struct {
	// To is the recipient of the [Value].
	To codec.Address `serialize:"true" json:"to"`

	// AssetID to transfer.
	AssetID []byte `serialize:"true" json:"asset_id"`

	// Amount are transferred to [To].
	Value uint64 `serialize:"true" json:"value"`

	// Optional message to accompany transaction.
	Memo []byte `serialize:"true" json:"memo"`
}

func (*Transfer) GetTypeID() uint8 {
	return nconsts.TransferID
}

func (t *Transfer) StateKeys(actor codec.Address) state.Keys {
	assetID, _ := utils.GetAssetIDBySymbol(string(t.AssetID))
	// Initialize the base stateKeys map
	stateKeys := state.Keys{
		string(storage.BalanceKey(actor, assetID)): state.Read | state.Write,
		string(storage.BalanceKey(t.To, assetID)):  state.All,
	}

	// Check if t.Asset is not empty, then add to stateKeys
	if assetID != ids.Empty {
		stateKeys[string(storage.AssetKey(assetID))] = state.Read | state.Write
		stateKeys[string(storage.AssetNFTKey(assetID))] = state.Read | state.Write
	}

	// Return the modified stateKeys
	return stateKeys
}

func (t *Transfer) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	assetID, err := utils.GetAssetIDBySymbol(string(t.AssetID))
	if err != nil {
		return nil, err
	}

	// Handle NFT transfers
	if assetID != ids.Empty {
		exists, collectionID, uniqueID, uri, metadata, owner, _ := storage.GetAssetNFT(ctx, mu, assetID)
		if exists {
			// Check if the sender is the owner of the NFT
			if owner != actor {
				return nil, ErrOutputWrongOwner
			}
			if t.Value != 1 {
				return nil, ErrOutputNFTValueMustBeOne
			}
			// Subtract the balance from NFT collection for the original NFT owner
			if _, err := storage.SubBalance(ctx, mu, actor, collectionID, t.Value); err != nil {
				return nil, err
			}
			// Add the balance to NFT collection for the new NFT owner
			if _, err := storage.AddBalance(ctx, mu, t.To, collectionID, t.Value, true); err != nil {
				return nil, err
			}
			// Update the NFT Info
			nftID := utils.GenerateIDWithIndex(collectionID, uniqueID)
			if err := storage.SetAssetNFT(ctx, mu, collectionID, uniqueID, nftID, uri, metadata, t.To); err != nil {
				return nil, err
			}
		}
	}

	if t.Value == 0 {
		return nil, ErrOutputValueZero
	}
	if len(t.Memo) > MaxMemoSize {
		return nil, ErrOutputMemoTooLarge
	}
	senderBalance, err := storage.SubBalance(ctx, mu, actor, assetID, t.Value)
	if err != nil {
		return nil, err
	}
	receiverBalance, err := storage.AddBalance(ctx, mu, t.To, assetID, t.Value, true)
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

// Implementing chain.Marshaler is optional but can be used to optimize performance when hitting TPS limits
var _ chain.Marshaler = (*Transfer)(nil)

func (t *Transfer) Size() int {
	return codec.AddressLen + codec.BytesLen(t.AssetID) + consts.Uint64Len + codec.BytesLen(t.Memo)
}

func (t *Transfer) Marshal(p *codec.Packer) {
	p.PackAddress(t.To)
	p.PackBytes(t.AssetID)
	p.PackLong(t.Value)
	p.PackBytes(t.Memo)
}

func UnmarshalTransfer(p *codec.Packer) (chain.Action, error) {
	var transfer Transfer
	p.UnpackAddress(&transfer.To)
	p.UnpackBytes(ids.IDLen, false, &transfer.AssetID)
	transfer.Value = p.UnpackUint64(true)
	p.UnpackBytes(MaxMemoSize, false, &transfer.Memo)
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
