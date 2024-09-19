// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import "errors"

var (
	ErrOutputValueZero          = errors.New("value is zero")
	ErrOutputMemoTooLarge       = errors.New("memo is too large")
	ErrOutputAssetIsNative      = errors.New("cannot mint native asset")
	ErrOutputAssetAlreadyExists = errors.New("asset already exists")
	ErrOutputAssetMissing       = errors.New("asset missing")
	ErrOutputInTickZero         = errors.New("in rate is zero")
	ErrOutputOutTickZero        = errors.New("out rate is zero")
	ErrOutputSupplyZero         = errors.New("supply is zero")
	ErrOutputSupplyMisaligned   = errors.New("supply is misaligned")
	ErrOutputOrderMissing       = errors.New("order is missing")
	ErrOutputUnauthorized       = errors.New("unauthorized")
	ErrOutputWrongIn            = errors.New("wrong in asset")
	ErrOutputWrongOut           = errors.New("wrong out asset")
	ErrOutputWrongOwner         = errors.New("wrong owner")
	ErrOutputInsufficientInput  = errors.New("insufficient input")
	ErrOutputInsufficientOutput = errors.New("insufficient output")
	ErrOutputValueMisaligned    = errors.New("value is misaligned")

	ErrAssetNotFound                = errors.New("asset not found")
	ErrOutputAssetTypeInvalid       = errors.New("asset type is invalid. It must be 0(fungible), 1(non-fungible), and 2(dataset)")
	ErrOutputWrongAssetType         = errors.New("wrong asset type")
	ErrOutputNameInvalid            = errors.New("name is empty or too large")
	ErrOutputURIInvalid             = errors.New("uri is empty or too large")
	ErrOutputSymbolInvalid          = errors.New("symbol is empty or too large")
	ErrOutputDecimalsInvalid        = errors.New("decimal is invalid")
	ErrOutputMaxSupplyReached       = errors.New("max supply reached")
	ErrOutputMetadataInvalid        = errors.New("metadata is empty or too large")
	ErrOutputWrongMintActor         = errors.New("wrong mint actor")
	ErrOutputNFTAlreadyExists       = errors.New("NFT already exists")
	ErrOutputIDGreaterThanMaxSupply = errors.New("ID is greater than max supply")
	ErrOutputNFTValueGreaterThanOne = errors.New("NFT value must be 1")

	ErrOutputSameInOut          = errors.New("same asset used for in and out")
	ErrOutputWrongDestination   = errors.New("wrong destination")
	ErrOutputMustFill           = errors.New("must fill request")
	ErrOutputInvalidDestination = errors.New("invalid destination")

	ErrOutputStakeMissing  = errors.New("stake is missing")
	ErrOutputInvalidNodeID = errors.New("invalid node ID")
	ErrOutputStakeNotEnded = errors.New("stake not ended")

	// staking
	// register_validator_stake.go
	ErrOutputNotValidator                 = errors.New("not a validator")
	ErrOutputDifferentSignerThanActor     = errors.New("different signer than actor")
	ErrOutputValidatorStakedAmountInvalid = errors.New("invalid stake amount")
	ErrOutputInvalidStakeStartBlock       = errors.New("invalid stake start block")
	ErrOutputInvalidStakeEndBlock         = errors.New("invalid stake end block")
	ErrOutputInvalidStakeDuration         = errors.New("invalid stake duration")
	ErrOutputInvalidDelegationFeeRate     = errors.New("delegation fee rate must be over 2 and under 100")
	ErrOutputValidatorAlreadyRegistered   = errors.New("validator already registered for staking")
	ErrOutputStakeNotStarted              = errors.New("stake not started")

	// delegate_user_stake.go
	ErrOutputDelegateStakedAmountInvalid = errors.New("staked amount must be at least 25 NAI")
	ErrOutputUserAlreadyStaked           = errors.New("user already staked")
	ErrOutputValidatorNotYetRegistered   = errors.New("validator not yet registered for staking")

	// datasets
	ErrNotDatasetOwner                 = errors.New("not dataset owner")
	ErrOutputDescriptionInvalid        = errors.New("description is empty or too large")
	ErrOutputCategoriesInvalid         = errors.New("categories are empty or too large")
	ErrOutputLicenseNameInvalid        = errors.New("license name is empty or too large")
	ErrOutputLicenseSymbolInvalid      = errors.New("license symbol is empty or too large")
	ErrOutputLicenseURLInvalid         = errors.New("license URL is empty or too large")
	ErrDatasetNotFound                 = errors.New("dataset not found")
	ErrOutputMustUpdateAtLeastOneField = errors.New("must update at least one field")

	ErrOutputDataLocationInvalid     = errors.New("data location is invalid")
	ErrDatasetNotOpenForContribution = errors.New("dataset not open for contribution")

	ErrBaseAssetNotSupported             = errors.New("base asset not supported")
	ErrBasePriceInvalid                  = errors.New("base price is invalid")
	ErrDatasetAlreadyOnSale              = errors.New("dataset already on sale")
	ErrDatasetNotOnSale                  = errors.New("dataset not on sale")
	ErrOutputNumBlocksToSubscribeInvalid = errors.New("num blocks to subscribe is invalid")
	ErrMarketplaceIDInvalid              = errors.New("marketplace ID is invalid")
	ErrNoPaymentRemaining                = errors.New("no payment remaining")
)
