package collector

import (
	"github.com/patrickjahns/openvpn_exporter/pkg/openvpn"
	"github.com/prometheus/client_golang/prometheus"
)

// OpenVPNCollector collects metrics from openvpn status files
type OpenVPNCollector struct {
	statusFile       string
	LastUpdated      *prometheus.Desc
	ConnectedClients *prometheus.Desc
	BytesReceived    *prometheus.Desc
	BytesSent        *prometheus.Desc
	ConnectedSince   *prometheus.Desc
}

// NewOpenVPNCollector returns a new OpenVPNCollector
func NewOpenVPNCollector(statusFile string) *OpenVPNCollector {
	return &OpenVPNCollector{
		statusFile: statusFile,

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
	labels := []string{"udp"}
	status, err := openvpn.ParseFile(c.statusFile)
	if err != nil {
		return
	}
	connectedClients := 0
	for _, client := range status.ClientList {
		connectedClients++
		ch <- prometheus.MustNewConstMetric(
			c.BytesReceived,
			prometheus.GaugeValue,
			client.BytesReceived,
			labels[0], client.CommonName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSent,
			prometheus.GaugeValue,
			client.BytesSent,
			labels[0], client.CommonName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ConnectedSince,
			prometheus.GaugeValue,
			float64(client.ConnectedSince.Unix()),
			labels[0], client.CommonName,
		)
	}
	ch <- prometheus.MustNewConstMetric(
		c.ConnectedClients,
		prometheus.GaugeValue,
		float64(connectedClients),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.LastUpdated,
		prometheus.GaugeValue,
		float64(status.UpdatedAt.Unix()),
		labels...,
	)
}
