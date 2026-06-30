// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"context"

	"github.com/amimof/huego"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)

var _ = Describe("CheckCreator", func() {
	var (
		ctx      context.Context
		provider pkg.BridgesProvider
	)

	BeforeEach(func() {
		ctx = context.Background()
		provider = pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return []*huego.Bridge{nil}, nil
			},
		)
	})

	DescribeTable("aquarium window",
		func(summerMode bool) {
			creator := check.NewCheckCreator(provider, summerMode)
			checks, err := creator.CreateChecks(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(checks).To(HaveLen(9))
		},
		Entry("summer mode disabled", false),
		Entry("summer mode enabled", true),
	)
})
