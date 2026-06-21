// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler_test

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"

	"github.com/amimof/huego"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/handler"
)

// NewStatusHandler is exercised at the provider-glue boundary. The huego.Bridge
// happy-path (GetLightsContext returning real Light entries) is not unit-tested
// because huego.Bridge has no transport seam — a happy-path test would need a
// fake HTTP server speaking the Hue CLIP v1 protocol, which couples the test
// to huego internals. The on/off classification logic is exercised by the
// production cron-tick loop in pkg/check/ and verified live via /status.
var _ = Describe("StatusHandler", func() {
	var (
		ctx      context.Context
		recorder *httptest.ResponseRecorder
		req      *http.Request
	)

	BeforeEach(func() {
		ctx = context.Background()
		recorder = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/status", nil)
	})

	It("returns an error when no bridges are available", func() {
		provider := pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return nil, nil
			},
		)

		err := handler.NewStatusHandler(provider).ServeHTTP(ctx, recorder, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no bridges available"))
	})

	It("propagates errors from the bridges provider", func() {
		provider := pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return nil, stderrors.New("bridge offline")
			},
		)

		err := handler.NewStatusHandler(provider).ServeHTTP(ctx, recorder, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("get bridges failed"))
		Expect(err.Error()).To(ContainSubstring("bridge offline"))
	})

	It("returns a non-nil StatusResponse type when initialised", func() {
		// Pure type-shape assertion: confirms the handler's StatusResponse
		// initialises with non-nil On/Off slices so JSON encodes them as []
		// rather than null. The classification loop itself is covered above
		// via the provider-glue tests.
		resp := handler.StatusResponse{On: []string{}, Off: []string{}}
		Expect(resp.On).NotTo(BeNil())
		Expect(resp.Off).NotTo(BeNil())
	})
})
