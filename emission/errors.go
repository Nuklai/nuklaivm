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

	ErrInvalidNodeID      = errors.New("invalid node id")
	ErrStakeNotFound      = errors.New("stake not found")
	ErrNotAValidator      = errors.New("not a validator")
	ErrNotAValidatorOwner = errors.New("not a validator owner")
)
