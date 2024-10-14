package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/security"
	"github.com/shadyziedan/metrica/internal/server/config"
	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/server/middleware"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/services"
	"github.com/shadyziedan/metrica/internal/server/storage"
	"github.com/shadyziedan/metrica/internal/server/storage/postgres"
)

type metricsRepository interface {
	Find(ctx context.Context, name string) (*models.Metric, error)
	Create(ctx context.Context, name string, mType string) error
	FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error)
	FindAll(ctx context.Context) ([]*models.Metric, error)
	FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error)
	UpdateCounter(ctx context.Context, name string, delta int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	Attach(observer storage.MetricsObserver)
	Detach(observer storage.MetricsObserver)
}

type hasher interface {
	Hash([]byte) (string, error)
}

func main() {
	cnf := config.ParseConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}

	conn, err := pgxpool.New(ctx, cnf.DatabaseDsn)
	if err != nil {
		logger.Log.Error("Unable to create connection pool", zap.Error(err))
	} else {
		defer conn.Close()
	}
	var appStorage metricsRepository

	if conn != nil {
		appStorage, err = postgres.NewDBStorage(conn)
		if err != nil {
			logger.Log.Error("unable to initialize db", zap.Error(err))
			appStorage = storage.NewMemStorage()
		}
	} else {
		appStorage = storage.NewMemStorage()
	}

	fileStorageServiceConfig := services.FileStorageServiceConfig{
		FileStoragePath: cnf.FileStoragePath,
		StoreInterval:   cnf.StoreInterval,
		Restore:         cnf.Restore,
	}

	fileStorageService := services.NewFileStorageService(appStorage, fileStorageServiceConfig)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		fileStorageService.Run(ctx)
	}()

	var hasherimpl hasher
	if cnf.Key != "" {
		hasherimpl = security.NewDefaultHasher(cnf.Key)
	}
	router := handlers.NewRouter(
		conn,
		appStorage,
		middleware.RequestLogger,
		middleware.HashChecker(hasherimpl),
		middleware.Compress,
	)

	router.Handle(`/debug/pprof/*`, http.DefaultServeMux)

	srv := server.NewWebServer(
		cnf.Address,
		router,
	)

	go func() {
		defer wg.Done()
		err = srv.ListenAndServe(ctx)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}
