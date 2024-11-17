package main

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/security"
	pb "github.com/shadyziedan/metrica/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/signal"
	"syscall"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"github.com/shadyziedan/metrica/internal/agent/logger"
	"github.com/shadyziedan/metrica/internal/agent/services"
)

var (
	BuildVersion string
	BuildDate    string
	BuildCommit  string
)

func main() {
	showBuildInfo()
	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}

	cnf := config.ParseConfig()

	var options []services.MetricsSenderOption
	if cnf.CryptoKey != "" {
		encryptor, err := security.NewDefaultEncryptorFromFile(cnf.CryptoKey)
		if err != nil {
			logger.Log.Error("encryption key is not loaded", zap.Error(err))
		} else {
			options = append(options, services.WithEncryptor(encryptor))
		}
	}
	if cnf.Key != "" {
		hasher := security.NewDefaultHasher(cnf.Key)
		options = append(options, services.WithHasher(hasher))
	}
	var httpClient *resty.Client
	var grpcClient pb.MetricsServiceClient
	if cnf.Address != "" {
		httpClient = resty.New().SetBaseURL(cnf.Address)
	}
	if cnf.GrpcAddress != "" {
		conn, err := grpc.NewClient(cnf.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Log.Fatal("grpc connection error", zap.Error(err))
		}
		grpcClient = pb.NewMetricsServiceClient(conn)
	}
	newAgent := agent.NewAgent(
		cnf,
		services.NewMetricsCollector(),
		services.NewMetricsSender(httpClient, grpcClient, options...),
	)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	newAgent.Run(ctx)
}

func showBuildInfo() {
	if BuildVersion != "" {
		fmt.Println("Build version: ", BuildVersion)
	} else {
		fmt.Println("Build version: N/A")
	}
	if BuildDate != "" {
		fmt.Println("Build date: ", BuildDate)
	} else {
		fmt.Println("Build date: N/A")
	}
	if BuildCommit != "" {
		fmt.Println("Build commit: ", BuildCommit)
	} else {
		fmt.Println("Build commit: N/A")
	}
}
