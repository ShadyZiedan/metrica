package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

type DBStorage struct {
	conn      pgConn
	observers []storage.MetricsObserver
}

type pgConn interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewDBStorage(conn pgConn) (*DBStorage, error) {
	postgresConn := &pgConnWrapper{conn: conn}
	_, err := postgresConn.Exec(context.Background(), `create table if not exists metrics
(
    id      serial,
    name    varchar not null,
    m_type  varchar not null,
    counter bigint,
    gauge   decimal
)`)
	if err != nil {
		return nil, err
	}
	return &DBStorage{conn: postgresConn}, nil
}

func (db *DBStorage) Find(ctx context.Context, name string) (*models.Metric, error) {
	row := db.conn.QueryRow(ctx, `SELECT name, m_type, gauge, counter FROM metrics WHERE name = $1;`, name)
	var metric models.Metric
	err := row.Scan(&metric.Name, &metric.MType, &metric.Gauge, &metric.Counter)
	if err != nil {
		return nil, err
	}
	return &metric, err
}

func (db *DBStorage) Create(ctx context.Context, name string, mType string) error {
	_, err := db.conn.Exec(ctx, `INSERT INTO metrics (name, m_type) values ($1, $2)`, name, mType)
	if err != nil {
		return err
	}
	return nil
}

func (db *DBStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var prevCounter int64
	row := tx.QueryRow(ctx, `SELECT coalesce(counter, 0) FROM metrics where name = $1`, name)
	err = row.Scan(&prevCounter)
	if err != nil {
		logger.Log.Error("Error getting counter", zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE metrics set counter = $1 where name = $2`, prevCounter+delta, name)
	if err != nil {
		logger.Log.Error("Error updating counter", zap.Error(err))
		return err
	}
	return tx.Commit(ctx)
}

func (db *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := db.conn.Exec(ctx, `UPDATE metrics set gauge = $1 where name = $2`, value, name)
	if err != nil {
		return err
	}
	return nil
}

func (db *DBStorage) FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error) {
	metric, err := db.Find(ctx, name)
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			logger.Log.Error("find or create", zap.Error(err))
		}
	}()
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if err == nil {
		return metric, nil
	}
	err = db.Create(ctx, name, mType)
	if err != nil {
		return nil, err
	}
	metric, err = db.Find(ctx, name)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (db *DBStorage) FindAll(ctx context.Context) ([]*models.Metric, error) {
	rows, err := db.conn.Query(ctx, `SELECT name, m_type, gauge, counter FROM metrics`)
	if err != nil {
		return nil, err
	}
	var metrics []*models.Metric

	for rows.Next() {
		metric := &models.Metric{}
		err := rows.Scan(&metric.Name, &metric.MType, &metric.Gauge, &metric.Counter)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, err
}

func (db *DBStorage) Attach(observer storage.MetricsObserver) {
	db.observers = append(db.observers, observer)
}

func (db *DBStorage) Detach(observer storage.MetricsObserver) {
	db.observers = slices.DeleteFunc(db.observers, func(o2 storage.MetricsObserver) bool {
		return o2 == observer
	})
}