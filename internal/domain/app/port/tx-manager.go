package port

import (
	"context"
	"database/sql"
)

type TxManager interface {
	WithTransaction(
		ctx context.Context,
		txOpts *sql.TxOptions,
		fn func(ctx context.Context) error,
	) error
}
