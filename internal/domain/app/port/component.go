package port

import "context"

type (
	RunFunc      func(ctx context.Context) error
	ShutdownFunc func(ctx context.Context) error
)


type Component interface {
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
