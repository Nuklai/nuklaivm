// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

// CommonResult holds the shared fields for action results
type CommonResult struct {
	Actor    string `serialize:"true" json:"actor"`
	Receiver string `serialize:"true" json:"receiver"`
}

// FillCommonResult fills the common fields for action results.
func FillCommonResult(actor, receiver string) CommonResult {
	return CommonResult{
		Actor:    actor,
		Receiver: receiver,
	}
}
