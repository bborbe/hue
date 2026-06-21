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

// NewListLightsHandler is exercised at the provider-glue boundary. See
// status_test.go for why the huego.Bridge happy-path is not unit-tested.
var _ = Describe("ListLightsHandler", func() {
	var (
		ctx      context.Context
		recorder *httptest.ResponseRecorder
		req      *http.Request
	)

	BeforeEach(func() {
		ctx = context.Background()
		recorder = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/lights", nil)
	})

	It("returns an error when no bridges are available", func() {
		provider := pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return nil, nil
			},
		)

		err := handler.NewListLightsHandler(provider).ServeHTTP(ctx, recorder, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no bridges available"))
	})

	It("propagates errors from the bridges provider", func() {
		provider := pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return nil, stderrors.New("bridge offline")
			},
		)

		err := handler.NewListLightsHandler(provider).ServeHTTP(ctx, recorder, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("get bridges failed"))
		Expect(err.Error()).To(ContainSubstring("bridge offline"))
	})
})
