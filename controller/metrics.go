// Copyright (C) 2023, AllianceBlock. All rights reserved.
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

	stake prometheus.Counter

	mintNAI prometheus.Counter
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

		mintNAI: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "mintNAI",
			Help:      "number of NAI tokens minted",
		}),
	}
	r := prometheus.NewRegistry()
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.transfer),

		r.Register(m.stake),

		r.Register(m.mintNAI),

		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}
