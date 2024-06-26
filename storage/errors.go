// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import "errors"

var (
	ErrInvalidBalance = errors.New("invalid balance")
	ErrInvalidStake   = errors.New("invalid stake")
	ErrStakeNotFound  = errors.New("stake not found")
)
