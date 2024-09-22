package postgres

import (
	"context"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/shadyziedan/metrica/internal/models"
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
	_, err := postgresConn.Exec(context.Background(), `
create table if not exists metrics
(
    id      serial,
    name    varchar not null,
    m_type  varchar not null,
    counter bigint,
    gauge   decimal
);
create unique index if not exists metrics_name_uindex on metrics (name);
`)
	if err != nil {
		return nil, err
	}
	return &DBStorage{conn: postgresConn}, nil
}

const findMetric = `SELECT name, m_type, gauge, counter FROM metrics WHERE name = $1;`

const createMetric = `INSERT INTO metrics (name, m_type) values ($1, $2)`

const updateCounter = `
        INSERT INTO metrics (name, m_type, counter)
        VALUES ($1, 'counter', $2)
        ON CONFLICT (name) DO UPDATE
        SET counter = coalesce(metrics.counter, 0) + $2
    `

const updateGauge = `UPDATE metrics SET gauge = $1 WHERE name = $2;`

const findOrCreateMetric = `
WITH inserted AS (
    INSERT INTO metrics (name, m_type) values ($1, $2)
    ON CONFLICT DO NOTHING
    RETURNING name, m_type, gauge, counter
)
SELECT * FROM inserted
UNION
SELECT name, m_type, gauge, counter FROM metrics WHERE name = $1;`

const findAllMetrics = `SELECT name, m_type, gauge, counter FROM metrics`
const findMetricsByName = `SELECT name, m_type, gauge, counter FROM metrics where name IN ($1)`

func (db *DBStorage) Find(ctx context.Context, name string) (*models.Metric, error) {
	row := db.conn.QueryRow(ctx, findMetric, name)
	var metric models.Metric
	err := row.Scan(&metric.Name, &metric.MType, &metric.Gauge, &metric.Counter)
	if err != nil {
		return nil, err
	}
	return &metric, err
}

func (db *DBStorage) Create(ctx context.Context, name string, mType string) error {
	_, err := db.conn.Exec(ctx, createMetric, name, mType)
	return err
}

func (db *DBStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, updateCounter, name, delta)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	updatedMetric, err := db.Find(ctx, name)
	if err != nil {
		return err
	}
	return db.notify(ctx, updatedMetric)
}

func (db *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := db.conn.Exec(ctx, updateGauge, value, name)
	if err != nil {
		return err
	}
	updatedModel, err := db.Find(ctx, name)
	if err != nil {
		return err
	}
	return db.notify(ctx, updatedModel)
}

func (db *DBStorage) FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error) {
	row := db.conn.QueryRow(ctx, findOrCreateMetric, name, mType)
	var metric models.Metric
	err := row.Scan(&metric.Name, &metric.MType, &metric.Gauge, &metric.Counter)
	if err != nil {
		return nil, err
	}
	return &metric, err
}

func (db *DBStorage) FindAll(ctx context.Context) ([]*models.Metric, error) {
	rows, err := db.conn.Query(ctx, findAllMetrics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DBStorage) FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error) {
	rows, err := db.conn.Query(ctx, findMetricsByName, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics := make([]*models.Metric, 0, len(names))
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

func (db *DBStorage) notify(ctx context.Context, model *models.Metric) error {
	for _, observer := range db.observers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := observer.Notify(model); err != nil {
				return err
			}
		}
	}
	return nil
}
