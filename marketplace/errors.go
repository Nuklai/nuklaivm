// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import "errors"

var (
	ErrDataAlreadyAdded                = errors.New("data already added")
	ErrAlreadyContributedToThisDataset = errors.New("already contributed to this dataset")
	ErrContributionNotFound            = errors.New("contribution not found")
	ErrDatasetNotFound                 = errors.New("dataset not found")
)
