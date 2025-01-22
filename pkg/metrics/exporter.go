package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter struct {
	podCount          *prometheus.GaugeVec
	serviceCount      *prometheus.GaugeVec
	networkTraffic    *prometheus.CounterVec
	packetDrops       *prometheus.CounterVec
	connectionLatency *prometheus.HistogramVec
	packetSize        *prometheus.HistogramVec
	connectionStates  *prometheus.GaugeVec
	protocolTraffic   *prometheus.CounterVec
	// retransmissions   *prometheus.CounterVec
}

func NewExporter() (*Exporter, error) {
	e := &Exporter{
		podCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubenetinsight_pod_count",
				Help: "Number of pods in the cluster",
			},
			[]string{"namespace"},
		),
		serviceCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubenetinsight_service_count",
				Help: "Number of services in the cluster",
			},
			[]string{"namespace"},
		),
		networkTraffic: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubenetinsight_network_traffic_bytes",
				Help: "Total network traffic in bytes",
			},
			[]string{"source_ip", "destination_ip"},
		),
		packetDrops: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubenetinsight_packet_drops",
				Help: "Number of dropped packets",
			},
			[]string{"reason"},
		),
		connectionLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kubenetinsight_connection_latency_seconds",
				Help:    "Latency of network connections",
				Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15),
			},
			[]string{"source_ip", "destination_ip"},
		),
		packetSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kubenetinsight_packet_size_bytes",
				Help:    "Distribution of packet sizes",
				Buckets: prometheus.ExponentialBuckets(64, 2, 10), // 64B to 32KB
			},
			[]string{"source", "destination", "protocol"},
		),

		connectionStates: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubenetinsight_connection_states",
				Help: "Number of connections in each state",
			},
			[]string{"state", "source_ip", "destination_ip", "protocol", "source_port", "destination_port"},
		),

		protocolTraffic: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubenetinsight_protocol_traffic_total",
				Help: "Traffic breakdown by protocol",
			},
			[]string{"protocol", "source", "destination"},
		),
	}

	prometheus.MustRegister(e.podCount, e.serviceCount, e.networkTraffic, e.packetDrops, e.connectionLatency, e.packetSize, e.connectionStates, e.protocolTraffic)
	return e, nil
}

func (e *Exporter) UpdatePodCount(namespace string, count int) {
	e.podCount.WithLabelValues(namespace).Set(float64(count))
}

func (e *Exporter) UpdateServiceCount(namespace string, count int) {
	e.serviceCount.WithLabelValues(namespace).Set(float64(count))
}

func (e *Exporter) AddNetworkTraffic(sourceIP, destIP string, bytes float64) {
	e.networkTraffic.WithLabelValues(sourceIP, destIP).Add(bytes)
}

func (e *Exporter) IncrementPacketDrops(reason string) {
	e.packetDrops.WithLabelValues(reason).Inc()
}

func (e *Exporter) ObserveConnectionLatency(sourceIP, destIP string, latency float64) {
	e.connectionLatency.WithLabelValues(sourceIP, destIP).Observe(latency)
}

func (e *Exporter) ObservePacketSize(source, destination, protocol string, size float64) {
	e.packetSize.WithLabelValues(source, destination, protocol).Observe(size)
}

func (e *Exporter) SetConnectionState(sourceIP, destIP string, sourcePort, destPort uint16, protocol, state string, count float64) {
	e.connectionStates.WithLabelValues(
		state,
		sourceIP,
		destIP,
		protocol,
		fmt.Sprintf("%d", sourcePort),
		fmt.Sprintf("%d", destPort),
	).Set(count)
}

func (e *Exporter) AddProtocolTraffic(protocol, source, destination string, bytes float64) {
	e.protocolTraffic.WithLabelValues(protocol, source, destination).Add(bytes)
}

func (e *Exporter) StartServer(port string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+port, nil)
}
