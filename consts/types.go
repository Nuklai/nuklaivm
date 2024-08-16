// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

// Note: Registry will error during initialization if a duplicate ID is assigned. We explicitly assign IDs to avoid accidental remapping.
const (
	// Action TypeIDs
	TransferID uint8 = 0

	CreateAssetID  uint8 = 1
	MintAssetFTID  uint8 = 2
	MintAssetNFTID uint8 = 3
	BurnAssetFTID  uint8 = 4
	BurnAssetNFTID uint8 = 5
	ExportAssetID  uint8 = 6
	ImportAssetID  uint8 = 7

	RegisterValidatorStakeID     uint8 = 8
	ClaimValidatorStakeRewardsID uint8 = 9
	WithdrawValidatorStakeID     uint8 = 10
	DelegateUserStakeID          uint8 = 11
	ClaimDelegationStakeRewards  uint8 = 12
	UndelegateUserStakeID        uint8 = 13

	// Auth TypeIDs
	ED25519ID   uint8 = 0
	SECP256R1ID uint8 = 1
	BLSID       uint8 = 2
)
