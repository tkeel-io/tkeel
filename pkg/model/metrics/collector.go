package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics lable.
	MetricsLabelTenant = "tenant_id"
	MetricsLabelCode   = "code"
	MetricsLabelPath   = "path"
	MetricsLabelPlugin = "plugin"

	// metrics name.
	MetricsNameTkapiRequestTotal           = "tkapi_request_total"
	MetricsNameTkapiRequestDurationSeconds = "tkapi_request_duration_seconds"

	MetricsNameUserTotal = "user_num"
	MetricsNameRoleTotal = "role_num"
)

var CollectorTKApiRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: MetricsNameTkapiRequestTotal,
		Help: "tkeel api request counter.",
	},
	[]string{MetricsLabelTenant, MetricsLabelPlugin, MetricsLabelCode},
)

var CollectorTKApiRequestDurations = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    MetricsNameTkapiRequestDurationSeconds,
		Help:    "tkapi request latency distributions.",
		Buckets: []float64{0.1, 0.2, 0.4, 0.8, 1.6, 3.2, 5.0},
	},
	[]string{MetricsLabelTenant, MetricsLabelPlugin},
)

var CollectorUser = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameUserTotal,
		Help: "tkeel user num",
	},
	[]string{MetricsLabelTenant},
)

var CollectorRole = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameRoleTotal,
		Help: "tkeel role num",
	},
	[]string{MetricsLabelTenant},
)
