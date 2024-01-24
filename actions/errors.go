// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import "errors"

var (
	ErrFieldNotPopulated = []byte("field is not populated")
	ErrNoSwapToFill      = errors.New("no swap to fill")
)
