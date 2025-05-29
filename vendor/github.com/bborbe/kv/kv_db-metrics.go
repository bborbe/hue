// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
)

func NewDBWithMetrics(
	db DB,
	metrics Metrics,
) DB {
	return &dbWithMetrics{
		db:      db,
		metrics: metrics,
	}
}

type dbWithMetrics struct {
	metrics Metrics
	db      DB
}

func (d *dbWithMetrics) Update(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
	d.metrics.DbUpdateInc()
	return d.db.Update(ctx, func(ctx context.Context, tx Tx) error {
		return fn(ctx, tx)
	})
}

func (d *dbWithMetrics) View(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
	d.metrics.DbViewInc()
	return d.db.View(ctx, func(ctx context.Context, tx Tx) error {
		return fn(ctx, tx)
	})
}

func (d *dbWithMetrics) Sync() error {
	return d.db.Sync()
}

func (d *dbWithMetrics) Close() error {
	return d.db.Close()
}

func (d *dbWithMetrics) Remove() error {
	return d.db.Remove()
}
