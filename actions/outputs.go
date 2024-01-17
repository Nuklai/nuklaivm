// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

var (
	OutputValueZero = []byte("value is zero")

	OutputStakedAmountZero          = []byte("staked amount is zero")
	OutputLockupPeriodInvalid       = []byte("lockup period is invalid")
	OutputStakeMissing              = []byte("stake is missing")
	OutputUnauthorized              = []byte("unauthorized")
	OutputInvalidNodeID             = []byte("invalid node ID")
	OutputDifferentNodeIDThanStaked = []byte("node ID is different than staked")
)
