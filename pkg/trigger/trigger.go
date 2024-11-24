// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trigger

import "context"

//counterfeiter:generate -o ../../mocks/trigger.go --fake-name Trigger . Trigger
type Trigger interface {
	Trigger(ctx context.Context, ch chan<- struct{}) error
}
