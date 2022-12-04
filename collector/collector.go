package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"m7s.live/engine/v4/config"
)

const (
	Namespace = "monibuca"
)

var GlobalLabel prometheus.Labels

type CollectorBuilder func(cfg config.Config) (Collector, error)

var (
	builders = make(map[string]CollectorBuilder)
)

func RegisterCollector(name string, builder CollectorBuilder) {
	builders[name] = builder
}

func Available() []string {
	cs := make([]string, 0, len(builders))
	for c := range builders {
		cs = append(cs, c)
	}
	return cs
}
func Build(collector string, cfg config.Config) (Collector, error) {
	builder, exists := builders[collector]
	if !exists {
		return nil, fmt.Errorf("Unknown CollectorConfig %q", collector)
	}
	return builder(cfg)
}

type Collector interface {
	prometheus.Collector
	OnEvent(event any)
}
