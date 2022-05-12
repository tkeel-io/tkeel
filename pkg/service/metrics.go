package service

import (
	"net/http"

	go_restful "github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsService struct {
	MetricsHandler http.Handler
}

func NewMetricsService(mtrCollectors ...prometheus.Collector) *MetricsService {
	// Create a new registry.
	reg := prometheus.NewRegistry()
	reg.MustRegister(mtrCollectors...)
	metricHandler := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: false,
		},
	)

	return &MetricsService{metricHandler}
}

func (svc *MetricsService) Metrics(req *go_restful.Request, resp *go_restful.Response) {
	svc.MetricsHandler.ServeHTTP(resp, req.Request)
}
