// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/dataset"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	CompleteContributeDatasetComputeUnits = 5
)

var (
	ErrDatasetContributionAlreadyComplete              = errors.New("dataset contribution already complete")
	ErrDatasetAddressMismatch                          = errors.New("dataset address mismatch")
	ErrDatasetContributorMismatch                      = errors.New("dataset contributor mismatch")
	_                                     chain.Action = (*CompleteContributeDataset)(nil)
)

type CompleteContributeDataset struct {
	// Contribution ID
	DatasetContributionID ids.ID `serialize:"true" json:"dataset_contribution_id"`

	// DatasetAddress
	DatasetAddress codec.Address `serialize:"true" json:"dataset_address"`

	// DatasetContributor
	DatasetContributor codec.Address `serialize:"true" json:"dataset_contributor"`
}

func (*CompleteContributeDataset) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (d *CompleteContributeDataset) StateKeys(_ codec.Address) state.Keys {
	nftAddress := codec.CreateAddress(nconsts.AssetFractionalTokenID, d.DatasetContributionID)
	return state.Keys{
		string(storage.AssetInfoKey(d.DatasetAddress)): state.Read | state.Write,
		string(storage.AssetInfoKey(nftAddress)):       state.All,

		string(storage.DatasetInfoKey(d.DatasetAddress)):                                                                                   state.Read,
		string(storage.DatasetContributionInfoKey(d.DatasetContributionID)):                                                                state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(dataset.GetDatasetConfig().CollateralAssetAddressForDataContribution, d.DatasetContributor)): state.Read | state.Write,
		string(storage.AssetAccountBalanceKey(d.DatasetAddress, d.DatasetContributor)):                                                     state.Allocate | state.Write,
		string(storage.AssetAccountBalanceKey(nftAddress, d.DatasetContributor)):                                                           state.All,
	}
}

func (d *CompleteContributeDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the dataset contribution exists
	datasetAddress, dataLocation, dataIdentifier, contributor, active, err := storage.GetDatasetContributionInfoNoController(ctx, mu, d.DatasetContributionID)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, ErrDatasetContributionAlreadyComplete
	}
	if datasetAddress != d.DatasetAddress {
		return nil, ErrDatasetAddressMismatch
	}
	if contributor != d.DatasetContributor {
		return nil, ErrDatasetContributorMismatch
	}

	// Check if the dataset exists
	_, _, _, _, _, _, _, _, marketplaceAssetAddress, _, _, _, _, _, _, owner, err := storage.GetDatasetInfoNoController(ctx, mu, d.DatasetAddress)
	if err != nil {
		return nil, err
	}
	if actor != owner {
		return nil, ErrWrongOwner
	}

	// Check if the dataset is already on sale
	if marketplaceAssetAddress != codec.EmptyAddress {
		return nil, ErrDatasetAlreadyOnSale
	}

	// Retrieve the asset info
	assetType, name, symbol, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, d.DatasetAddress)
	if err != nil {
		return nil, err
	}

	// Check if the nftAddress already exists
	nftAddress := codec.CreateAddress(nconsts.AssetFractionalTokenID, d.DatasetContributionID)
	if storage.AssetExists(ctx, mu, nftAddress) {
		return nil, ErrNFTAlreadyExists
	}

	// Minting logic for non-fungible tokens
	if _, err := storage.MintAsset(ctx, mu, d.DatasetAddress, d.DatasetContributor, 1); err != nil {
		return nil, err
	}
	// Set the metadata for the NFT
	metadataNFTMap := make(map[string]string, 0)
	metadataNFTMap["dataLocation"] = string(dataLocation)
	metadataNFTMap["dataIdentifier"] = string(dataIdentifier)
	metadataNFT, err := utils.MapToBytes(metadataNFTMap)
	if err != nil {
		return nil, err
	}
	symbol = utils.CombineWithSuffix(symbol, totalSupply+1, storage.MaxSymbolSize)
	if err := storage.SetAssetInfo(ctx, mu, nftAddress, assetType, name, symbol, 0, metadataNFT, []byte(d.DatasetAddress.String()), 0, 1, d.DatasetContributor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}
	if _, err := storage.MintAsset(ctx, mu, nftAddress, d.DatasetContributor, 1); err != nil {
		return nil, err
	}

	// Update the dataset contribution
	if err := storage.SetDatasetContributionInfo(ctx, mu, d.DatasetContributionID, datasetAddress, dataLocation, dataIdentifier, contributor, true); err != nil {
		return nil, err
	}

	// Refund the collateral back to the contributor
	dataConfig := dataset.GetDatasetConfig()
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, dataConfig.CollateralAssetAddressForDataContribution, d.DatasetContributor)
	if err != nil {
		return nil, err
	}
	newBalance, err := smath.Add(balance, dataConfig.CollateralAmountForDataContribution)
	if err != nil {
		return nil, err
	}
	if err = storage.SetAssetAccountBalance(ctx, mu, dataConfig.CollateralAssetAddressForDataContribution, d.DatasetContributor, newBalance); err != nil {
		return nil, err
	}

	return &CompleteContributeDatasetResult{
		CollateralAssetAddress:   dataConfig.CollateralAssetAddressForDataContribution,
		CollateralAmountRefunded: dataConfig.CollateralAmountForDataContribution,
		DatasetChildNftAddress:   nftAddress,
		To:                       d.DatasetContributor,
		DataLocation:             string(dataLocation),
		DataIdentifier:           string(dataIdentifier),
	}, nil
}

