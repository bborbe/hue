package trigger

import (
	"context"
)

func NewRightAway() Trigger {
	return fn(func(ctx context.Context, ch chan<- struct{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- struct{}{}:
			return nil
		}
	})
}
