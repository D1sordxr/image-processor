package app

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/app/port"
)

type App struct {
	log        port.Logger
	components []port.Component
}

func NewApp(
	log port.Logger,
	components ...port.Component,
) *App {
	return &App{
		log:        log,
		components: components,
	}
}

func (a *App) Run(ctx context.Context) {
	defer a.shutdown()

	errChan := make(chan error, 1)
	errGroup, ctx := errgroup.WithContext(ctx)
	go func() { errChan <- errGroup.Wait() }()

	for i, c := range a.components {
		idx, component := i, c
		errGroup.Go(func() error {
			a.log.Info("Starting component",
				"idx", idx,
				"type", fmt.Sprintf("%T", component),
			)
			startErr := c.Run(ctx)
			if startErr != nil {
				a.log.Error("Component failed",
					"idx", idx,
					"type", fmt.Sprintf("%T", component),
					"error", startErr.Error(),
				)
			}
			return startErr
		})
	}

	select {
	case err := <-errChan:
		a.log.Error("App received an error", "error", err.Error())
	case <-ctx.Done():
		a.log.Info("App received a terminate signal")
	}
}

func (a *App) shutdown() {
	a.log.Info("App shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	errs := make([]error, 0, len(a.components))
	for i := len(a.components) - 1; i >= 0; i-- {
		a.log.Info("Shutting down Component", "idx", i)
		if err := a.components[i].Shutdown(shutdownCtx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		a.log.Info("App successfully shutdown")
	} else {
		a.log.Error(
			"App shutdown with errors",
			"errors", errors.Join(errs...).Error(),
		)
	}
}
