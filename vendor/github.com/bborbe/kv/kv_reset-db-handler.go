// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/glog"
)

// NewResetHandler returns a http.Handler
// that allow delete the complete database
func NewResetHandler(db DB, cancel context.CancelFunc) http.Handler {
	var lock sync.Mutex
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if lock.TryLock() == false {
			glog.V(2).Infof("reset db already running")
			http.Error(resp, fmt.Sprintf("reset db already running"), http.StatusInternalServerError)
			return
		}
		defer lock.Unlock()
		glog.V(2).Infof("reset db started")

		defer cancel()
		if err := db.Close(); err != nil {
			http.Error(resp, fmt.Sprintf("reset db failed: %v", err), http.StatusInternalServerError)
			return
		}
		if err := db.Remove(); err != nil {
			http.Error(resp, fmt.Sprintf("remove db failed: %v", err), http.StatusInternalServerError)
			return
		}
		resp.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(resp, "reset db successful")
		glog.V(2).Infof("reset db successful")
	})
}
