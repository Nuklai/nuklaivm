// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import "errors"

var (
	ErrValidatorAlreadyRegistered = errors.New("validator already registered")
	ErrStakedAmountInvalid        = errors.New("staked amount invalid")
	ErrValidatorNotFound          = errors.New("validator not found")
	ErrInvalidStakeDuration       = errors.New("invalid stake duration")
	ErrInsufficientRewards        = errors.New("insufficient rewards")
	ErrStakeExpired               = errors.New("stake expired")
	ErrDelegatorAlreadyStaked     = errors.New("delegator already staked")
	ErrDelegatorNotFound          = errors.New("delegator not found")
	ErrInvalidBlockHeight         = errors.New("invalid block height")
	ErrValidatorNotActive         = errors.New("validator not active")

	ErrInvalidNodeID      = errors.New("invalid node id")
	ErrStakeNotFound      = errors.New("stake not found")
	ErrNotAValidator      = errors.New("not a validator")
	ErrNotAValidatorOwner = errors.New("not a validator owner")
)
