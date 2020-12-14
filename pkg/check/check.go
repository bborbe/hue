package check

import "context"

type Checks []Check

//go:generate counterfeiter -o ../../mocks/check.go --fake-name Check . Check
type Check interface {
	Apply(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
	Name() string
}
