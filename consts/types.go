// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

const (
	// Action TypeIDs
	TransferID                    uint8 = iota // 0
	ContractCallID                             // 1
	ContractDeployID                           // 2
	ContractPublishID                          // 3
	CreateAssetID                              // 4
	UpdateAssetID                              // 5
	MintAssetFTID                              // 6
	MintAssetNFTID                             // 7
	BurnAssetFTID                              // 8
	BurnAssetNFTID                             // 9
	RegisterValidatorStakeID                   // 10
	WithdrawValidatorStakeID                   // 11
	ClaimValidatorStakeRewardsID               // 12
	DelegateUserStakeID                        // 13
	UndelegateUserStakeID                      // 14
	ClaimDelegationStakeRewardsID              // 15
	CreateDatasetID                            // 16
	UpdateDatasetID                            // 17
	InitiateContributeDatasetID                // 18
	CompleteContributeDatasetID                // 19
	PublishDatasetMarketplaceID                // 20
	SubscribeDatasetMarketplaceID              // 21
	ClaimMarketplacePaymentID                  // 22
)

const (
	// Asset TypeIDs
	AssetFungibleTokenID    uint8 = iota // 0
	AssetNonFungibleTokenID              // 1
	AssetFractionalTokenID               // 2
	AssetMarketplaceTokenID              // 3

	// Asset Names for TypeIDs
	AssetFungibleTokenDesc    = "Fungible Token"     // #nosec
	AssetNonFungibleTokenDesc = "Non-Fungible Token" // #nosec
	AssetFractionalTokenDesc  = "Fractional Token"   // #nosec
	AssetMarketplaceTokenDesc = "Marketplace Token"  // #nose
)
