package trigger

import "context"

func NewFunc(f func(ctx context.Context, ch chan<- struct{}) error) Trigger {
	return fn(f)
}

type fn func(ctx context.Context, ch chan<- struct{}) error

func (f fn) Trigger(ctx context.Context, ch chan<- struct{}) error {
	return f(ctx, ch)
}