func (*CompleteContributeDataset) ComputeUnits(chain.Rules) uint64 {
	return CompleteContributeDatasetComputeUnits
}

func (*CompleteContributeDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalCompleteContributeDataset(p *codec.Packer) (chain.Action, error) {
	var complete CompleteContributeDataset
	p.UnpackID(true, &complete.DatasetContributionID)
	p.UnpackAddress(&complete.DatasetAddress)
	p.UnpackAddress(&complete.DatasetContributor)
	return &complete, p.Err()
}

var (
	_ codec.Typed     = (*CompleteContributeDatasetResult)(nil)
	_ chain.Marshaler = (*CompleteContributeDatasetResult)(nil)
)

type CompleteContributeDatasetResult struct {
	CollateralAssetAddress   codec.Address `serialize:"true" json:"collateral_asset_address"`
	CollateralAmountRefunded uint64        `serialize:"true" json:"collateral_amount_refunded"`
	DatasetChildNftAddress   codec.Address `serialize:"true" json:"dataset_child_nft_address"`
	To                       codec.Address `serialize:"true" json:"to"`
	DataLocation             string        `serialize:"true" json:"data_location"`
	DataIdentifier           string        `serialize:"true" json:"data_identifier"`
}

func (*CompleteContributeDatasetResult) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (r *CompleteContributeDatasetResult) Size() int {
	return codec.AddressLen*3 + consts.Uint64Len + codec.StringLen(r.DataLocation) + codec.StringLen(r.DataIdentifier)
}

func (r *CompleteContributeDatasetResult) Marshal(p *codec.Packer) {
	p.PackAddress(r.CollateralAssetAddress)
	p.PackUint64(r.CollateralAmountRefunded)
	p.PackAddress(r.DatasetChildNftAddress)
	p.PackAddress(r.To)
	p.PackString(r.DataLocation)
	p.PackString(r.DataIdentifier)
}

func UnmarshalCompleteContributeDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result CompleteContributeDatasetResult
	p.UnpackAddress(&result.CollateralAssetAddress)
	result.CollateralAmountRefunded = p.UnpackUint64(false)
	p.UnpackAddress(&result.DatasetChildNftAddress)
	p.UnpackAddress(&result.To)
	result.DataLocation = p.UnpackString(true)
	result.DataIdentifier = p.UnpackString(true)
	return &result, p.Err()
}
