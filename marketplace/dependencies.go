// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package marketplace

import (
	"github.com/ava-labs/avalanchego/utils/logging"
)

type Controller interface {
	Logger() logging.Logger
}
