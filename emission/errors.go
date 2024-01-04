// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import "errors"

var (
	ErrMaxValidatorsReached     = errors.New("max validators reached")
	ErrValidatorNotFound        = errors.New("validator not found")
	ErrStakeNotFound            = errors.New("stake not found")
	ErrInvalidNodeID            = errors.New("invalid node id")
	ErrMinStakeAmountNotReached = errors.New("min stake amount not reached")
	ErrNotAValidator            = errors.New("not a validator")
	ErrNotAValidatorOwner       = errors.New("not a validator owner")
)
