// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	hutils "github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*ClaimMarketplacePayment)(nil)

type ClaimMarketplacePayment struct {
	// Dataset ID
	Dataset ids.ID `json:"dataset"`

	// Marketplace ID(This is also the asset ID in the marketplace that represents the dataset)
	MarketplaceID ids.ID `json:"marketplaceID"`

	// Asset to use for the payment
	AssetForPayment ids.ID `json:"assetForPayment"`
}

func (*ClaimMarketplacePayment) GetTypeID() uint8 {
	return nconsts.ClaimMarketplacePaymentID
}

func (c *ClaimMarketplacePayment) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(c.Dataset)):                state.Read,
		string(storage.AssetKey(c.MarketplaceID)):            state.Read | state.Write,
		string(storage.AssetKey(c.AssetForPayment)):          state.Read,
		string(storage.BalanceKey(actor, c.AssetForPayment)): state.Allocate | state.Write,
	}
}

func (*ClaimMarketplacePayment) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks, storage.AssetChunks, storage.AssetChunks, storage.BalanceChunks}
}

func (c *ClaimMarketplacePayment) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if the dataset exists
	exists, _, _, _, _, _, _, _, _, _, baseAsset, _, _, _, _, _, owner, err := storage.GetDataset(ctx, mu, c.Dataset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}

	// Check if the user is the owner of the dataset
	// Only the onwer can claim the payment
	if owner != actor {
		return nil, ErrOutputWrongOwner
	}

	// Ensure assetForPayment is supported
	if c.AssetForPayment != baseAsset {
		return nil, ErrBaseAssetNotSupported
	}

	// Check if the marketplace asset exists
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, c.MarketplaceID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetMarketplaceTokenID {
		return nil, ErrOutputWrongAssetType
	}

	// Unmarshal the metadata JSON into a map
	var metadataMap map[string]string
	if err := json.Unmarshal(metadata, &metadataMap); err != nil {
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
		exists, _, _, _, decimals, _, _, _, _, _, _, _, _, _, err = storage.GetAsset(ctx, mu, c.AssetForPayment)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrAssetNotFound
		}
		decimalsToUse = decimals
	}
	baseValueOfOneUnit, _ := hutils.ParseBalance("1", decimalsToUse)
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
	if err := storage.SetAsset(ctx, mu, c.MarketplaceID, assetType, name, symbol, decimals, updatedMetadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// TODO: Distribute the rewards to all the users who contributed to the dataset
	if err := storage.AddBalance(ctx, mu, actor, c.AssetForPayment, totalAccumulatedReward, true); err != nil {
		return nil, err
	}

	sr := &ClaimPaymentResult{totalAccumulatedReward}
	output, err := sr.Marshal()
	if err != nil {
		return nil, err
	}
	return [][]byte{output}, nil
}

func (*ClaimMarketplacePayment) ComputeUnits(chain.Rules) uint64 {
	return ClaimMarketplacePaymentComputeUnits
}

func (*ClaimMarketplacePayment) Size() int {
	return ids.IDLen * 3
}

func (c *ClaimMarketplacePayment) Marshal(p *codec.Packer) {
	p.PackID(c.Dataset)
	p.PackID(c.MarketplaceID)
	p.PackID(c.AssetForPayment)
}

func UnmarshalClaimMarketplacePayment(p *codec.Packer) (chain.Action, error) {
	var claimPaymentResult ClaimMarketplacePayment
	p.UnpackID(true, &claimPaymentResult.Dataset)
	p.UnpackID(true, &claimPaymentResult.MarketplaceID)
	p.UnpackID(false, &claimPaymentResult.AssetForPayment)
	return &claimPaymentResult, p.Err()
}

func (*ClaimMarketplacePayment) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

type ClaimPaymentResult struct {
	RewardAmount uint64
}

func UnmarshalClaimPaymentResult(b []byte) (*ClaimPaymentResult, error) {
	p := codec.NewReader(b, consts.Uint64Len)
	var result ClaimPaymentResult
	result.RewardAmount = p.UnpackUint64(false)
	return &result, p.Err()
}

func (s *ClaimPaymentResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(consts.Uint64Len, consts.Uint64Len)
	p.PackUint64(s.RewardAmount)
	return p.Bytes(), p.Err()
}
