package postgres

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Set expectation for table creation in NewDBStorage
	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	metricName := "test_metric"
	gaugeValue := 0.5 // example gauge value

	// Prepare the mock query result
	mock.ExpectQuery(`SELECT name, m_type, gauge, counter FROM metrics WHERE name = \$1`).
		WithArgs(metricName).
		WillReturnRows(pgxmock.NewRows([]string{"name", "m_type", "gauge", "counter"}).
			AddRow(metricName, "gauge", &gaugeValue, nil))

	metric, err := storage.Find(context.Background(), metricName)
	require.NoError(t, err)

	// Assert the results
	assert.Equal(t, metricName, metric.Name)
	assert.Equal(t, "gauge", metric.MType)

	assert.Equal(t, &gaugeValue, metric.Gauge)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Set expectation for table creation in NewDBStorage
	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	metricName := "test_metric"
	metricType := "gauge"
	mock.ExpectExec(`INSERT INTO metrics \(name, m_type\) values \(\$1, \$2\)`).
		WithArgs(metricName, metricType).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = storage.Create(context.Background(), metricName, metricType)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCounter(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Set expectation for table creation in NewDBStorage
	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	metricName := "test_metric"
	delta := int64(5)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO metrics \(name, m_type, counter\)`).
		WithArgs(metricName, delta).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT name, m_type, gauge, counter FROM metrics WHERE name = \$1`).
		WithArgs(metricName).
		WillReturnRows(pgxmock.NewRows([]string{"name", "m_type", "gauge", "counter"}).
			AddRow(metricName, "counter", nil, &delta))

	err = storage.UpdateCounter(context.Background(), metricName, delta)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateGauge(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Set expectation for table creation in NewDBStorage
	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	metricName := "test_metric"
	value := 0.7

	mock.ExpectExec(`UPDATE metrics SET gauge = \$1 WHERE name = \$2`).
		WithArgs(value, metricName).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectQuery(`SELECT name, m_type, gauge, counter FROM metrics WHERE name = \$1`).
		WithArgs(metricName).
		WillReturnRows(pgxmock.NewRows([]string{"name", "m_type", "gauge", "counter"}).
			AddRow(metricName, "gauge", &value, nil))

	err = storage.UpdateGauge(context.Background(), metricName, value)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	metricName := "find_or_create_metric"
	metricType := "gauge"

	mock.ExpectQuery(`WITH inserted AS \(`).
		WithArgs(metricName, metricType).
		WillReturnRows(pgxmock.NewRows([]string{"name", "m_type", "gauge", "counter"}).
			AddRow(metricName, metricType, nil, nil)) // Return nil for Gauge and Counter

	metric, err := storage.FindOrCreate(context.Background(), metricName, metricType)
	require.NoError(t, err)

	// Assert the results
	assert.Equal(t, metricName, metric.Name)
	assert.Equal(t, metricType, metric.MType)
	assert.Nil(t, metric.Gauge)
	assert.Nil(t, metric.Counter)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	gauge1 := float64(1.0)
	counter2 := int64(20)

	mock.ExpectQuery(`SELECT name, m_type, gauge, counter FROM metrics`).
		WillReturnRows(pgxmock.NewRows([]string{"name", "m_type", "gauge", "counter"}).
			AddRow("metric1", "gauge", &gauge1, nil).
			AddRow("metric2", "counter", nil, &counter2))

	metrics, err := storage.FindAll(context.Background())
	require.NoError(t, err)

	// Assert the results
	assert.Len(t, metrics, 2)
	assert.Equal(t, "metric1", metrics[0].Name)
	assert.Equal(t, "gauge", metrics[0].MType)
	assert.Equal(t, gauge1, *metrics[0].Gauge)

	assert.Equal(t, "metric2", metrics[1].Name)
	assert.Equal(t, "counter", metrics[1].MType)
	assert.Nil(t, metrics[1].Gauge)
	assert.Equal(t, counter2, *metrics[1].Counter)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_FindAllByName(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Expect the table creation
	mock.ExpectExec(`create table if not exists metrics`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))

	storage, err := NewDBStorage(mock)
	require.NoError(t, err)

	gauge1 := float64(1.0)
	counter2 := int64(20)

	// Here, we expect the SQL to use IN ($1)
	mock.ExpectQuery(`SELECT name, m_type, gauge, counter FROM metrics where name IN \(\$1\)`).
		WithArgs(pgxmock.AnyArg()). // Allow for an array of values here
		WillReturnRows(pgxmock.NewRows([]string{"name", "m_type", "gauge", "counter"}).
			AddRow("metric1", "gauge", &gauge1, nil).
			AddRow("metric2", "counter", nil, &counter2))

	metrics, err := storage.FindAllByName(context.Background(), []string{"metric1", "metric2"})
	require.NoError(t, err)

	// Assert the results
	assert.Len(t, metrics, 2)
	assert.Equal(t, "metric1", metrics[0].Name)
	assert.Equal(t, "gauge", metrics[0].MType)
	assert.Equal(t, gauge1, *metrics[0].Gauge)

	assert.Equal(t, "metric2", metrics[1].Name)
	assert.Equal(t, "counter", metrics[1].MType)
	assert.Equal(t, counter2, *metrics[1].Counter)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
