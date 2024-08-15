// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/binary"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
)

func GenerateID(id ids.ID, i uint64) ids.ID {
	actionBytes := make([]byte, ids.IDLen+consts.Uint64Len)
	copy(actionBytes, id[:])
	binary.BigEndian.PutUint64(actionBytes[ids.IDLen:], i)
	return utils.ToID(actionBytes)
}
