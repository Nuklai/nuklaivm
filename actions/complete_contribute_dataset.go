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
	DatasetContributionID string `serialize:"true" json:"dataset_contribution_id"`

	// DatasetAddress
	DatasetAddress codec.Address `serialize:"true" json:"dataset_address"`

	// DatasetContributor
	DatasetContributor codec.Address `serialize:"true" json:"dataset_contributor"`
}

func (*CompleteContributeDataset) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (d *CompleteContributeDataset) StateKeys(_ codec.Address) state.Keys {
	datasetContributionID, _ := ids.FromString(d.DatasetContributionID)
	nftAddress := codec.CreateAddress(nconsts.AssetFractionalTokenID, datasetContributionID)
	return state.Keys{
		string(storage.AssetInfoKey(d.DatasetAddress)): state.Read | state.Write,
		string(storage.AssetInfoKey(nftAddress)):       state.All,

		string(storage.DatasetInfoKey(d.DatasetAddress)):                                                                                   state.Read,
		string(storage.DatasetContributionInfoKey(datasetContributionID)):                                                                  state.Read | state.Write,
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
	datasetContributionID, err := ids.FromString(d.DatasetContributionID)
	if err != nil {
		return nil, err
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

	// Check if the dataset contribution exists
	datasetAddress, dataLocation, dataIdentifier, contributor, active, err := storage.GetDatasetContributionInfoNoController(ctx, mu, datasetContributionID)
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

	// Retrieve the asset info
	_, name, symbol, _, _, _, totalSupply, _, _, _, _, _, _, err := storage.GetAssetInfoNoController(ctx, mu, d.DatasetAddress)
	if err != nil {
		return nil, err
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
	nftAddress := codec.CreateAddress(nconsts.AssetFractionalTokenID, datasetContributionID)
	symbol = utils.CombineWithSuffix(symbol, totalSupply, storage.MaxSymbolSize)
	if err := storage.SetAssetInfo(ctx, mu, nftAddress, nconsts.AssetNonFungibleTokenID, name, symbol, 0, metadataNFT, []byte(d.DatasetAddress.String()), 0, 1, d.DatasetContributor, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress, codec.EmptyAddress); err != nil {
		return nil, err
	}
	if _, err := storage.MintAsset(ctx, mu, nftAddress, d.DatasetContributor, 1); err != nil {
		return nil, err
	}

	// Update the dataset contribution
	if err := storage.SetDatasetContributionInfo(ctx, mu, datasetContributionID, datasetAddress, dataLocation, dataIdentifier, contributor, true); err != nil {
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
		CollateralAssetAddress:   dataConfig.CollateralAssetAddressForDataContribution.String(),
		CollateralAmountRefunded: dataConfig.CollateralAmountForDataContribution,
		DatasetChildNftAddress:   nftAddress.String(),
		To:                       d.DatasetContributor.String(),
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
	complete.DatasetContributionID = p.UnpackString(true)
	p.UnpackAddress(&complete.DatasetAddress)
	p.UnpackAddress(&complete.DatasetContributor)
	return &complete, p.Err()
}

var (
	_ codec.Typed     = (*CompleteContributeDatasetResult)(nil)
	_ chain.Marshaler = (*CompleteContributeDatasetResult)(nil)
)

type CompleteContributeDatasetResult struct {
	CollateralAssetAddress   string `serialize:"true" json:"collateral_asset_address"`
	CollateralAmountRefunded uint64 `serialize:"true" json:"collateral_amount_refunded"`
	DatasetChildNftAddress   string `serialize:"true" json:"dataset_child_nft_address"`
	To                       string `serialize:"true" json:"to"`
	DataLocation             string `serialize:"true" json:"data_location"`
	DataIdentifier           string `serialize:"true" json:"data_identifier"`
}

func (*CompleteContributeDatasetResult) GetTypeID() uint8 {
	return nconsts.CompleteContributeDatasetID
}

func (r *CompleteContributeDatasetResult) Size() int {
	return codec.StringLen(r.CollateralAssetAddress) + consts.Uint64Len + codec.StringLen(r.DatasetChildNftAddress) + codec.StringLen(r.To) + codec.StringLen(r.DataLocation) + codec.StringLen(r.DataIdentifier)
}

func (r *CompleteContributeDatasetResult) Marshal(p *codec.Packer) {
	p.PackString(r.CollateralAssetAddress)
	p.PackUint64(r.CollateralAmountRefunded)
	p.PackString(r.DatasetChildNftAddress)
	p.PackString(r.To)
	p.PackString(r.DataLocation)
	p.PackString(r.DataIdentifier)
}

func UnmarshalCompleteContributeDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result CompleteContributeDatasetResult
	result.CollateralAssetAddress = p.UnpackString(true)
	result.CollateralAmountRefunded = p.UnpackUint64(false)
	result.DatasetChildNftAddress = p.UnpackString(true)
	result.To = p.UnpackString(true)
	result.DataLocation = p.UnpackString(true)
	result.DataIdentifier = p.UnpackString(true)
	return &result, p.Err()
}
