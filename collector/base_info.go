package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	"m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	RegisterCollector("base", newBaseCollector)
}

func version2float(version string) float64 {
	version = strings.ReplaceAll(version, "v", "")
	ver := strings.Split(version, ".")
	verF := 0.0
	for i, v := range ver {
		num, err := strconv.Atoi(v)
		if err != nil {
			log.Warn("wrong format version: " + version)
			return 0
		}
		verF += float64(num) / math.Pow10(2*i)
	}
	return verF
}

type baseCollectorBasic struct {
	RunTime           *prometheus.Desc
	ProcessMemory     *prometheus.Desc
	ProcessCpuTime    *prometheus.Desc
	ProcessCpuPercent *prometheus.Desc
	pid               int
}

func (c *baseCollectorBasic) OnEvent(event any) {

}

func (c *baseCollectorBasic) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.RunTime
	ch <- c.ProcessCpuTime
	ch <- c.ProcessCpuPercent
	ch <- c.ProcessMemory
}
func (c *baseCollectorBasic) Collect(ch chan<- prometheus.Metric) {
	start := engine.SysInfo.StartTime

	ch <- prometheus.MustNewConstMetric(
		c.RunTime, prometheus.CounterValue, time.Now().Sub(start).Seconds(), start.Format("2006-01-02 15:04:05"),
	)

	m7sProcess, err := process.NewProcess(int32(c.pid))
	if err != nil {
		return
	}
	mem, _ := m7sProcess.MemoryInfo()
	if mem != nil {
		ch <- prometheus.MustNewConstMetric(
			c.ProcessMemory, prometheus.GaugeValue, float64(mem.RSS>>20), "physical",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessMemory, prometheus.GaugeValue, float64(mem.VMS>>20), "virtual",
		)
	}
	cpuTime, _ := m7sProcess.Times()
	if cpuTime != nil {
		ch <- prometheus.MustNewConstMetric(
			c.ProcessCpuTime, prometheus.GaugeValue, cpuTime.User, "user",
		)
	}

	percent, _ := m7sProcess.CPUPercent()
	ch <- prometheus.MustNewConstMetric(
		c.ProcessCpuPercent, prometheus.GaugeValue, percent,
	)

}

func onetime_baseInfo(subsystem string) {
	label := make(prometheus.Labels)
	for k, v := range GlobalLabel {
		label[k] = v
	}
	label["ip"] = engine.SysInfo.LocalIP
	label["version"] = engine.SysInfo.Version

	m7sVer := version2float(engine.SysInfo.Version)

	baseGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   Namespace,
		Subsystem:   subsystem,
		Name:        "info",
		Help:        "Monibuca 基础信息",
		ConstLabels: label,
	})
	prometheus.MustRegister(baseGauge)
	baseGauge.Set(m7sVer)

	delete(label, "ip")
	delete(label, "version")
	platform, family, kernelVersion, _ := host.PlatformInformation()

	label["platform"] = platform
	label["family"] = family
	label["kernel_version"] = kernelVersion
	label["pid"] = strconv.Itoa(os.Getpid())
	systemGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   Namespace,
		Subsystem:   subsystem,
		Name:        "os",
		Help:        "系统基础信息",
		ConstLabels: label,
	})

	prometheus.MustRegister(systemGauge)
	systemGauge.Set(m7sVer)
}

func newBaseCollector(cfg config.Config) (Collector, error) {
	const subsystem = "base"
	onetime_baseInfo(subsystem)
	return &baseCollectorBasic{
		RunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "running_time"),
			"Monibuca 运行时间",
			[]string{"start_time"},
			GlobalLabel,
		),
		ProcessMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "process_memory"),
			"Monibuca 内存占用(单位M)",
			[]string{"memory_type"},
			GlobalLabel,
		),
		ProcessCpuPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "process_cpu_percent"),
			"Monibuca Cpu占用百分比",
			nil,
			GlobalLabel,
		),
		ProcessCpuTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "process_cpu_time"),
			"Monibuca Cpu时间",
			[]string{"time_type"},
			GlobalLabel,
		),
		pid: os.Getpid(),
	}, nil
}
