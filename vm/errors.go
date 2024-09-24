// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import "errors"

var (
	ErrAssetNotFound   = errors.New("asset not found")
	ErrDatasetNotFound = errors.New("dataset not found")
)
