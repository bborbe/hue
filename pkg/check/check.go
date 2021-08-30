// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import "context"

type Checks []Check

//go:generate counterfeiter -o ../../mocks/check.go --fake-name Check . Check

// Check is something is in the right state
type Check interface {
	Apply(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
	Name() string
}
