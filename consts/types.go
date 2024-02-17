// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

// Note: Registry will error during initialization if a duplicate ID is assigned. We explicitly assign IDs to avoid accidental remapping.
const (
	// Action TypeIDs
	TransferID uint8 = 0

	CreateAssetID uint8 = 1
	MintAssetID   uint8 = 2
	BurnAssetID   uint8 = 3
	ExportAssetID uint8 = 4
	ImportAssetID uint8 = 5

	RegisterValidatorStakeID uint8 = 6
	WithdrawValidatorStakeID uint8 = 7
	DelegateUserStakeID      uint8 = 8
	UndelegateUserStakeID    uint8 = 9
	ClaimStakingRewardsID    uint8 = 10

	// Auth TypeIDs
	ED25519ID   uint8 = 0
	SECP256R1ID uint8 = 1
	BLSID       uint8 = 2
)
