package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/patrickjahns/openvpn_exporter/pkg/openvpn"
	"github.com/prometheus/client_golang/prometheus"
)

// OpenVPNCollector collects metrics from openvpn status files
type OpenVPNCollector struct {
	logger           log.Logger
	name             string
	statusFile       string
	LastUpdated      *prometheus.Desc
	ConnectedClients *prometheus.Desc
	BytesReceived    *prometheus.Desc
	BytesSent        *prometheus.Desc
	ConnectedSince   *prometheus.Desc
}

// NewOpenVPNCollector returns a new OpenVPNCollector
func NewOpenVPNCollector(logger log.Logger, name string, statusFile string) *OpenVPNCollector {
	return &OpenVPNCollector{
		logger:     logger,
		statusFile: statusFile,
		name:       name,

		LastUpdated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_updated"),
			"Unix timestamp when the last time the status was updated",
			[]string{"server"},
			nil,
		),
		ConnectedClients: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections"),
			"Amount of currently connected clients",
			[]string{"server"},
			nil,
		),
		BytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bytes_received"),
			"Amount of data received via the connection",
			[]string{"server", "common_name"},
			nil,
		),
		BytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bytes_sent"),
			"Amount of data sent via the connection",
			[]string{"server", "common_name"},
			nil,
		),
		ConnectedSince: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connected_since"),
			"Unixtimestamp when the connection was established",
			[]string{"server", "common_name"},
			nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *OpenVPNCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.LastUpdated
	ch <- c.ConnectedClients
	ch <- c.BytesSent
	ch <- c.BytesReceived
	ch <- c.ConnectedSince
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *OpenVPNCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log(
		"statusFile", c.statusFile,
		"name", c.name,
	)
	status, err := openvpn.ParseFile(c.statusFile)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "error parsing statusfile",
			"err", err,
		)
		return
	}

	connectedClients := 0
	for _, client := range status.ClientList {
		connectedClients++
		level.Debug(c.logger).Log(
			"commonName", client.CommonName,
			"connectedSince", client.ConnectedSince.Unix(),
			"bytesReceived", client.BytesReceived,
			"bytesSent", client.BytesSent,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesReceived,
			prometheus.GaugeValue,
			client.BytesReceived,
			c.name, client.CommonName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSent,
			prometheus.GaugeValue,
			client.BytesSent,
			c.name, client.CommonName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ConnectedSince,
			prometheus.GaugeValue,
			float64(client.ConnectedSince.Unix()),
			c.name, client.CommonName,
		)
	}
	level.Debug(c.logger).Log(
		"updatedAt", status.UpdatedAt,
		"connectedClients", connectedClients,
		"maxBcastMcastQueueLen", status.GlobalStats.MaxBcastMcastQueueLen,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ConnectedClients,
		prometheus.GaugeValue,
		float64(connectedClients),
		c.name,
	)
	ch <- prometheus.MustNewConstMetric(
		c.LastUpdated,
		prometheus.GaugeValue,
		float64(status.UpdatedAt.Unix()),
		c.name,
	)
}
