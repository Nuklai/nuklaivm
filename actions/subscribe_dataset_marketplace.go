// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nchain "github.com/nuklai/nuklaivm/chain"
	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/marketplace"
	"github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*SubscribeDatasetMarketplace)(nil)

type SubscribeDatasetMarketplace struct {
	// Dataset ID
	Dataset ids.ID `json:"dataset"`

	// Marketplace ID(This is also the asset ID in the marketplace that represents the dataset)
	MarketplaceID ids.ID `json:"marketplaceID"`

	// Asset to use for the subscription
	AssetForPayment ids.ID `json:"assetForPayment"`

	// Total amount of blocks to subscribe to
	NumBlocksToSubscribe uint64 `json:"numBlocksToSubscribe"`
}

func (*SubscribeDatasetMarketplace) GetTypeID() uint8 {
	return nconsts.SubscribeDatasetMarketplaceID
}

func (d *SubscribeDatasetMarketplace) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	nftID := nchain.GenerateIDWithAddress(d.MarketplaceID, actor)
	return state.Keys{
		string(storage.DatasetKey(d.Dataset)):                state.Read,
		string(storage.AssetKey(d.MarketplaceID)):            state.Read | state.Write,
		string(storage.AssetNFTKey(nftID)):                   state.Allocate | state.Write,
		string(storage.BalanceKey(actor, d.AssetForPayment)): state.Read | state.Write,
		string(storage.BalanceKey(actor, d.MarketplaceID)):   state.Allocate | state.Write,
		string(storage.BalanceKey(actor, nftID)):             state.Allocate | state.Write,
	}
}

func (*SubscribeDatasetMarketplace) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.DatasetChunks, storage.AssetChunks, storage.AssetNFTChunks, storage.BalanceChunks, storage.BalanceChunks, storage.BalanceChunks}
}

