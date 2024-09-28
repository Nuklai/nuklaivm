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
	RegisterValidatorStakeID      uint8 = 10
	WithdrawValidatorStakeID      uint8 = 11
	ClaimValidatorStakeRewardsID  uint8 = 12
	DelegateUserStakeID           uint8 = 13
	UndelegateUserStakeID         uint8 = 14
	ClaimDelegationStakeRewards   uint8 = 15
	CreateDatasetID               uint8 = 16
	UpdateDatasetID               uint8 = 17
	InitiateContributeDatasetID   uint8 = 18
	CompleteContributeDatasetID   uint8 = 19
	PublishDatasetMarketplaceID   uint8 = 20
	SubscribeDatasetMarketplaceID uint8 = 21
	ClaimMarketplacePaymentID     uint8 = 22

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
