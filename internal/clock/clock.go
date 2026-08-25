package clock

import (
	"context"
	"time"
)

type Clock interface{ Now() time.Time }
type Real struct{}

func (Real) Now() time.Time { return time.Now().UTC() }

type Fixed struct{ T time.Time }

func (f Fixed) Now() time.Time { return f.T }

type contextKey struct{}

func With(ctx context.Context, c Clock) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}
func From(ctx context.Context) Clock {
	if c, ok := ctx.Value(contextKey{}).(Clock); ok {
		return c
	}
	return Real{}
}
