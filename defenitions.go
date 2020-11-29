package prostrumenter

import (
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//Prostrumenter Postrumenter
type Prostrumenter struct {
	Addr        string
	Port        string
	MetricNames []string
}

//PromMetric Defention of the PromMetricinterface
type PromMetric interface {
	getName() string
	getHelp() string
	getUpdateTime() int
}

//Used for storing names if fields in a metric
type Field struct {
	Tag  string
	Name string
	Type string
}

type counterWrapper struct {
	counter prometheus.Counter
	Type    string
	value   int
}

type guageWrapper struct {
	guage prometheus.Gauge
	Type  string
	value int
}

type listener struct {
	counters   map[reflect.Value]*counterWrapper
	gagues     map[reflect.Value]*guageWrapper
	updateTime time.Duration
}
