package collector

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/net"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/log"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterCollector("net", NewNetworkCollector)
}

var (
	nicNameToUnderscore = regexp.MustCompile("[^a-zA-Z0-9]")
	netConfig           = struct {
		NicWhitelist string
		NicBlacklist string
	}{".*", ""}
)

type netInfo struct {
	net.IOCountersStat
	ReceiveSpeed float64
	SentSpeed    float64
}

// A NetworkCollector is a Prometheus Collector for Perflib Network Interface metrics
type NetworkCollector struct {
	BytesReceivedTotal *prometheus.Desc
	BytesSentTotal     *prometheus.Desc
	BytesTotal         *prometheus.Desc

	BytesReceiveSpeed *prometheus.Desc
	BytesSentSpeed    *prometheus.Desc

	PacketsReceivedTotal *prometheus.Desc
	PacketsSentTotal     *prometheus.Desc
	PacketsTotal         *prometheus.Desc

	ErrIn    *prometheus.Desc
	ErrOut   *prometheus.Desc
	ErrTotal *prometheus.Desc

	nicWhitelistPattern *regexp.Regexp
	nicBlacklistPattern *regexp.Regexp

	lastNetWork map[string]*netInfo
	lastTime    time.Time
}

func (c *NetworkCollector) OnEvent(event any) {

}

func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.BytesReceivedTotal
	ch <- c.BytesSentTotal
	ch <- c.BytesTotal

	ch <- c.BytesReceiveSpeed
	ch <- c.BytesSentSpeed

	ch <- c.PacketsTotal
	ch <- c.PacketsReceivedTotal
	ch <- c.PacketsSentTotal

	ch <- c.ErrIn
	ch <- c.ErrOut
	ch <- c.ErrTotal

}

func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	nv, _ := net.IOCounters(true)
	now := time.Now()
	for _, nic := range nv {
		if c.nicBlacklistPattern.MatchString(nic.Name) ||
			!c.nicWhitelistPattern.MatchString(nic.Name) {
			continue
		}

		if ni, exist := c.lastNetWork[nic.Name]; exist {
			delta := now.Sub(c.lastTime).Seconds()

			ni.ReceiveSpeed = float64(nic.BytesRecv-c.lastNetWork[nic.Name].BytesRecv) / delta
			ni.SentSpeed = float64(nic.BytesSent-c.lastNetWork[nic.Name].BytesSent) / delta
			ni.IOCountersStat = nic
		} else {
			c.lastNetWork[nic.Name] = &netInfo{IOCountersStat: nic}
		}
		log.Infof("aaaa now: %d  last %d", now.Unix(), c.lastTime.Unix())
		c.lastTime = now

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedTotal,
			prometheus.CounterValue,
			float64(nic.BytesRecv),
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSentTotal,
			prometheus.CounterValue,
			float64(nic.BytesSent),
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesTotal,
			prometheus.CounterValue,
			float64(nic.BytesSent+nic.BytesRecv),
			nic.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceiveSpeed,
			prometheus.GaugeValue,
			c.lastNetWork[nic.Name].ReceiveSpeed,
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSentSpeed,
			prometheus.GaugeValue,
			c.lastNetWork[nic.Name].SentSpeed,
			nic.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceivedTotal,
			prometheus.CounterValue,
			float64(nic.PacketsRecv),
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsSentTotal,
			prometheus.CounterValue,
			float64(nic.PacketsSent),
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PacketsTotal,
			prometheus.CounterValue,
			float64(nic.PacketsSent+nic.PacketsRecv),
			nic.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ErrIn,
			prometheus.CounterValue,
			float64(nic.Errin),
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ErrOut,
			prometheus.CounterValue,
			float64(nic.Errout),
			nic.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ErrTotal,
			prometheus.CounterValue,
			float64(nic.Errin+nic.Errout),
			nic.Name,
		)
	}

}

func NewNetworkCollector(cfg config.Config) (Collector, error) {
	const subsystem = "net"
	if cfg != nil {
		cfg.Unmarshal(&netConfig)
	}
	return &NetworkCollector{
		BytesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_received_total"),
			"网络接收字节总数 byte",
			[]string{"nic"},
			nil,
		),
		BytesSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sent_total"),
			"网络发送字节总数 byte",
			[]string{"nic"},
			nil,
		),
		BytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_total"),
			"网络收发字节总数",
			[]string{"nic"},
			nil,
		),

		BytesReceiveSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_received_speed"),
			"网络接收速度 byte/s",
			[]string{"nic"},
			nil,
		),
		BytesSentSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sent_speed"),
			"网络发送速度 byte/s",
			[]string{"nic"},
			nil,
		),

		PacketsReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_total"),
			"网络接收数据包总数",
			[]string{"nic"},
			nil,
		),
		PacketsSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_sent_total"),
			"网络发送数据包总数",
			[]string{"nic"},
			nil,
		),
		PacketsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_total"),
			"网络收发数据包总数",
			[]string{"nic"},
			nil,
		),

		ErrIn: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_errors_total"),
			"网络接收错误总数",
			[]string{"nic"},
			nil,
		),
		ErrOut: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_sent_errors_total"),
			"网络发送错误总数",
			[]string{"nic"},
			nil,
		),
		ErrTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_errors_total"),
			"网络收发错误总数",
			[]string{"nic"},
			nil,
		),

		nicWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", netConfig.NicWhitelist)),
		nicBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", netConfig.NicBlacklist)),
		lastNetWork:         make(map[string]*netInfo),
	}, nil
}
