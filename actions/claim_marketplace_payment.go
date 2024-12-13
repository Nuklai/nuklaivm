// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	ClaimMarketplacePaymentComputeUnits = 5
)

var (
	ErrNoPaymentRemaining              = errors.New("no payment remaining")
	_                     chain.Action = (*ClaimMarketplacePayment)(nil)
)

type ClaimMarketplacePayment struct {
	// Marketplace asset address that represents the dataset subscription in the
	// marketplace
	MarketplaceAssetAddress codec.Address `serialize:"true" json:"marketplace_asset_address"`

	// Asset to use for the payment
	PaymentAssetAddress codec.Address `serialize:"true" json:"payment_asset_address"`
}

func (*ClaimMarketplacePayment) GetTypeID() uint8 {
	return nconsts.ClaimMarketplacePaymentID
}

func (c *ClaimMarketplacePayment) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.AssetInfoKey(c.MarketplaceAssetAddress)):              state.Read | state.Write,
		string(storage.AssetInfoKey(c.PaymentAssetAddress)):                  state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(c.PaymentAssetAddress, actor)): state.All,
	}
}

func (c *ClaimMarketplacePayment) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check for the asset
	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAssetInfoNoController(ctx, mu, c.MarketplaceAssetAddress)
	if err != nil {
		return nil, err
	}
	// Ensure the asset is a marketplace token
	if assetType != nconsts.AssetMarketplaceTokenID {
		return nil, ErrAssetTypeInvalid
	}
	// Check if the user is the owner of the asset
	// Only the owner can claim the payment
	if owner != actor {
		return nil, ErrWrongOwner
	}

	// Convert the metdata to a map
	metadataMap, err := utils.BytesToMap(metadata)
	if err != nil {
		return nil, err
	}
	// Ensure paymentAssetAddress is supported
	if metadataMap["paymentAssetAddress"] != c.PaymentAssetAddress.String() {
		return nil, ErrPaymentAssetNotSupported
	}

	// Parse existing values from the metadata
	paymentRemaining, err := strconv.ParseUint(metadataMap["paymentRemaining"], 10, 64)
	if err != nil {
		return nil, err
	}
	if paymentRemaining == 0 {
		return nil, ErrNoPaymentRemaining
	}
	paymentClaimed, err := strconv.ParseUint(metadataMap["paymentClaimed"], 10, 64)
	if err != nil {
		return nil, err
	}
	lastClaimedBlock, err := strconv.ParseUint(metadataMap["lastClaimedBlock"], 10, 64)
	if err != nil {
		return nil, err
	}

	// Store the initial total before updating
	initialTotal := paymentRemaining + paymentClaimed
	_, _, _, paymentAssetDecimals, _, _, _, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, c.PaymentAssetAddress)
	if err != nil {
		return nil, err
	}
	baseValueOfOneUnit, _ := utils.ParseBalance("1", paymentAssetDecimals)
	// Get the current block height
	currentBlockHeight := emission.GetEmission().GetLastAcceptedBlockHeight()
	// Calculate the number of blocks the subscription has been active
	numBlocksSubscribed := currentBlockHeight - lastClaimedBlock
	// Calculate the total accumulated reward since the subscription started
	totalAccumulatedReward := numBlocksSubscribed * baseValueOfOneUnit
	// Cap the reward at the remaining payment if necessary
	if totalAccumulatedReward > paymentRemaining {
		totalAccumulatedReward = paymentRemaining
	}
	// Update paymentRemaining to reflect the updated balance
	paymentRemaining -= totalAccumulatedReward
	// Add the new accumulated reward to paymentClaimed
	paymentClaimed += totalAccumulatedReward
	// Update the lastClaimedBlock to the current block height so that next time we only accumulate from here
	lastClaimedBlock = currentBlockHeight
	finalTotal := paymentRemaining + paymentClaimed

	// Ensure the final total is consistent with the initial total
	if initialTotal != finalTotal {
		return nil, fmt.Errorf("inconsistent state: initial total (%d) does not match final total (%d)", initialTotal, finalTotal)
	}

	// Now, paymentRemaining, paymentClaimed, and lastClaimedBlock are updated based on the reward accumulated per block
	// You can now proceed to update these values in the relevant data structure or write them back to metadataMap
	metadataMap["paymentRemaining"] = strconv.FormatUint(paymentRemaining, 10)
	metadataMap["paymentClaimed"] = strconv.FormatUint(paymentClaimed, 10)
	metadataMap["lastClaimedBlock"] = strconv.FormatUint(lastClaimedBlock, 10)

	// Marshal the map back to a JSON byte slice
	updatedMetadata, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}

	// Update the asset with the updated metadata
	if err := storage.SetAssetInfo(ctx, mu, c.MarketplaceAssetAddress, assetType, name, symbol, decimals, updatedMetadata, uri, totalSupply, maxSupply, owner, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// TODO: Distribute the rewards to all the users who contributed to the dataset
	// This only distributes the rewards to the owner of the dataset
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, c.PaymentAssetAddress, actor)
	if err != nil {
		return nil, err
	}
	newBalance, err := smath.Add(balance, totalAccumulatedReward)
	if err != nil {
		return nil, err
	}
	if err = storage.SetAssetAccountBalance(ctx, mu, c.PaymentAssetAddress, actor, newBalance); err != nil {
		return nil, err
	}

	return &ClaimMarketplacePaymentResult{
		Actor:             actor.String(),
		Receiver:          actor.String(),
		LastClaimedBlock:  lastClaimedBlock,
		PaymentClaimed:    paymentClaimed,
		PaymentRemaining:  paymentRemaining,
		DistributedReward: totalAccumulatedReward,
		DistributedTo:     actor.String(),
	}, nil
}

func (*ClaimMarketplacePayment) ComputeUnits(chain.Rules) uint64 {
	return ClaimMarketplacePaymentComputeUnits
}

func (*ClaimMarketplacePayment) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalClaimMarketplacePayment(p *codec.Packer) (chain.Action, error) {
	var claimPaymentResult ClaimMarketplacePayment
	p.UnpackAddress(&claimPaymentResult.MarketplaceAssetAddress)
	p.UnpackAddress(&claimPaymentResult.PaymentAssetAddress)
	return &claimPaymentResult, p.Err()
}

var (
	_ codec.Typed = (*ClaimMarketplacePaymentResult)(nil)
)

type ClaimMarketplacePaymentResult struct {
	Actor             string `serialize:"true" json:"actor"`
	Receiver          string `serialize:"true" json:"receiver"`
	LastClaimedBlock  uint64 `serialize:"true" json:"last_claimed_block"`
	PaymentClaimed    uint64 `serialize:"true" json:"payment_claimed"`
	PaymentRemaining  uint64 `serialize:"true" json:"payment_remaining"`
	DistributedReward uint64 `serialize:"true" json:"distributed_reward"`
	DistributedTo     string `serialize:"true" json:"distributed_to"`
}

func (*ClaimMarketplacePaymentResult) GetTypeID() uint8 {
	return nconsts.ClaimMarketplacePaymentID
}

func UnmarshalClaimMarketplacePaymentResult(p *codec.Packer) (codec.Typed, error) {
	var result ClaimMarketplacePaymentResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	result.LastClaimedBlock = p.UnpackUint64(false)
	result.PaymentClaimed = p.UnpackUint64(false)
	result.PaymentRemaining = p.UnpackUint64(false)
	result.DistributedReward = p.UnpackUint64(false)
	result.DistributedTo = p.UnpackString(true)
	return &result, p.Err()
}
