// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	feesDistributed prometheus.Counter
	mintedNAI       prometheus.Counter

	transfer prometheus.Counter

	createAsset prometheus.Counter
	mintAsset   prometheus.Counter
	burnAsset   prometheus.Counter
	importAsset prometheus.Counter
	exportAsset prometheus.Counter

	stakeAmount            prometheus.Gauge
	registerValidatorStake prometheus.Counter
	withdrawValidatorStake prometheus.Counter
	delegateUserStake      prometheus.Counter
	undelegateUserStake    prometheus.Counter
	claimStakingRewards    prometheus.Counter
}

func newMetrics(gatherer ametrics.MultiGatherer) (*metrics, error) {
	m := &metrics{
		feesDistributed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "feesDistributed",
			Help:      "number of NAI tokens distributed as fees",
		}),
		mintedNAI: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "mintedNAI",
			Help:      "number of NAI tokens minted",
		}),

		transfer: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "transfer",
			Help:      "number of transfer actions",
		}),

		createAsset: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "create_asset",
			Help:      "number of create asset actions",
		}),
		mintAsset: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "mint_asset",
			Help:      "number of mint asset actions",
		}),
		burnAsset: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "burn_asset",
			Help:      "number of burn asset actions",
		}),
		importAsset: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "import_asset",
			Help:      "number of import asset actions",
		}),
		exportAsset: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "export_asset",
			Help:      "number of export asset actions",
		}),

		stakeAmount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "actions",
			Name:      "stake_amount",
			Help:      "amount of staked tokens",
		}),
		registerValidatorStake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "register_validator_stake",
			Help:      "number of register validator stake actions",
		}),
		withdrawValidatorStake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "withdraw_validator_stake",
			Help:      "number of withdraw validator stake actions",
		}),
		delegateUserStake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "delegate_user_stake",
			Help:      "number of delegate user stake actions",
		}),
		undelegateUserStake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "undelegate_user_stake",
			Help:      "number of undelegate user stake actions",
		}),
		claimStakingRewards: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "claim_staking_rewards",
			Help:      "number of claim staking rewards actions",
		}),
	}
	r := prometheus.NewRegistry()
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.feesDistributed),
		r.Register(m.mintedNAI),

		r.Register(m.transfer),

		r.Register(m.createAsset),
		r.Register(m.mintAsset),
		r.Register(m.burnAsset),
		r.Register(m.importAsset),
		r.Register(m.exportAsset),

		r.Register(m.stakeAmount),
		r.Register(m.registerValidatorStake),
		r.Register(m.withdrawValidatorStake),
		r.Register(m.delegateUserStake),
		r.Register(m.undelegateUserStake),
		r.Register(m.claimStakingRewards),

		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}
