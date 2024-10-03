// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/marketplace"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	InitiateContributeDatasetComputeUnits = 15
)

var (
	ErrDatasetNotOpenForContribution              = errors.New("dataset is not open for contribution")
	ErrDatasetAlreadyOnSale                       = errors.New("dataset is already on sale")
	ErrOutputDataLocationInvalid                  = errors.New("data location is invalid")
	_                                chain.Action = (*InitiateContributeDataset)(nil)
)

type InitiateContributeDataset struct {
	// DatasetID ID
	DatasetID ids.ID `serialize:"true" json:"dataset_id"`

	// Data location(default, S3, Filecoin, etc.)
	DataLocation []byte `serialize:"true" json:"data_location"`

	// Data Identifier(id/hash/URL)
	DataIdentifier []byte `serialize:"true" json:"data_identifier"`
}

func (*InitiateContributeDataset) GetTypeID() uint8 {
	return nconsts.InitiateContributeDatasetID
}

func (d *InitiateContributeDataset) StateKeys(actor codec.Address) state.Keys {
	return state.Keys{
		string(storage.DatasetKey(d.DatasetID)):      state.Read,
		string(storage.BalanceKey(actor, ids.Empty)): state.Read | state.Write,
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
	// Check if the dataset exists
	exists, _, _, _, _, _, _, _, isCommunityDataset, saleID, _, _, _, _, _, _, _, err := storage.GetDataset(ctx, mu, d.DatasetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDatasetNotFound
	}
	if !isCommunityDataset {
		return nil, ErrDatasetNotOpenForContribution
	}

	// Check if the dataset is already on sale
	if saleID != ids.Empty {
		return nil, ErrDatasetAlreadyOnSale
	}

	// Check if the data location is valid
	dataLocation := []byte("default")
	if len(d.DataLocation) > 0 {
		dataLocation = d.DataLocation
	}
	if len(dataLocation) < 3 || len(dataLocation) > MaxTextSize {
		return nil, ErrOutputDataLocationInvalid
	}
	// Check if the data identifier is valid(MaxMetadataSize - MaxTextSize because the data location and data identifier are stored together as metadata in the NFT metadata)
	if len(d.DataIdentifier) == 0 || len(d.DataIdentifier) > (MaxMetadataSize-MaxTextSize) {
		return nil, ErrOutputURIInvalid
	}

	// Get the marketplace instance
	marketplaceInstance := marketplace.GetMarketplace()
	if err := marketplaceInstance.InitiateContributeDataset(d.DatasetID, dataLocation, d.DataIdentifier, actor); err != nil {
		return nil, err
	}

	// Reduce the balance of the contributor with the collateral needed to contribute to the dataset
	// This will be refunded if the contribution is successful
	// This is done to prevent spamming the network with fake contributions
	dataConfig := marketplace.GetDatasetConfig()
	if _, err := storage.SubBalance(ctx, mu, actor, dataConfig.CollateralAssetIDForDataContribution, dataConfig.CollateralAmountForDataContribution); err != nil {
		return nil, err
	}

	return &InitiateContributeDatasetResult{
		CollateralAssetID:     dataConfig.CollateralAssetIDForDataContribution,
		CollateralAmountTaken: dataConfig.CollateralAmountForDataContribution,
	}, nil
}

func (*InitiateContributeDataset) ComputeUnits(chain.Rules) uint64 {
	return InitiateContributeDatasetComputeUnits
}

func (*InitiateContributeDataset) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*InitiateContributeDataset)(nil)

func (d *InitiateContributeDataset) Size() int {
	return ids.IDLen + codec.BytesLen(d.DataLocation) + codec.BytesLen(d.DataIdentifier)
}

func (d *InitiateContributeDataset) Marshal(p *codec.Packer) {
	p.PackID(d.DatasetID)
	p.PackBytes(d.DataLocation)
	p.PackBytes(d.DataIdentifier)
}

func UnmarshalInitiateContributeDataset(p *codec.Packer) (chain.Action, error) {
	var initiate InitiateContributeDataset
	p.UnpackID(true, &initiate.DatasetID)
	p.UnpackBytes(MaxTextSize, false, &initiate.DataLocation)
	p.UnpackBytes(MaxMetadataSize-MaxTextSize, true, &initiate.DataIdentifier)
	return &initiate, p.Err()
}

var (
	_ codec.Typed     = (*InitiateContributeDatasetResult)(nil)
	_ chain.Marshaler = (*InitiateContributeDatasetResult)(nil)
)

type InitiateContributeDatasetResult struct {
	CollateralAssetID     ids.ID `serialize:"true" json:"collateral_asset_id"`
	CollateralAmountTaken uint64 `serialize:"true" json:"collateral_amount_taken"`
}

func (*InitiateContributeDatasetResult) GetTypeID() uint8 {
	return nconsts.InitiateContributeDatasetID
}

func (*InitiateContributeDatasetResult) Size() int {
	return ids.IDLen + consts.Uint64Len
}

func (r *InitiateContributeDatasetResult) Marshal(p *codec.Packer) {
	p.PackID(r.CollateralAssetID)
	p.PackUint64(r.CollateralAmountTaken)
}

func UnmarshalInitiateContributeDatasetResult(p *codec.Packer) (codec.Typed, error) {
	var result InitiateContributeDatasetResult
	p.UnpackID(false, &result.CollateralAssetID)
	result.CollateralAmountTaken = p.UnpackUint64(true)
	return &result, p.Err()
}
