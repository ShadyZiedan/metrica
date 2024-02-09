package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/shadyziedan/metrica/internal/server/config"
	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/server/middleware"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/services"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

func main() {
	cnf := config.ParseConfig()
	memStorage := storage.NewMemStorage()
	fileStorageService := services.NewFileStorageService(memStorage, services.FileStorageServiceConfig{
		FileStoragePath: cnf.FileStoragePath,
		StoreInterval:   cnf.StoreInterval,
		Restore:         cnf.Restore,
	})

	wg := &sync.WaitGroup{}
	wg.Add(2)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		defer wg.Done()
		fileStorageService.Run(ctx)
	}()

	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}
	router := handlers.NewRouter(memStorage, middleware.RequestLogger, middleware.Compress)
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
