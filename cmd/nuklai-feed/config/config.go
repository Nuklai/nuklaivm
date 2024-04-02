// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/consts"
)

type Config struct {
	HTTPHost string `json:"host"`
	HTTPPort int    `json:"port"`

	NuklaiRPC string `json:"nuklaiRPC"`

	Recipient     string `json:"recipient"`
	recipientAddr codec.Address

	FeedSize               int    `json:"feedSize"`
	MinFee                 uint64 `json:"minFee"`
	FeeDelta               uint64 `json:"feeDelta"`
	MessagesPerEpoch       int    `json:"messagesPerEpoch"`
	TargetDurationPerEpoch int64  `json:"targetDurationPerEpoch"` // seconds
}

func (c *Config) RecipientAddress() (codec.Address, error) {
	if c.recipientAddr != codec.EmptyAddress {
		return c.recipientAddr, nil
	}
	addr, err := codec.ParseAddressBech32(consts.HRP, c.Recipient)
	if err == nil {
		c.recipientAddr = addr
	}
	return addr, err
}
