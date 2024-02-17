// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package emission

import (
	"time"

	hutils "github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

type RewardConfig struct {
	// MintingPeriod is period that the staking calculator runs on. It is
	// not valid for a validator's stake duration to be larger than this.
	MintingPeriod time.Duration `json:"mintingPeriod"`

	// SupplyCap is the target value that the reward calculation should be
	// asymptotic to.
	SupplyCap uint64 `json:"supplyCap"`
}

type StakingConfig struct {
	// Staking uptime requirements
	UptimeRequirement uint64 `json:"uptimeRequirement"`
	// Minimum stake, in NAI, required to validate the nuklai network
	MinValidatorStake uint64 `json:"minValidatorStake"`
	// Maximum stake, in NAI, allowed to be placed on a single validator in
	// the nuklai network
	MaxValidatorStake uint64 `json:"maxValidatorStake"`
	// Minimum stake, in NAI, that can be delegated on the nuklai network
	MinDelegatorStake uint64 `json:"minDelegatorStake"`
	// Minimum delegation fee, in the range [0, 100], that can be charged
	// for delegation on the nuklai network.
	MinDelegationFee uint64 `json:"minDelegationFee"`
	// MinValidatorStakeDuration is the minimum amount of time a validator can validate
	// for in a single period.
	MinValidatorStakeDuration time.Duration `json:"minValidatorStakeDuration"`
	// MaxStakeDuration is the maximum amount of time a validator can validate
	// for in a single period.
	MaxValidatorStakeDuration time.Duration `json:"maxValidatorStakeDuration"`
	// MinDelegatorStakeDuration is the minimum amount of time a user can delegate
	// for in a single period.
	MinDelegatorStakeDuration time.Duration `json:"minDelegatorStakeDuration"`
	// RewardConfig is the config for the reward function.
	RewardConfig RewardConfig `json:"rewardConfig"`
}

func GetStakingConfig() StakingConfig {
	// TODO: Enable this in production
	// minValidatorStake, _ := hutils.ParseBalance("1500000", nconsts.Decimals)
	minValidatorStake, _ := hutils.ParseBalance("100", nconsts.Decimals)
	maxValidatorStake, _ := hutils.ParseBalance("1000000000", nconsts.Decimals)
	minDelegatorStake, _ := hutils.ParseBalance("25", nconsts.Decimals)
	supplyCap, _ := hutils.ParseBalance("10000000000", nconsts.Decimals)
	return StakingConfig{
		UptimeRequirement:         80, // 80%
		MinValidatorStake:         minValidatorStake,
		MaxValidatorStake:         maxValidatorStake,
		MinDelegatorStake:         minDelegatorStake,
		MinDelegationFee:          2,               // 2%
		MinValidatorStakeDuration: 1 * time.Minute, // 1 minute
		// MinStakeDuration:  6 * 4 * 7 * 24 * time.Hour, // 6 months TODO: Enable this in production
		MaxValidatorStakeDuration: 365 * 24 * time.Hour, // 1 year,
		MinDelegatorStakeDuration: 1 * time.Minute,      // 1 minute
		// MinStakeDuration:  2 * 7 * 24 * time.Hour, // 2 weeks TODO: Enable this in production
		RewardConfig: RewardConfig{
			MintingPeriod: 365 * 24 * time.Hour,
			SupplyCap:     supplyCap,
		},
	}
}