func (d *SubscribeDatasetMarketplace) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	// Check if the nftID already exists(This means the user is already subscribed)
	nftID := nchain.GenerateIDWithAddress(d.MarketplaceID, actor)
	exists, _, _, _, _, _, _ := storage.GetAssetNFT(ctx, mu, nftID)
	if exists {
		return nil, ErrOutputNFTAlreadyExists
	}

	// Check if the dataset exists
	exists, _, _, _, _, _, _, _, _, saleID, baseAsset, basePrice, _, _, _, _, _, err := storage.GetDataset(ctx, mu, d.Dataset)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}

	// Check if the dataset is on sale
	if saleID == ids.Empty {
		return nil, ErrDatasetNotOnSale
	}
	// Check if the marketplace ID is correct
	if saleID != d.MarketplaceID {
		return nil, ErrMarketplaceIDInvalid
	}

	// Ensure assetForPayment is supported
	if d.AssetForPayment != baseAsset {
		return nil, ErrBaseAssetNotSupported
	}

	// Ensure numBlocksToSubscribe is valid
	dataConfig := marketplace.GetDatasetConfig()
	if d.NumBlocksToSubscribe < dataConfig.MinBlocksToSubscribe {
		return nil, ErrOutputNumBlocksToSubscribeInvalid
	}

	// Check for the asset
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAsset(ctx, mu, d.MarketplaceID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrOutputAssetMissing
	}
	if assetType != nconsts.AssetMarketplaceTokenID {
		return nil, ErrOutputWrongAssetType
	}

	// Mint the subscription non-fungible token to represent the user is subscribed
	// to the dataset
	amountOfToken := uint64(1)
	newSupply, err := smath.Add64(totalSupply, amountOfToken)
	if err != nil {
		return nil, err
	}
	if maxSupply != 0 && newSupply > maxSupply {
		return nil, ErrOutputMaxSupplyReached
	}
	totalSupply = newSupply

	// Calculate the total cost of the subscription
	totalCost := d.NumBlocksToSubscribe * basePrice

	// Get the emission instance
	emissionInstance := emission.GetEmission()
	currentBlock := emissionInstance.GetLastAcceptedBlockHeight()

	// Mint the NFT for the subscription
	metadataNFT := []byte("{\"dataset\":\"" + d.Dataset.String() + "\",\"marketplaceID\":\"" + d.MarketplaceID.String() + "\",\"datasetPricePerBlock\":\"" + fmt.Sprint(basePrice) + "\",\"totalCost\":\"" + fmt.Sprint(totalCost) + "\",\"assetForPayment\":\"" + d.AssetForPayment.String() + "\",\"issuanceBlock\":\"" + fmt.Sprint(currentBlock) + "\",\"expirationBlock\":\"" + fmt.Sprint(currentBlock+d.NumBlocksToSubscribe) + "\",\"numBlocksToSubscribe\":\"" + fmt.Sprint(d.NumBlocksToSubscribe) + "\"}")
	if err := storage.SetAssetNFT(ctx, mu, d.MarketplaceID, totalSupply, nftID, []byte(d.Dataset.String()), metadataNFT, actor); err != nil {
		return nil, err
	}

	// Unmarshal the metadata JSON into a map
	var metadataMap map[string]string
	if err := json.Unmarshal(metadata, &metadataMap); err != nil {
		return nil, err
	}

	// Update the paymentRemaining, subscriptions and lastClaimedBlock fields
	prevPaymentRemaining, err := strconv.ParseUint(metadataMap["paymentRemaining"], 10, 64)
	if err != nil {
		return nil, err
	}
	metadataMap["paymentRemaining"] = fmt.Sprint(prevPaymentRemaining + totalCost)

	prevSubscriptions, err := strconv.ParseUint(metadataMap["subscriptions"], 10, 64)
	if err != nil {
		return nil, err
	}
	metadataMap["subscriptions"] = fmt.Sprint(prevSubscriptions + 1)

	prevStartSubscriptionBlock, err := strconv.ParseUint(metadataMap["lastClaimedBlock"], 10, 64)
	if err != nil {
		return nil, err
	}
	if prevStartSubscriptionBlock == 0 {
		metadataMap["lastClaimedBlock"] = fmt.Sprint(currentBlock)
	}
	// Marshal the map back to a JSON byte slice
	updatedMetadata, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}

	// Update the asset with the new total supply and updated metadata
	if err := storage.SetAsset(ctx, mu, d.MarketplaceID, assetType, name, symbol, decimals, updatedMetadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor); err != nil {
		return nil, err
	}

	// Add the balance to NFT collection
	if err := storage.AddBalance(ctx, mu, actor, d.MarketplaceID, amountOfToken, true); err != nil {
		return nil, err
	}

	// Add the balance to individual NFT
	if err := storage.AddBalance(ctx, mu, actor, nftID, amountOfToken, true); err != nil {
		return nil, err
	}

	// Check if the actor has enough balance to subscribe
	if totalCost > 0 {
		if err := storage.SubBalance(ctx, mu, actor, d.AssetForPayment, totalCost); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (*SubscribeDatasetMarketplace) ComputeUnits(chain.Rules) uint64 {
	return SubscribeDatasetMarketplaceComputeUnits
}

func (*SubscribeDatasetMarketplace) Size() int {
	return ids.IDLen*3 + consts.Uint64Len
}

func (d *SubscribeDatasetMarketplace) Marshal(p *codec.Packer) {
	p.PackID(d.Dataset)
	p.PackID(d.MarketplaceID)
	p.PackID(d.AssetForPayment)
	p.PackUint64(d.NumBlocksToSubscribe)
}

func UnmarshalSubscribeDatasetMarketplace(p *codec.Packer) (chain.Action, error) {
	var subscribe SubscribeDatasetMarketplace
	p.UnpackID(true, &subscribe.Dataset)
	p.UnpackID(true, &subscribe.MarketplaceID)
	p.UnpackID(false, &subscribe.AssetForPayment)
	subscribe.NumBlocksToSubscribe = p.UnpackUint64(true)
	return &subscribe, p.Err()
}

func (*SubscribeDatasetMarketplace) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
