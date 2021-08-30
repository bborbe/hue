// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"flag"
	"runtime"

	"github.com/bborbe/argument"
	"github.com/golang/glog"
)

// Application to run
type Application interface {
	Run(ctx context.Context) error
}

// Main function for all main.go
func Main(ctx context.Context, app Application) int {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	_ = flag.Set("logtostderr", "true")

	if err := argument.Parse(app); err != nil {
		glog.Errorf("parse app failed: %v", err)
		return 4
	}

	glog.V(0).Infof("application started")
	if err := app.Run(contextWithSig(ctx)); err != nil {
		glog.Error(err)
		return 1
	}
	glog.V(0).Infof("application finished")
	return 0
}
