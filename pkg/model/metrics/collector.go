package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics label.
	MetricsLabelTenant = "tenant_id"
	MetricsLabelCode   = "code"
	MetricsLabelPath   = "path"
	MetricsLabelPlugin = "plugin"
	MetricsLabelKey    = "key"

	// metrics name.
	MetricsNameTkapiRequestTotal           = "tkapi_request_total"
	MetricsNameTkapiRequestDurationSeconds = "tkapi_request_duration_seconds"

	MetricsNameUserNum       = "user_num"
	MetricsNameRoleNum       = "role_num"
	MetricsNameTKeelProfiles = "tkeel_profiles"
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
		Name: MetricsNameUserNum,
		Help: "tkeel user num",
	},
	[]string{MetricsLabelTenant},
)

var CollectorRole = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameRoleNum,
		Help: "tkeel role num",
	},
	[]string{MetricsLabelTenant},
)

var CollectorTKeelProfiles = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameTKeelProfiles,
		Help: "tkeel profiles",
	},
	[]string{MetricsLabelTenant, MetricsLabelKey},
)
