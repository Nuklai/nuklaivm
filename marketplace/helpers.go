// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"sync"
)

var (
	once        sync.Once
	marketplace Hub
)
