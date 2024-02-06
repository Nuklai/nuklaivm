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
	transfer prometheus.Counter

	stake   prometheus.Counter
	unstake prometheus.Counter

	mintNAI prometheus.Counter

	createAsset prometheus.Counter
	mintAsset   prometheus.Counter
	burnAsset   prometheus.Counter

	importAsset prometheus.Counter
	exportAsset prometheus.Counter
}

func newMetrics(gatherer ametrics.MultiGatherer) (*metrics, error) {
	m := &metrics{
		transfer: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "transfer",
			Help:      "number of transfer actions",
		}),

		stake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "stake",
			Help:      "number of stake actions",
		}),
		unstake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "unstake",
			Help:      "number of unstake actions",
		}),

		mintNAI: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "mintNAI",
			Help:      "number of NAI tokens minted",
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
	}
	r := prometheus.NewRegistry()
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.transfer),

		r.Register(m.stake),
		r.Register(m.unstake),

		r.Register(m.mintNAI),

		r.Register(m.createAsset),
		r.Register(m.mintAsset),
		r.Register(m.burnAsset),

		r.Register(m.importAsset),
		r.Register(m.exportAsset),

		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}
