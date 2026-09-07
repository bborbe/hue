// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg/handler"
)

// NewSetLogLevelHandler is exercised through a real mux.Router so the
// mux.Vars path-variable seam (the boundary production traffic crosses) is
// covered. The 20ms auto-reset TTL keeps the tests fast; Eventually /
// Consistently absorb goroutine-scheduling jitter around the reset.
var _ = Describe("SetLogLevelHandler", func() {
	var (
		levelVar *slog.LevelVar
		router   *mux.Router
		recorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		levelVar = &slog.LevelVar{}
		levelVar.Set(slog.LevelDebug)
		router = mux.NewRouter()
		router.Path("/setloglevel/{level}").
			Handler(handler.NewSetLogLevelHandler(levelVar, slog.LevelDebug, 20*time.Millisecond))
		recorder = httptest.NewRecorder()
	})

	It("maps level 0 to Info and level 1 to Debug", func() {
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/setloglevel/0", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(levelVar.Level()).To(Equal(slog.LevelInfo))

		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/setloglevel/1", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(levelVar.Level()).To(Equal(slog.LevelDebug))
	})

	It("rejects invalid levels without changing the level", func() {
		for _, level := range []string{"abc", "-1", "99"} {
			router.ServeHTTP(
				recorder,
				httptest.NewRequest(http.MethodGet, "/setloglevel/"+level, nil),
			)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(levelVar.Level()).To(Equal(slog.LevelDebug))
		}
	})

	It("resets to the default level after the auto-reset duration", func() {
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/setloglevel/0", nil))
		Expect(levelVar.Level()).To(Equal(slog.LevelInfo))

		Eventually(levelVar.Level, 5*time.Second, 10*time.Millisecond).
			Should(Equal(slog.LevelDebug), "level should auto-reset to Debug")
	})

	It("keeps a re-set level past the auto-reset", func() {
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/setloglevel/0", nil))
		// Separate the two sets by more than timer wakeup latency so the
		// first reset goroutine reliably wakes within the TTL of the re-set
		// and its lastSetTime guard skips the reset (level stays Debug).
		time.Sleep(5 * time.Millisecond)
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/setloglevel/1", nil))

		Consistently(levelVar.Level, 100*time.Millisecond, 10*time.Millisecond).
			Should(Equal(slog.LevelDebug), "level stays at the re-set Debug")
	})
})
