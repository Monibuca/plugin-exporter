package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"m7s.live/engine/v4/config"
	"time"
)

func init() {
	RegisterCollector("cpu", newCPUCollector)
}

type cpuCollectorBasic struct {
	UserTime   *prometheus.Desc
	Usage      *prometheus.Desc
	SystemTime *prometheus.Desc
	IdleTime   *prometheus.Desc
}

var cpuConfig = struct {
	PerCpu bool
}{PerCpu: false}

func (c *cpuCollectorBasic) OnEvent(event any) {

}

func (c *cpuCollectorBasic) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.UserTime
	ch <- c.Usage
	ch <- c.SystemTime
	ch <- c.IdleTime
}
func (c *cpuCollectorBasic) Collect(ch chan<- prometheus.Metric) {

	if usages, err := cpu.Percent(time.Second, cpuConfig.PerCpu); err == nil {

		prefix := "cpu"

		for i, u := range usages {
			labelVal := ""
			if cpuConfig.PerCpu == false {
				labelVal = fmt.Sprintf("%s-%s", prefix, "total")
			} else {
				labelVal = fmt.Sprintf("%s-%d", prefix, i)
			}
			ch <- prometheus.MustNewConstMetric(
				c.Usage, prometheus.GaugeValue, u, labelVal,
			)
		}
	}

	if cpuStats, err := cpu.Times(cpuConfig.PerCpu); err == nil {
		for _, cpuStat := range cpuStats {
			ch <- prometheus.MustNewConstMetric(
				c.UserTime, prometheus.GaugeValue, cpuStat.User, cpuStat.CPU,
			)
			ch <- prometheus.MustNewConstMetric(
				c.SystemTime, prometheus.GaugeValue, cpuStat.System, cpuStat.CPU,
			)
			ch <- prometheus.MustNewConstMetric(
				c.IdleTime, prometheus.GaugeValue, cpuStat.Idle, cpuStat.CPU,
			)
		}
	}
}

func newCPUCollector(cfg config.Config) (Collector, error) {
	const subsystem = "cpu"
	if cfg != nil {
		cfg.Unmarshal(&cpuConfig)
	}
	jiffiesDesc := "(单位：jiffiesDesc 1jiffies=0.01秒)"
	return &cpuCollectorBasic{
		UserTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "user_time"),
			"用户态的CPU时间"+jiffiesDesc,
			[]string{"core"},
			GlobalLabel,
		),
		Usage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usage"),
			"CPU 利用率",
			[]string{"core"},
			GlobalLabel,
		),
		SystemTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "system_time"),
			"系统态的CPU时间"+jiffiesDesc,
			[]string{"core"},
			GlobalLabel,
		),
		IdleTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "idle_time"),
			"空闲态的CPU时间"+jiffiesDesc,
			[]string{"core"},
			GlobalLabel,
		),
	}, nil
}
