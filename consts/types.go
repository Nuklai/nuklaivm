// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

// Note: Registry will error during initialization if a duplicate ID is assigned. We explicitly assign IDs to avoid accidental remapping.
const (
	// Action TypeIDs
	TransferID uint8 = 0

	CreateAssetID  uint8 = 1
	UpdateAssetID  uint8 = 2
	MintAssetFTID  uint8 = 3
	MintAssetNFTID uint8 = 4
	BurnAssetFTID  uint8 = 5
	BurnAssetNFTID uint8 = 6
	ExportAssetID  uint8 = 7
	ImportAssetID  uint8 = 8

	RegisterValidatorStakeID     uint8 = 9
	ClaimValidatorStakeRewardsID uint8 = 10
	WithdrawValidatorStakeID     uint8 = 11
	DelegateUserStakeID          uint8 = 12
	ClaimDelegationStakeRewards  uint8 = 13
	UndelegateUserStakeID        uint8 = 14

	CreateDatasetID             uint8 = 15
	UpdateDatasetID             uint8 = 16
	InitiateContributeDatasetID uint8 = 17
	CompleteContributeDatasetID uint8 = 18

	// Asset TypeIDs
	AssetFungibleTokenID    uint8 = 0
	AssetNonFungibleTokenID uint8 = 1
	AssetDatasetTokenID     uint8 = 2
	// Asset Names for TypeIDs
	AssetFungibleTokenDesc    = "Fungible Token"     // #nosec
	AssetNonFungibleTokenDesc = "Non-Fungible Token" // #nosec
	AssetDatasetTokenDesc     = "Dataset Token"      // #nosec

	// Auth TypeIDs
	ED25519ID   uint8 = 0
	SECP256R1ID uint8 = 1
	BLSID       uint8 = 2
)
