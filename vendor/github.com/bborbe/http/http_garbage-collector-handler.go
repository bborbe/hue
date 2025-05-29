// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
)

func NewGarbageCollectorHandler() http.Handler {
	asMegabyte := func(b uint64) uint64 {
		return b / 1024 / 1024
	}
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		printMemStats := func(stage string) {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Fprintf(resp, "Memory Stats %s:\n", stage)
			fmt.Fprintf(resp, "  Allocated (used) memory: %d MB\n", asMegabyte(m.Alloc))
			fmt.Fprintf(resp, "  Total memory obtained from OS (reserved): %d MB\n", asMegabyte(m.Sys))
			fmt.Fprintf(resp, "  Heap in use: %d MB\n", asMegabyte(m.HeapInuse))
			fmt.Fprintf(resp, "  Heap released to OS: %d MB\n", asMegabyte(m.HeapReleased))
			fmt.Fprintf(resp, "  Heap idle (could be released): %d MB\n", asMegabyte(m.HeapIdle))
		}

		printMemStats("Before GC")

		runtime.GC()
		printMemStats("After GC")

		debug.FreeOSMemory()
		printMemStats("After FreeOSMemory")
	})
}
