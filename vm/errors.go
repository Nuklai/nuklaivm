// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import "errors"

var (
	ErrValidatorStakeNotFound = errors.New("validator stake not found")
	ErrDelegatorStakeNotFound = errors.New("delegator stake not found")
)
