package executor

import (
	"context"
	"database/sql"
	"errors"
	"github.com/wb-go/wbf/dbpg"
)

type (
	txKey    struct{}
	Executor struct {
		*dbpg.DB
	}
	IExecutor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
		PrepareContext(context.Context, string) (*sql.Stmt, error)
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
)

func New(conn *dbpg.DB) *Executor {
	return &Executor{DB: conn}
}

func (e *Executor) injectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func (e *Executor) extractTx(ctx context.Context) *sql.Tx {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok {
		return nil
	}
	return tx
}

func (e *Executor) WithTransaction(
	ctx context.Context,
	txOpts *sql.TxOptions,
	fn func(ctx context.Context) error,
) error {
	tx, err := e.DB.Master.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	ctxWithTx := e.injectTx(ctx, tx)
	if err = fn(ctxWithTx); err != nil {
		return err
	}

	return tx.Commit()
}

func (e *Executor) GetExecutor(ctx context.Context) IExecutor {
	tx := e.extractTx(ctx)
	if tx != nil {
		return tx
	}
	return e.Master // e.Master == *sql.DB
}

var errNoTxInCtx = errors.New("no tx in context")

func (e *Executor) GetTx(ctx context.Context) (*sql.Tx, error) {
	tx := e.extractTx(ctx)
	if tx != nil {
		return tx, nil
	}
	return nil, errNoTxInCtx
}

func (e *Executor) IsTx(exec IExecutor) bool {
	if exec == nil {
		return false
	}
	_, ok := exec.(*sql.Tx)
	return ok
}
