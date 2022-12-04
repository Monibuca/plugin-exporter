package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"m7s.live/engine/v4"
	"m7s.live/engine/v4/config"
)

func init() {
	RegisterCollector("media", newMediaCollector)
}

type mediaCollectorBasic struct {
	OnlineClients *prometheus.Desc
	OnlineStreams *prometheus.Desc
	TotalClients  *prometheus.Desc
	TotalStreams  *prometheus.Desc

	StreamBps         *prometheus.Desc
	StreamSubscribers *prometheus.Desc

	mediaTotal  int64
	clientTotal int64
}

func (c *mediaCollectorBasic) OnEvent(event any) {
	switch event.(type) {
	case engine.SEpublish:
		c.mediaTotal += 1
	case engine.ISubscriber:
		c.clientTotal += 1
	}
}

func (c *mediaCollectorBasic) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.OnlineClients
	ch <- c.OnlineStreams
	ch <- c.TotalClients
	ch <- c.TotalStreams
	ch <- c.StreamBps
	ch <- c.StreamSubscribers
}
func (c *mediaCollectorBasic) Collect(ch chan<- prometheus.Metric) {
	onlineClientCnt := 0
	engine.Streams.Range(func(name string, ss *engine.Stream) {
		ch <- prometheus.MustNewConstMetric(
			c.StreamBps, prometheus.GaugeValue, float64(ss.Summary().BPS), name,
		)
		clientCnt := ss.Summary().Subscribers
		onlineClientCnt += clientCnt
		ch <- prometheus.MustNewConstMetric(
			c.StreamSubscribers, prometheus.GaugeValue, float64(clientCnt), name,
		)
	})

	ch <- prometheus.MustNewConstMetric(
		c.TotalStreams, prometheus.CounterValue, float64(c.mediaTotal),
	)
	ch <- prometheus.MustNewConstMetric(
		c.TotalClients, prometheus.CounterValue, float64(c.clientTotal),
	)
	ch <- prometheus.MustNewConstMetric(
		c.OnlineStreams, prometheus.GaugeValue, float64(engine.Streams.Len()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.OnlineClients, prometheus.GaugeValue, float64(onlineClientCnt),
	)
}

func newMediaCollector(cfg config.Config) (Collector, error) {
	const subsystem = "media"

	return &mediaCollectorBasic{
		OnlineStreams: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "online_stream_count"),
			"在线媒体流数目",
			nil,
			GlobalLabel,
		),
		OnlineClients: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "online_client_count"),
			"在线客户端数目",
			nil,
			GlobalLabel,
		),
		TotalClients: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_client_sum"),
			"历史客户端总数",
			nil,
			GlobalLabel,
		),
		TotalStreams: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_stream_sum"),
			"历史媒体流总数",
			nil,
			GlobalLabel,
		),
		StreamBps: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "stream_bps"),
			"媒体流 bps",
			[]string{"name"},
			GlobalLabel,
		),
		StreamSubscribers: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "stream_client_count"),
			"媒体流在线客户端数目",
			[]string{"name"},
			GlobalLabel,
		),
	}, nil
}
