package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"

	"github.com/patrickjahns/openvpn_exporter/pkg/openvpn"
)

// OpenVPNCollector collects metrics from openvpn status files
type OpenVPNCollector struct {
	logger                log.Logger
	collectClientMetrics  bool
	allowDuplicateCn      bool
	OpenVPNServer         []OpenVPNServer
	LastUpdated           *prometheus.Desc
	ConnectedClients      *prometheus.Desc
	BytesReceived         *prometheus.Desc
	BytesSent             *prometheus.Desc
	ConnectedSince        *prometheus.Desc
	MaxBcastMcastQueueLen *prometheus.Desc
	ServerInfo            *prometheus.Desc
	CollectionError       *prometheus.CounterVec
}

// OpenVPNServer contains information of which servers will be scraped
type OpenVPNServer struct {
	Name       string
	StatusFile string
	ParseError float64
}

// NewOpenVPNCollector returns a new OpenVPNCollector
func NewOpenVPNCollector(logger log.Logger, openVPNServer []OpenVPNServer, collectClientMetrics bool, allowDuplicateCn bool) *OpenVPNCollector {
	return &OpenVPNCollector{
		logger:               logger,
		OpenVPNServer:        openVPNServer,
		collectClientMetrics: collectClientMetrics,
		allowDuplicateCn:     allowDuplicateCn,

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
		MaxBcastMcastQueueLen: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_bcast_mcast_queue_len"),
			"MaxBcastMcastQueueLen of the server",
			[]string{"server"},
			nil,
		),
		BytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bytes_received"),
			"Amount of data received via the connection",
			[]string{"server", "common_name", "unique_id"},
			nil,
		),
		BytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bytes_sent"),
			"Amount of data sent via the connection",
			[]string{"server", "common_name", "unique_id"},
			nil,
		),
		ConnectedSince: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connected_since"),
			"Unixtimestamp when the connection was established",
			[]string{"server", "common_name", "unique_id"},
			nil,
		),
		ServerInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "server_info"),
			"A metric with a constant '1' value labeled by version information",
			[]string{"server", "version", "arch"},
			nil,
		),
		CollectionError: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, "", "collection_error"),
				Help: "Error occurred during collection",
			},
			[]string{"server"},
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *OpenVPNCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.LastUpdated
	ch <- c.ConnectedClients
	ch <- c.MaxBcastMcastQueueLen
	ch <- c.ServerInfo
	if c.collectClientMetrics {
		ch <- c.BytesSent
		ch <- c.BytesReceived
		ch <- c.ConnectedSince
	}
	c.CollectionError.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *OpenVPNCollector) Collect(ch chan<- prometheus.Metric) {
	for _, ovpn := range c.OpenVPNServer {
		c.collect(ovpn, ch)
	}
}

func (c *OpenVPNCollector) collect(ovpn OpenVPNServer, ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log(
		"statusFile", ovpn.StatusFile,
		"name", ovpn.Name,
	)
	status, err := openvpn.ParseFile(ovpn.StatusFile)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "error parsing statusfile",
			"err", err,
		)
		c.CollectionError.WithLabelValues(ovpn.Name).Add(1)
		c.CollectionError.Collect(ch)
		return
	}

	connectedClients := 0
	hasFacedZeroPeerId := false
	var clientCommonNames []string
	for _, client := range status.ClientList {
		connectedClients++
		level.Debug(c.logger).Log(
			"commonName", client.CommonName,
			"connectedSince", client.ConnectedSince.Unix(),
			"bytesReceived", client.BytesReceived,
			"bytesSent", client.BytesSent,
		)
		if c.collectClientMetrics {
			if client.CommonName == "UNDEF" {
				continue
			}
			if contains(clientCommonNames, client.CommonName) {
				if !c.allowDuplicateCn {
					level.Warn(c.logger).Log(
						"msg", "duplicate client common name in statusfile - duplicate metric dropped (use --allow-duplicate-cn flag)",
						"commonName", client.CommonName,
					)
					continue
				}
				if c.allowDuplicateCn && client.PeerID == -1 {
					level.Warn(c.logger).Log(
						"msg", "allow-duplicate-cn flag with a version 1 statusfile - duplicate metric dropped (use version 2 or 3)",
						"commonName", client.CommonName,
					)
					continue
				}
			}
			clientCommonNames = append(clientCommonNames, client.CommonName)
			uniqueId := client.PeerID
			if uniqueId == 0 { // In TCP mode, PeerID is always 0
				if hasFacedZeroPeerId { // Use ClientID only if a 0 PeerID is matched twice (maybe it's the first item and we are in UDP mode)
					uniqueId = -client.ClientID - 1 // ClientID starts at 0; But it may be duplicated with another 0 PeerID
				}
				hasFacedZeroPeerId = true
			}
			ch <- prometheus.MustNewConstMetric(
				c.BytesReceived,
				prometheus.CounterValue,
				client.BytesReceived,
				ovpn.Name, client.CommonName, strconv.FormatInt(uniqueId, 10),
			)
			ch <- prometheus.MustNewConstMetric(
				c.BytesSent,
				prometheus.CounterValue,
				client.BytesSent,
				ovpn.Name, client.CommonName, strconv.FormatInt(uniqueId, 10),
			)
			ch <- prometheus.MustNewConstMetric(
				c.ConnectedSince,
				prometheus.GaugeValue,
				float64(client.ConnectedSince.Unix()),
				ovpn.Name, client.CommonName, strconv.FormatInt(uniqueId, 10),
			)
		}
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
		ovpn.Name,
	)
	ch <- prometheus.MustNewConstMetric(
		c.LastUpdated,
		prometheus.GaugeValue,
		float64(status.UpdatedAt.Unix()),
		ovpn.Name,
	)
	ch <- prometheus.MustNewConstMetric(
		c.MaxBcastMcastQueueLen,
		prometheus.GaugeValue,
		float64(status.GlobalStats.MaxBcastMcastQueueLen),
		ovpn.Name,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServerInfo,
		prometheus.GaugeValue,
		1.0,
		ovpn.Name,
		status.ServerInfo.Version,
		status.ServerInfo.Arch,
	)
}

func contains(list []string, item string) bool {
	for _, e := range list {
		if e == item {
			return true
		}
	}
	return false
}
