package promcase

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Various type of metric types and their short codes used in Udp message
const (
	MetricTypeCounter   = "c"
	MetricTypeGauge     = "g"
	MetricTypeHistogram = "h"
)

// Various type of action types and their short codes used in Udp message
const (
	ActionTypeInc     = "i"
	ActionTypeDec     = "d"
	ActionTypeSet     = "s"
	ActionTypeObserve = "o"
)

// Registry represents per metric type registered prometheus collectors
type Registry struct {
	TypeCounter   map[string]*prometheus.CounterVec
	TypeGauge     map[string]*prometheus.GaugeVec
	TypeHistogram map[string]*prometheus.HistogramVec
}

// Register is the only Registry in use now, equivalent to Prometheus's default registry
var Register Registry

func init() {
	Register = Registry{
		make(map[string]*prometheus.CounterVec),
		make(map[string]*prometheus.GaugeVec),
		make(map[string]*prometheus.HistogramVec)}
}
