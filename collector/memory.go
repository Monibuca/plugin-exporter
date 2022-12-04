package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/mem"
	"m7s.live/engine/v4/config"
)

func init() {
	RegisterCollector("memory", newMemoryCollector)
}

type memoryCollectorBasic struct {
	Free        *prometheus.Desc
	Total       *prometheus.Desc
	Used        *prometheus.Desc
	UsedPercent *prometheus.Desc
}

func (c *memoryCollectorBasic) OnEvent(event any) {

}

func (c *memoryCollectorBasic) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Free
	ch <- c.Total
	ch <- c.Used
	ch <- c.UsedPercent
}
func (c *memoryCollectorBasic) Collect(ch chan<- prometheus.Metric) {
	path := "/"
	d, _ := mem.VirtualMemory()

	ch <- prometheus.MustNewConstMetric(
		c.Free, prometheus.GaugeValue, float64(d.Free>>20), path,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Used, prometheus.GaugeValue, float64(d.Used>>20), path,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Total, prometheus.GaugeValue, float64(d.Total>>20), path,
	)
	ch <- prometheus.MustNewConstMetric(
		c.UsedPercent, prometheus.GaugeValue, d.UsedPercent, path,
	)
}

func newMemoryCollector(cfg config.Config) (Collector, error) {
	const subsystem = "memory"

	return &memoryCollectorBasic{
		Total: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total"),
			"内存总空间(单位M)",
			[]string{"path"},
			GlobalLabel,
		),
		Free: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free"),
			"内存剩余空间(单位M)",
			[]string{"path"},
			GlobalLabel,
		),
		Used: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "used"),
			"内存已用空间(单位M)",
			[]string{"path"},
			GlobalLabel,
		),
		UsedPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "used_percent"),
			"内存已用百分比",
			[]string{"path"},
			GlobalLabel,
		),
	}, nil
}
