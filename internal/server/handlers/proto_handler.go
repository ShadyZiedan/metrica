package handlers

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/shadyziedan/metrica/proto" // Import the generated protobuf package
)

type GRPCMetricServiceServer struct {
	pb.UnimplementedMetricsServiceServer
	repository metricsRepository
}

func NewGRPCMetricServiceServer(repository metricsRepository) *GRPCMetricServiceServer {
	return &GRPCMetricServiceServer{repository: repository}
}

// UpdateBatchMetrics handles a batch update of metrics using protobuf.
func (s *GRPCMetricServiceServer) UpdateBatchMetrics(ctx context.Context, req *pb.UpdateBatchRequest) (*pb.UpdateBatchResponse, error) {
	var response pb.UpdateBatchResponse
	for _, item := range req.Metrics {
		var mType string
		switch item.MType {
		case pb.Metric_GAUGE:
			mType = "gauge"
		case pb.Metric_COUNTER:
			mType = "counter"
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Invalid metric type: %s", item.MType)
		}
		metric, err := s.repository.FindOrCreate(ctx, item.Id, mType)
		if err != nil {
			return nil, fmt.Errorf("error finding or creating metric: %w", err)
		}

		switch item.MType {
		case pb.Metric_COUNTER:
			if item.Delta == nil {
				return nil, fmt.Errorf("error updating null counter value metric")
			}
			if err = s.repository.UpdateCounter(ctx, metric.Name, *item.Delta); err != nil {
				return nil, fmt.Errorf("error updating counter metric: %w", err)
			}
		case pb.Metric_GAUGE:
			if item.Value == nil {
				return nil, fmt.Errorf("error updating null gauge value metric")
			}
			if err = s.repository.UpdateGauge(ctx, metric.Name, *item.Value); err != nil {
				return nil, fmt.Errorf("error updating gauge metric: %w", err)
			}
		}

		// Fetch updated metric and prepare response
		updatedMetric, err := s.repository.Find(ctx, metric.Name)
		if err != nil {
			return nil, fmt.Errorf("error fetching updated metric: %w", err)
		}
		responseMetric := &pb.Metric{
			Id:    updatedMetric.Name,
			MType: item.MType,
		}
		switch updatedMetric.MType {
		case "counter":
			responseMetric.Delta = updatedMetric.Counter
		case "gauge":
			responseMetric.Value = updatedMetric.Gauge
		}
		response.Metrics = append(response.Metrics, responseMetric)
	}

	return &response, nil
}
