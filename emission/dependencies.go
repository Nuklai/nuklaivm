// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"github.com/ava-labs/avalanchego/utils/logging"
)

type Controller interface {
	Logger() logging.Logger
}
