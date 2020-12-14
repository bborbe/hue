package check

import "context"

func Func(
	name string,
	satisfied func(ctx context.Context) (bool, error),
	apply func(ctx context.Context) error,
) Check {
	return fn{
		name:      name,
		apply:     apply,
		satisfied: satisfied,
	}
}

type fn struct {
	name      string
	apply     func(ctx context.Context) error
	satisfied func(ctx context.Context) (bool, error)
}

func (f fn) Name() string {
	return f.name
}

func (f fn) Apply(ctx context.Context) error {
	return f.apply(ctx)
}

func (f fn) Satisfied(ctx context.Context) (bool, error) {
	return f.satisfied(ctx)
}
