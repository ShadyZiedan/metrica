package server

import (
	"context"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"

	pb "github.com/shadyziedan/metrica/proto"
)

type GRPCServer struct {
	address           string
	metricsRepository metricsRepository
}

func NewGRPCServer(address string, metricsRepository metricsRepository) *GRPCServer {
	return &GRPCServer{address: address, metricsRepository: metricsRepository}
}

type metricsRepository interface {
	Find(ctx context.Context, name string) (*models.Metric, error)
	Create(ctx context.Context, name string, mType string) error
	FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error)
	FindAll(ctx context.Context) ([]*models.Metric, error)
	UpdateCounter(ctx context.Context, name string, delta int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error)
}

func (s *GRPCServer) Run(ctx context.Context) {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		logger.Log.Fatal("failed to listen to port", zap.String("address", s.address), zap.Error(err))
	}

	server := grpc.NewServer()
	metricService := handlers.NewGRPCMetricServiceServer(s.metricsRepository)

	pb.RegisterMetricsServiceServer(server, metricService)
	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Log.Fatal("failed to serve", zap.String("address", s.address), zap.Error(err))
		}
	}()
	<-ctx.Done()
	server.GracefulStop()
}
