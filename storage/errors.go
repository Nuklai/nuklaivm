// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import "errors"

var (
	ErrInvalidAddress           = errors.New("invalid address")
	ErrInvalidBalance           = errors.New("invalid balance")
	ErrMaxSupplyExceeded        = errors.New("max supply exceeded")
	ErrInsufficientAssetBalance = errors.New("insufficient asset balance")
)
