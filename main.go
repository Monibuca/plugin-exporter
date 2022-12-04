package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/log"
	"m7s.live/plugin/exporter/v4/collector"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	defaultCollectors            = "base,cpu,memory,disk,net,media"
	defaultCollectorsPlaceholder = "[defaults]" //如果是 defaults，在 yaml 里要用双引号
)

func loadCollectors(list string, cfg config.Config) map[string]collector.Collector {
	collectors := map[string]collector.Collector{}
	enabled := expandEnabledCollectors(list)

	var (
		cCfg config.Config
		ok   bool
	)
	for _, name := range enabled {

		if cfg.Has(name) {
			collectorCfg := cfg.Get(name)
			cCfg, ok = collectorCfg.(config.Config)
			if !ok {
				log.Warnf("Exporter loadCollector %s config err, config is not map", name)
				continue
			}
		} else {
			cCfg = nil
		}
		c, err := collector.Build(name, cCfg)
		if err != nil {
			log.Warnf("Exporter loadCollector %s err: %s", name, err)
			continue
		}
		collectors[name] = c
	}

	return collectors
}

func initExporter(p *ExporterConfig) prometheus.Gatherer {
	if p.PrintCollectors {
		collectors := collector.Available()
		collectorNames := make(sort.StringSlice, 0, len(collectors))
		for _, n := range collectors {
			collectorNames = append(collectorNames, n)
		}
		collectorNames.Sort()
		log.Info("Exporter Available collectors:")
		for _, n := range collectorNames {
			log.Info(" - " + n)
		}

	}

	collectors := loadCollectors(p.Enabled, p.CollectorConfig)
	reg := prometheus.NewPedanticRegistry()

	for name, c := range collectors {
		reg.MustRegister(c)
		p.collectors[name] = c
	}

	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		reg,
	}

	return gatherers

}

func expandEnabledCollectors(enabled string) []string {
	expanded := strings.Replace(enabled, defaultCollectorsPlaceholder, defaultCollectors, -1)
	separated := strings.Split(expanded, ",")
	unique := map[string]bool{}
	for _, s := range separated {
		if s != "" {
			unique[s] = true
		}
	}
	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}
	return result
}

type errLogger struct {
}

func (l errLogger) Println(v ...interface{}) {
	v = append([]interface{}{"Exporter promhttp err: "}, v...)
	log.Error(v...)
}

type ExporterConfig struct {
	NodeAddr        string //节点位置
	Enabled         string //开启的采集器
	PrintCollectors bool
	CollectorConfig config.Config //采集器的配置
	h               http.Handler
	collectors      map[string]collector.Collector
}

var exporter = ExporterConfig{
	NodeAddr:        "zh_cn",
	Enabled:         "[defaults]",
	PrintCollectors: true,
	CollectorConfig: config.Config{},
	collectors:      make(map[string]collector.Collector),
}

type netCfg struct {
	Black string
	White string
}

func (p *ExporterConfig) OnEvent(event any) {
	switch event.(type) {
	case FirstConfig:
		cfg := config.Config(event.(FirstConfig))
		p.CollectorConfig = cfg.GetChild("collector")
		hostname, err := os.Hostname()
		if err != nil {
			log.Error("Exporter get hostname err ", err)
		}
		collector.GlobalLabel = prometheus.Labels{
			"nodeaddr": p.NodeAddr,
			"hostname": hostname,
			//"version":  SysInfo.Version,
			//"ip":       SysInfo.LocalIP,
		}
		g := initExporter(p)

		p.h = promhttp.HandlerFor(g,
			promhttp.HandlerOpts{
				ErrorLog:      errLogger{},
				ErrorHandling: promhttp.ContinueOnError,
			})
		p._onevent(event)
	default:
		p._onevent(event)
	}
}

func (p *ExporterConfig) API_metrics(w http.ResponseWriter, r *http.Request) {
	if p.h == nil {
		w.WriteHeader(500)
		w.Write([]byte("exporter is not init,wait"))
		return
	}
	p.h.ServeHTTP(w, r)
}

func (p *ExporterConfig) _onevent(event any) {
	for _, c := range p.collectors {
		c.OnEvent(event)
	}
}

var plugin = InstallPlugin(&exporter)
