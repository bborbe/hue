// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	DbUpdateInc()
	DbViewInc()
}

func NewMetrics() Metrics {
	return &metrics{}
}

type metrics struct {
}

func (m *metrics) DbUpdateInc() {
	dbUpdateCounter.Inc()
}

func (m *metrics) DbViewInc() {
	dbViewCounter.Inc()
}

var (
	dbUpdateCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "kv",
		Subsystem: "db",
		Name:      "update",
		Help:      "Counts db updates",
	})
	dbViewCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "kv",
		Subsystem: "db",
		Name:      "view",
		Help:      "Counts db views",
	})
)

func init() {
	prometheus.MustRegister(
		dbUpdateCounter,
		dbViewCounter,
	)
}
