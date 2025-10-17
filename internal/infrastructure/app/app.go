package app

import (
	"context"
	"errors"
	"fmt"
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

	errChan := make(chan error)
	for i, c := range a.components {
		func(idx int, component port.Component) {
			var (
				startCtx, startCancel = context.WithTimeout(ctx, 5*time.Second)
				startErrChan          = make(chan error, 1)
				startErr              error
			)
			defer startCancel()
			defer close(startErrChan)

			go func() {
				a.log.Info("Starting Component",
					"idx", idx,
					"type", fmt.Sprintf("%T", c),
				)
				if startErr = c.Run(ctx); startErr != nil {
					a.log.Error("Component failed",
						"idx", idx,
						"type", fmt.Sprintf("%T", c),
						"error", startErr.Error(),
					)
					startErrChan <- startErr
				}
			}()

			select {
			case err := <-startErrChan:
				errChan <- err
				break
			case <-startCtx.Done():
			}
		}(i, c)
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
