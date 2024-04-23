// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/cmd/nuklai-feed/manager"
)

type Manager interface {
	GetFeedInfo(context.Context) (codec.Address, uint64, error)
	GetFeed(context.Context, string, string) ([]*manager.FeedObject, error)
	UpdateNuklaiRPC(context.Context, string) error
}
