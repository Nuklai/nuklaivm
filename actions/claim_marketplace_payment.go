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
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

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
	// DatasetID ID
	DatasetID ids.ID `json:"dataset_id"`

	// Marketplace ID(This is also the asset ID in the marketplace that represents the dataset)
	MarketplaceAssetID ids.ID `json:"marketplace_asset_id"`

	// Asset to use for the payment
	AssetForPayment ids.ID `json:"asset_for_payment"`
}

func (*ClaimMarketplacePayment) GetTypeID() uint8 {
	return nconsts.ClaimMarketplacePaymentID
}

func (c *ClaimMarketplacePayment) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.DatasetInfoKey(c.DatasetID)):          state.Read,
		string(storage.AssetInfoKey(c.MarketplaceAssetID)):   state.Read | state.Write,
		string(storage.AssetInfoKey(c.AssetForPayment)):      state.Read,
		string(storage.BalanceKey(actor, c.AssetForPayment)): state.Allocate | state.Write,
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
	// Check if the dataset exists
	exists, _, _, _, _, _, _, _, _, _, baseAsset, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, mu, c.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}

	// Check if the user is the owner of the dataset
	// Only the onwer can claim the payment
	if owner != actor {
		return nil, ErrWrongOwner
	}

	// Ensure assetForPayment is supported
	if c.AssetForPayment != baseAsset {
		return nil, ErrBaseAssetNotSupported
	}

	// Check if the marketplace asset exists
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAssetInfoNoController(ctx, mu, c.MarketplaceAssetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAssetMissing
	}
	if assetType != nconsts.AssetMarketplaceTokenID {
		return nil, ErrOutputWrongAssetType
	}

	// Unmarshal the metadata JSON into a map
	metadataMap, err := utils.BytesToMap(metadata)
	if err != nil {
		return nil, err
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
	// Parse the value of 1 in the base unit according to the number of decimals
	decimalsToUse := uint8(nconsts.Decimals)
	if c.AssetForPayment != ids.Empty {
		exists, _, _, _, decimals, _, _, _, _, _, _, _, _, _, err = storage.GetAssetInfoNoController(ctx, mu, c.AssetForPayment)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrAssetNotFound
		}
		decimalsToUse = decimals
	}
	baseValueOfOneUnit, _ := utils.ParseBalance("1", decimalsToUse)
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
	if err := storage.SetAssetInfo(ctx, mu, c.MarketplaceAssetID, assetType, name, symbol, decimals, updatedMetadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// TODO: Distribute the rewards to all the users who contributed to the dataset
	// This only distributes the rewards to the owner of the dataset
	if _, err := storage.AddBalance(ctx, mu, actor, c.AssetForPayment, totalAccumulatedReward, true); err != nil {
		return nil, err
	}

	return &ClaimMarketplacePaymentResult{
		LastClaimedBlock:  lastClaimedBlock,
		PaymentClaimed:    paymentClaimed,
		PaymentRemaining:  paymentRemaining,
		DistributedReward: totalAccumulatedReward,
		DistributedTo:     actor,
	}, nil
}

func (*ClaimMarketplacePayment) ComputeUnits(chain.Rules) uint64 {
	return ClaimMarketplacePaymentComputeUnits
}

func (*ClaimMarketplacePayment) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*ClaimMarketplacePayment)(nil)

func (*ClaimMarketplacePayment) Size() int {
	return ids.IDLen * 3
}

func (c *ClaimMarketplacePayment) Marshal(p *codec.Packer) {
	p.PackID(c.DatasetID)
	p.PackID(c.MarketplaceAssetID)
	p.PackID(c.AssetForPayment)
}

func UnmarshalClaimMarketplacePayment(p *codec.Packer) (chain.Action, error) {
	var claimPaymentResult ClaimMarketplacePayment
	p.UnpackID(true, &claimPaymentResult.DatasetID)
	p.UnpackID(true, &claimPaymentResult.MarketplaceAssetID)
	p.UnpackID(false, &claimPaymentResult.AssetForPayment)
	return &claimPaymentResult, p.Err()
}

var (
	_ codec.Typed     = (*ClaimMarketplacePaymentResult)(nil)
	_ chain.Marshaler = (*ClaimMarketplacePaymentResult)(nil)
)

type ClaimMarketplacePaymentResult struct {
	LastClaimedBlock  uint64        `serialize:"true" json:"last_claimed_block"`
	PaymentClaimed    uint64        `serialize:"true" json:"payment_claimed"`
	PaymentRemaining  uint64        `serialize:"true" json:"payment_remaining"`
	DistributedReward uint64        `serialize:"true" json:"distributed_reward"`
	DistributedTo     codec.Address `serialize:"true" json:"distributed_to"`
}

func (*ClaimMarketplacePaymentResult) GetTypeID() uint8 {
	return nconsts.ClaimMarketplacePaymentID
}

func (*ClaimMarketplacePaymentResult) Size() int {
	return consts.Uint64Len*4 + codec.AddressLen
}

func (r *ClaimMarketplacePaymentResult) Marshal(p *codec.Packer) {
	p.PackUint64(r.LastClaimedBlock)
	p.PackUint64(r.PaymentClaimed)
	p.PackUint64(r.PaymentRemaining)
	p.PackUint64(r.DistributedReward)
	p.PackAddress(r.DistributedTo)
}

func UnmarshalClaimMarketplacePaymentResult(p *codec.Packer) (codec.Typed, error) {
	var result ClaimMarketplacePaymentResult
	result.LastClaimedBlock = p.UnpackUint64(false)
	result.PaymentClaimed = p.UnpackUint64(false)
	result.PaymentRemaining = p.UnpackUint64(false)
	result.DistributedReward = p.UnpackUint64(false)
	p.UnpackAddress(&result.DistributedTo)
	return &result, p.Err()
}
