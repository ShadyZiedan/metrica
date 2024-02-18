package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/utils"
)

type pgConnWrapper struct {
	conn pgConn
}

func (p *pgConnWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (res pgconn.CommandTag, err error) {
	err = utils.RetryWithBackoff(3, isConnectionError, func() error {
		res, err = p.conn.Exec(ctx, sql, arguments...)
		return err
	})
	return
}

func isConnectionError(err error) bool {
	if err != nil {
		logger.Log.Info("Error getting connection", zap.Error(err))
	}
	return err != nil
}

func (p *pgConnWrapper) Query(ctx context.Context, sql string, args ...interface{}) (res pgx.Rows, err error) {
	err = utils.RetryWithBackoff(3, isConnectionError, func() error {
		res, err = p.conn.Query(ctx, sql, args...)
		return err
	})
	return
}

func (p *pgConnWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.conn.QueryRow(ctx, sql, args...)
}

func (p *pgConnWrapper) Begin(ctx context.Context) (tx pgx.Tx, err error) {
	err = utils.RetryWithBackoff(3, isConnectionError, func() error {
		tx, err = p.conn.Begin(ctx)
		return err
	})
	return
}
