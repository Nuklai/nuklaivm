// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/dataset"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	smath "github.com/ava-labs/avalanchego/utils/math"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	InitiateContributeDatasetComputeUnits = 15
)

var (
	ErrDatasetNotOpenForContribution                 = errors.New("dataset is not open for contribution")
	ErrDatasetAlreadyOnSale                          = errors.New("dataset is already on sale")
	ErrDataLocationInvalid                           = errors.New("data location is invalid")
	ErrDataIdentifierInvalid                         = errors.New("data identifier is invalid")
	ErrDatasetContributionAlreadyExists              = errors.New("dataset contribution already exists")
	_                                   chain.Action = (*InitiateContributeDataset)(nil)
)

type InitiateContributeDataset struct {
	// DatasetAddress
	DatasetAddress codec.Address `serialize:"true" json:"dataset_address"`

	// Data location(default, S3, Filecoin, etc.)
	DataLocation string `serialize:"true" json:"data_location"`

	// Data Identifier(id/hash/URL)
	DataIdentifier string `serialize:"true" json:"data_identifier"`
}

func (*InitiateContributeDataset) GetTypeID() uint8 {
	return nconsts.InitiateContributeDatasetID
}

func (d *InitiateContributeDataset) StateKeys(actor codec.Address) state.Keys {
	datasetContributionID := storage.DatasetContributionID(d.DatasetAddress, []byte(d.DataLocation), []byte(d.DataIdentifier), actor)
	return state.Keys{
		string(storage.DatasetContributionInfoKey(datasetContributionID)):                                                   state.All,
		string(storage.DatasetInfoKey(d.DatasetAddress)):                                                                    state.Read,
		string(storage.AssetAccountBalanceKey(dataset.GetDatasetConfig().CollateralAssetAddressForDataContribution, actor)): state.Read | state.Write,
	}
}

func (d *InitiateContributeDataset) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	// Check if the dataset contribution exists
	datasetContributionID := storage.DatasetContributionID(d.DatasetAddress, []byte(d.DataLocation), []byte(d.DataIdentifier), actor)
	if storage.DatasetContributionExists(ctx, mu, datasetContributionID) {
		return nil, ErrDatasetContributionAlreadyExists
	}

	// Check if the dataset exists
	_, _, _, _, _, _, _, isCommunityDataset, marketplaceAssetAddress, _, _, _, _, _, _, _, err := storage.GetDatasetInfoNoController(ctx, mu, d.DatasetAddress)
	if err != nil {
		return nil, err
	}
	if !isCommunityDataset {
		return nil, ErrDatasetNotOpenForContribution
	}

	// Check if the dataset is already on sale
	if marketplaceAssetAddress != codec.EmptyAddress {
		return nil, ErrDatasetAlreadyOnSale
	}

	// Check if the data location is valid
	dataLocation := storage.DatasetDefaultLocation
	if len(d.DataLocation) > 0 {
		dataLocation = d.DataLocation
	}
	if len(dataLocation) > storage.MaxDatasetDataLocationSize {
		return nil, ErrDataLocationInvalid
	}
	// Check if the data identifier is valid(MaxAssetMetadataSize - MaxDatasetDataLocationSize because the data location and data identifier are stored together as metadata in the NFT metadata)
	if len(d.DataIdentifier) == 0 || len(d.DataIdentifier) > (storage.MaxAssetMetadataSize-storage.MaxDatasetDataLocationSize) {
		return nil, ErrDataIdentifierInvalid
	}

	// Set the dataset contribution info to storage
	if err := storage.SetDatasetContributionInfo(ctx, mu, datasetContributionID, d.DatasetAddress, []byte(d.DataLocation), []byte(d.DataIdentifier), actor, false); err != nil {
		return nil, err
	}

	// Reduce the balance of the contributor with the collateral needed to contribute to the dataset
	// This will be refunded if the contribution is successful
	// This is done to prevent spamming the network with fake contributions
	dataConfig := dataset.GetDatasetConfig()
	// Subtract the collateral amount from the balance
	// Ensure that the balance is sufficient
	balance, err := storage.GetAssetAccountBalanceNoController(ctx, mu, dataConfig.CollateralAssetAddressForDataContribution, actor)
	if err != nil {
		return nil, err
	}
	if balance < dataConfig.CollateralAmountForDataContribution {
		return nil, storage.ErrInsufficientAssetBalance
	}
	newBalance, err := smath.Sub(balance, dataConfig.CollateralAmountForDataContribution)
	if err != nil {
		return nil, err
	}
	if err = storage.SetAssetAccountBalance(ctx, mu, dataConfig.CollateralAssetAddressForDataContribution, actor, newBalance); err != nil {
		return nil, err
	}

	return &InitiateContributeDatasetResult{
		DatasetContributionID:  datasetContributionID.String(),
		CollateralAssetAddress: dataConfig.CollateralAssetAddressForDataContribution.String(),
		CollateralAmountTaken:  dataConfig.CollateralAmountForDataContribution,
	}, nil
}

func (*InitiateContributeDataset) ComputeUnits(chain.Rules) uint64 {
	return InitiateContributeDatasetComputeUnits
}

func (*InitiateContributeDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

func UnmarshalInitiateContributeDataset(p *codec.Packer) (chain.Action, error) {
	var initiate InitiateContributeDataset
	p.UnpackAddress(&initiate.DatasetAddress)
	p.UnpackString(false)
	p.UnpackString(true)
	return &initiate, p.Err()
}

var (
	_ codec.Typed     = (*InitiateContributeDatasetResult)(nil)
	_ chain.Marshaler = (*InitiateContributeDatasetResult)(nil)
)

type InitiateContributeDatasetResult struct {
	DatasetContributionID  string       `serialize:"true" json:"dataset_contribution_id"`
	CollateralAssetAddress string `serialize:"true" json:"collateral_asset_address"`
	CollateralAmountTaken  uint64        `serialize:"true" json:"collateral_amount_taken"`
}

func (*InitiateContributeDatasetResult) GetTypeID() uint8 {
	return nconsts.InitiateContributeDatasetID
}

func (i *InitiateContributeDatasetResult) Size() int {
	return codec.StringLen(i.DatasetContributionID) + codec.StringLen(i.CollateralAssetAddress) + consts.Uint64Len
}

func (r *InitiateContributeDatasetResult) Marshal(p *codec.Packer) {
	p.PackString(r.DatasetContributionID)
	p.PackString(r.CollateralAssetAddress)
	p.PackUint64(r.CollateralAmountTaken)
}

func UnmarshalInitiateContributeDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result InitiateContributeDatasetResult
	result.DatasetContributionID = p.UnpackString(true)
	result.CollateralAssetAddress = p.UnpackString(true)
	result.CollateralAmountTaken = p.UnpackUint64(true)
	return &result, p.Err()
}
