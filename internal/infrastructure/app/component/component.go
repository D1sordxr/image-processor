package component

import (
	"context"
	"github.com/D1sordxr/image-processor/internal/domain/app/port"
)

type Functional struct {
	runFunc      port.RunFunc
	shutdownFunc port.ShutdownFunc
}

func NewFunctional(
	runFunc port.RunFunc,
	shutdownFunc port.ShutdownFunc,
) *Functional {
	return &Functional{
		runFunc:      runFunc,
		shutdownFunc: shutdownFunc,
	}
}

func (c *Functional) Run(ctx context.Context) error {
	return c.runFunc(ctx)
}

func (c *Functional) Shutdown(ctx context.Context) error {
	return c.shutdownFunc(ctx)
}
