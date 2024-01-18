package handlers

type MetricHandler struct {
	repository MetricsRepository
}

func NewMetricHandler(repository MetricsRepository) *MetricHandler {
	return &MetricHandler{repository: repository}
}
