// Copyright (C) 2024, AllianceBlock. All rights reserved.
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
	ErrOutputSymbolEmpty        = errors.New("symbol is empty")
	ErrOutputSymbolIncorrect    = errors.New("symbol is incorrect")
	ErrOutputSymbolTooLarge     = errors.New("symbol is too large")
	ErrOutputDecimalsIncorrect  = errors.New("decimal is incorrect")
	ErrOutputDecimalsTooLarge   = errors.New("decimal is too large")
	ErrOutputMetadataEmpty      = errors.New("metadata is empty")
	ErrOutputMetadataTooLarge   = errors.New("metadata is too large")
	ErrOutputSameInOut          = errors.New("same asset used for in and out")
	ErrOutputWrongDestination   = errors.New("wrong destination")
	ErrOutputMustFill           = errors.New("must fill request")
	ErrOutputInvalidDestination = errors.New("invalid destination")

	ErrUnauthorized  = errors.New("unauthorized")
	ErrStakeMissing  = errors.New("stake is missing")
	ErrInvalidNodeID = errors.New("invalid node ID")
	ErrStakeNotEnded = errors.New("stake not ended")

	OutputLockupPeriodInvalid       = []byte("lockup period is invalid")
	OutputDifferentNodeIDThanStaked = []byte("node ID is different than staked")
	OutputAssetMissing              = []byte("asset missing")
	OutputSymbolEmpty               = []byte("symbol is empty")
	OutputSymbolTooLarge            = []byte("symbol is too large")
	OutputDecimalsTooLarge          = []byte("decimal is too large")
	OutputMetadataEmpty             = []byte("metadata is empty")
	OutputMetadataTooLarge          = []byte("metadata is too large")
	OutputNotWarpAsset              = []byte("not warp asset")
	OutputWrongDestination          = []byte("wrong destination")
	OutputWarpAsset                 = []byte("warp asset")
	OutputAnycast                   = []byte("anycast output")

	// import_asset.go
	OutputConflictingAsset       = []byte("warp has same asset as another")
	OutputSymbolIncorrect        = []byte("symbol is incorrect")
	OutputDecimalsIncorrect      = []byte("decimal is incorrect")
	OutputWarpVerificationFailed = []byte("warp verification failed")
	OutputInvalidDestination     = []byte("invalid destination")
	OutputMustFill               = []byte("must fill request")

	// mint_asset.go
	OutputAssetIsNative = []byte("cannot mint native asset")
	OutputWrongOwner    = []byte("wrong owner")

	// staking
	// register_validator_stake.go
	ErrNotValidator                 = errors.New("not a validator")
	ErrDifferentSignerThanActor     = errors.New("different signer than actor")
	ErrValidatorStakedAmountInvalid = errors.New("invalid stake amount")
	ErrInvalidStakeStartBlock       = errors.New("invalid stake start block")
	ErrInvalidStakeEndBlock         = errors.New("invalid stake end block")
	ErrInvalidStakeDuration         = errors.New("invalid stake duration")
	ErrInvalidDelegationFeeRate     = errors.New("delegation fee rate must be over 2 and under 100")
	ErrValidatorAlreadyRegistered   = errors.New("validator already registered for staking")
	ErrStakeNotStarted              = errors.New("stake not started")

	// delegate_user_stake.go
	ErrDelegateStakedAmountInvalid = errors.New("staked amount must be at least 25 NAI")
	ErrUserAlreadyStaked           = errors.New("user already staked")
	ErrValidatorNotYetRegistered   = errors.New("validator not yet registered for staking")
)
