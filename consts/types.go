// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

// Note: Registry will error during initialization if a duplicate ID is assigned. We explicitly assign IDs to avoid accidental remapping.
const (
	// Action TypeIDs
	TransferID           uint8 = 0
	StakeValidatorID     uint8 = 1
	UnstakeValidatorID   uint8 = 2
	ClaimStakingRewardID uint8 = 3
	CreateAssetID        uint8 = 4
	ExportAssetID        uint8 = 5
	ImportAssetID        uint8 = 6
	MintAssetID          uint8 = 7
	BurnAssetID          uint8 = 8

	// Auth TypeIDs
	ED25519ID   uint8 = 0
	SECP256R1ID uint8 = 1
	BLSID       uint8 = 2
)
