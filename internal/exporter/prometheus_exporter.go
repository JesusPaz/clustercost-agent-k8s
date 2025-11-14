package exporter

import (
	"net/http"

	"clustercost-agent-k8s/internal/aggregator"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusExporter exposes cluster cost metrics to Prometheus.
type PrometheusExporter struct {
	registry      *prometheus.Registry
	podCost       *prometheus.GaugeVec
	namespaceCost *prometheus.GaugeVec
	nodeCost      *prometheus.GaugeVec
	clusterCost   *prometheus.GaugeVec
}

// NewPrometheusExporter initializes metrics collectors.
func NewPrometheusExporter(clusterName string) *PrometheusExporter {
	reg := prometheus.NewRegistry()

	podGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "clustercost_pod_cost_hourly",
		Help: "Estimated hourly pod cost",
	}, []string{"namespace", "pod", "node", "team", "service", "env", "client", "cluster_name", "controller_kind", "controller_name"})

	nsGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "clustercost_namespace_cost_hourly",
		Help: "Estimated hourly namespace cost",
	}, []string{"namespace", "team", "service", "env", "client", "cluster_name"})

	nodeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "clustercost_node_cost_hourly",
		Help: "Estimated hourly node cost",
	}, []string{"node", "cluster_name"})

	clusterGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "clustercost_cluster_cost_hourly",
		Help: "Estimated hourly cluster cost",
	}, []string{"cluster_name"})

	reg.MustRegister(podGauge, nsGauge, nodeGauge, clusterGauge)

	exporter := &PrometheusExporter{
		registry:      reg,
		podCost:       podGauge,
		namespaceCost: nsGauge,
		nodeCost:      nodeGauge,
		clusterCost:   clusterGauge,
	}

	clusterGauge.WithLabelValues(clusterName).Set(0)
	return exporter
}

// Handler returns the HTTP handler for /metrics.
func (p *PrometheusExporter) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

// Update copies aggregated data into Prometheus metrics.
func (p *PrometheusExporter) Update(data aggregator.AggregatedData) {
	p.podCost.Reset()
	p.namespaceCost.Reset()
	p.nodeCost.Reset()
	p.clusterCost.Reset()

	clusterName := data.Cluster.ClusterName
	p.clusterCost.WithLabelValues(clusterName).Set(data.Cluster.HourlyCost)

	for _, pod := range data.Pods {
		labels := []string{
			pod.Namespace,
			pod.Pod,
			pod.Node,
			pod.Labels["team"],
			pod.Labels["service"],
			pod.Labels["env"],
			pod.Labels["client"],
			clusterName,
			pod.ControllerKind,
			pod.ControllerName,
		}
		p.podCost.WithLabelValues(labels...).Set(pod.HourlyCost)
	}

	for _, ns := range data.Namespaces {
		labels := []string{
			ns.Namespace,
			ns.Labels["team"],
			ns.Labels["service"],
			ns.Labels["env"],
			ns.Labels["client"],
			clusterName,
		}
		p.namespaceCost.WithLabelValues(labels...).Set(ns.HourlyCost)
	}

	for _, node := range data.Nodes {
		p.nodeCost.WithLabelValues(node.Node, clusterName).Set(node.HourlyCost)
	}
}
