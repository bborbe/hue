package trigger

import "context"

//go:generate counterfeiter -o ../../mocks/trigger.go --fake-name Trigger . Trigger
type Trigger interface {
	Trigger(ctx context.Context, ch chan<- struct{}) error
}
