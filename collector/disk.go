package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/disk"
	"m7s.live/engine/v4/config"
)

func init() {
	RegisterCollector("disk", newDiskCollector)
}

type diskCollectorBasic struct {
	Free        *prometheus.Desc
	Total       *prometheus.Desc
	Used        *prometheus.Desc
	UsedPercent *prometheus.Desc
}

func (c *diskCollectorBasic) OnEvent(event any) {

}

func (c *diskCollectorBasic) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Free
	ch <- c.Total
	ch <- c.Used
	ch <- c.UsedPercent
}
func (c *diskCollectorBasic) Collect(ch chan<- prometheus.Metric) {
	path := "/"
	d, _ := disk.Usage(path)

	ch <- prometheus.MustNewConstMetric(
		c.Free, prometheus.GaugeValue, float64(d.Free>>30), path,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Used, prometheus.GaugeValue, float64(d.Used>>30), path,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Total, prometheus.GaugeValue, float64(d.Total>>30), path,
	)
	ch <- prometheus.MustNewConstMetric(
		c.UsedPercent, prometheus.GaugeValue, d.UsedPercent, path,
	)
}

func newDiskCollector(cfg config.Config) (Collector, error) {
	const subsystem = "disk"

	return &diskCollectorBasic{
		Total: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total"),
			"Monibuca 所在分区总空间(单位G)",
			[]string{"path"},
			GlobalLabel,
		),
		Free: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "free"),
			"Monibuca 所在分区剩余空间(单位G)",
			[]string{"path"},
			GlobalLabel,
		),
		Used: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "used"),
			"Monibuca 所在分区已用空间(单位G)",
			[]string{"path"},
			GlobalLabel,
		),
		UsedPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "used_percent"),
			"Monibuca 所在分区已用百分比",
			[]string{"path"},
			GlobalLabel,
		),
	}, nil
}
