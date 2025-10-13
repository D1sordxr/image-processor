package port

import "context"

type (
	RunFunc      func(ctx context.Context) error
	ShutdownFunc func(ctx context.Context) error
)
