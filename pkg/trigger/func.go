// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trigger

import "context"

func NewFunc(f func(ctx context.Context, ch chan<- struct{}) error) Trigger {
	return fn(f)
}

type fn func(ctx context.Context, ch chan<- struct{}) error

func (f fn) Trigger(ctx context.Context, ch chan<- struct{}) error {
	return f(ctx, ch)
}
