package service

import (
	go_restful "github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsService struct {
}

func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

func (h *MetricsService) Metrics(req *go_restful.Request, resp *go_restful.Response) {
	promhttp.Handler().ServeHTTP(resp, req.Request)
}
