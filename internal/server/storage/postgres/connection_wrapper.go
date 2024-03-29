package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/shadyziedan/metrica/internal/retry"
)

type pgConnWrapper struct {
	conn pgConn
}

func (p *pgConnWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (res pgconn.CommandTag, err error) {
	err = retry.WithBackoff(ctx, 3, isConnectionError, func() error {
		res, err = p.conn.Exec(ctx, sql, arguments...)
		return err
	})
	return
}

func isConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "08")
}

func (p *pgConnWrapper) Query(ctx context.Context, sql string, args ...interface{}) (res pgx.Rows, err error) {
	err = retry.WithBackoff(ctx, 3, isConnectionError, func() error {
		res, err = p.conn.Query(ctx, sql, args...)
		return err
	})
	return
}

func (p *pgConnWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.conn.QueryRow(ctx, sql, args...)
}

func (p *pgConnWrapper) Begin(ctx context.Context) (tx pgx.Tx, err error) {
	err = retry.WithBackoff(ctx, 3, isConnectionError, func() error {
		tx, err = p.conn.Begin(ctx)
		return err
	})
	return
}
