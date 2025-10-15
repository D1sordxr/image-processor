package txmanager

import (
	"context"
	"database/sql"

	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/executor"
)

type (
	TxManager struct {
		executor *executor.Executor
	}
)

func New(executor *executor.Executor) *TxManager {
	return &TxManager{executor: executor}
}

func (t *TxManager) WithTransaction(
	ctx context.Context,
	txOpts *sql.TxOptions,
	fn func(ctx context.Context) error,
) error {
	return t.executor.WithTransaction(ctx, txOpts, fn)
}
