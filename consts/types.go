// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

const (
	// Action TypeIDs
	TransferID                    uint8 = 0
	ContractCallID                uint8 = 1
	ContractDeployID              uint8 = 2
	ContractPublishID             uint8 = 3
	CreateAssetID                 uint8 = 4
	UpdateAssetID                 uint8 = 5
	MintAssetFTID                 uint8 = 6
	MintAssetNFTID                uint8 = 7
	BurnAssetFTID                 uint8 = 8
	BurnAssetNFTID                uint8 = 9
	CreateDatasetID               uint8 = 10
	UpdateDatasetID               uint8 = 11
	InitiateContributeDatasetID   uint8 = 12
	CompleteContributeDatasetID   uint8 = 13
	PublishDatasetMarketplaceID   uint8 = 14
	SubscribeDatasetMarketplaceID uint8 = 15
	ClaimMarketplacePaymentID     uint8 = 16

	// Asset TypeIDs
	AssetFungibleTokenID    uint8 = 0
	AssetNonFungibleTokenID uint8 = 1
	AssetDatasetTokenID     uint8 = 2
	AssetMarketplaceTokenID uint8 = 3
	// Asset Names for TypeIDs
	AssetFungibleTokenDesc    = "Fungible Token"     // #nosec
	AssetNonFungibleTokenDesc = "Non-Fungible Token" // #nosec
	AssetDatasetTokenDesc     = "Dataset Token"      // #nosec
	AssetMarketplaceTokenDesc = "Marketplace Token"  // #nosec
)
