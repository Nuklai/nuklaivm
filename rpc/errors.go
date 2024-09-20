// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import "errors"

var (
	ErrTxNotFound    = errors.New("tx not found")
	ErrAssetNotFound = errors.New("asset not found")

	// register_validator_stake
	ErrValidatorStakeNotFound = errors.New("validator stake not found")

	// delegate_user_stake
	ErrUserStakeNotFound = errors.New("user stake not found")

	ErrDatasetNotFound = errors.New("dataset not found")
)
