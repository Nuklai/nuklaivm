// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import "errors"

var (
	ErrAssetNotFound          = errors.New("asset not found")
	ErrDatasetNotFound        = errors.New("dataset not found")
	ErrValidatorStakeNotFound = errors.New("validator stake not found")
	ErrUserStakeNotFound      = errors.New("user stake not found")
)
