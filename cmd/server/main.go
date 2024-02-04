package main

import (
	"context"
	"github.com/shadyziedan/metrica/internal/server/config"
	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/server/middleware"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/services"
	"github.com/shadyziedan/metrica/internal/server/storage"
	"os"
	"os/signal"
)

func main() {
	cnf := config.ParseConfig()
	memStorage := storage.NewMemStorage()
	fileStorageService := services.NewFileStorageService(memStorage, services.FileStorageServiceConfig{
		FileStoragePath: cnf.FileStoragePath,
		StoreInterval:   cnf.StoreInterval,
		Restore:         cnf.Restore,
	})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go fileStorageService.Run(ctx)

	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}
	router := handlers.NewRouter(memStorage, middleware.RequestLogger, middleware.Compress)
	srv := server.NewWebServer(
		cnf.Address,
		router,
	)
	err = srv.ListenAndServe(ctx)
	if err != nil {
		panic(err)
	}
}
