package collector

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// namespace defines the namespace for the exporter.
	namespace = "openvpn"
)

// GeneralCollector collects metrics, mostly runtime, about this exporter in general.
type GeneralCollector struct {
	logger    log.Logger
	version   string
	revision  string
	buildDate string
	goVersion string
	started   time.Time

	StartTime *prometheus.Desc
	BuildInfo *prometheus.Desc
}

// NewGeneralCollector returns a new GeneralCollector.
func NewGeneralCollector(logger log.Logger, version string, revision string, buildDate string, goVersion string, started time.Time) *GeneralCollector {
	return &GeneralCollector{
		logger:    logger,
		version:   version,
		revision:  revision,
		buildDate: buildDate,
		goVersion: goVersion,
		started:   started,

		StartTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "start_time"),
			"Unix timestamp of the start time of the exporter",
			nil,
			nil,
		),

		BuildInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "build_info"),
			"A metric with a constant '1' value labeled by version information",
			[]string{"version", "revision", "date", "go"},
			nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *GeneralCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.StartTime
	ch <- c.BuildInfo
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *GeneralCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log(
		"started", c.started.Unix(),
		"version", c.version,
		"revision", c.revision,
		"buildDate", c.buildDate,
		"goVersion", c.goVersion,
		"started", c.started,
	)
	ch <- prometheus.MustNewConstMetric(
		c.StartTime,
		prometheus.GaugeValue,
		float64(c.started.Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BuildInfo,
		prometheus.GaugeValue,
		1.0,
		c.version,
		c.revision,
		c.buildDate,
		c.goVersion,
	)
}
