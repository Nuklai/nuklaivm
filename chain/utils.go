// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/binary"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
)

func GenerateIDWithIndex(id ids.ID, i uint64) ids.ID {
	actionBytes := make([]byte, ids.IDLen+consts.Uint64Len)
	copy(actionBytes, id[:])
	binary.BigEndian.PutUint64(actionBytes[ids.IDLen:], i)
	return utils.ToID(actionBytes)
}

func GenerateIDWithAddress(id ids.ID, addr codec.Address) ids.ID {
	actionBytes := make([]byte, ids.IDLen+codec.AddressLen)
	copy(actionBytes, id[:])
	copy(actionBytes[ids.IDLen:], addr[:])
	return utils.ToID(actionBytes)
}
