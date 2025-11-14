package aggregator

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"clustercost-agent-k8s/internal/enricher"
	"clustercost-agent-k8s/internal/kube"
	"clustercost-agent-k8s/internal/pricing"
)

// PodCost represents cost and usage for a single pod.
type PodCost struct {
	Namespace          string            `json:"namespace"`
	Pod                string            `json:"pod"`
	Node               string            `json:"node"`
	HourlyCost         float64           `json:"hourlyCost"`
	CPURequestMilli    int64             `json:"cpuRequestMilli"`
	MemoryRequestBytes int64             `json:"memoryRequestBytes"`
	CPUUsageMilli      int64             `json:"cpuUsageMilli"`
	MemoryUsageBytes   int64             `json:"memoryUsageBytes"`
	Labels             map[string]string `json:"labels"`
	ControllerKind     string            `json:"controllerKind"`
	ControllerName     string            `json:"controllerName"`
}

// NamespaceCost aggregates cost at the namespace scope.
type NamespaceCost struct {
	Namespace          string            `json:"namespace"`
	HourlyCost         float64           `json:"hourlyCost"`
	CPURequestMilli    int64             `json:"cpuRequestMilli"`
	MemoryRequestBytes int64             `json:"memoryRequestBytes"`
	Labels             map[string]string `json:"labels"`
}

// NodeCost aggregates pod cost on a node.
type NodeCost struct {
	Node               string            `json:"node"`
	HourlyCost         float64           `json:"hourlyCost"`
	CPURequestMilli    int64             `json:"cpuRequestMilli"`
	MemoryRequestBytes int64             `json:"memoryRequestBytes"`
	Labels             map[string]string `json:"labels"`
}

// LabelCost stores aggregation per logical label.
type LabelCost struct {
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	HourlyCost float64 `json:"hourlyCost"`
}

// ClusterCostSummary captures the overall cluster picture.
type ClusterCostSummary struct {
	ClusterName          string  `json:"clusterName"`
	HourlyCost           float64 `json:"hourlyCost"`
	TotalCPURequestMilli int64   `json:"totalCpuRequestMilli"`
	TotalMemoryRequestB  int64   `json:"totalMemoryRequestBytes"`
	PodCount             int     `json:"podCount"`
	NodeCount            int     `json:"nodeCount"`
	GeneratedAtUnix      int64   `json:"generatedAtUnix"`
}

// AggregatedData holds the latest aggregation results.
type AggregatedData struct {
	GeneratedAt time.Time          `json:"generatedAt"`
	Cluster     ClusterCostSummary `json:"cluster"`
	Namespaces  []NamespaceCost    `json:"namespaces"`
	Pods        []PodCost          `json:"pods"`
	Nodes       []NodeCost         `json:"nodes"`
	Labels      []LabelCost        `json:"labels"`
}

// CostAggregator coordinates aggregations and stores the latest snapshot.
type CostAggregator struct {
	mu       sync.RWMutex
	data     AggregatedData
	calc     *pricing.Calculator
	enricher *enricher.LabelEnricher
	logger   *slog.Logger
}

// NewCostAggregator returns a configured aggregator.
func NewCostAggregator(calc *pricing.Calculator, enricher *enricher.LabelEnricher, logger *slog.Logger) *CostAggregator {
	return &CostAggregator{calc: calc, enricher: enricher, logger: logger}
}

// Aggregate computes fresh cost data and stores it internally.
func (a *CostAggregator) Aggregate(snapshot kube.ClusterSnapshot, usage map[string]kube.PodUsage) AggregatedData {
	now := snapshot.Timestamp
	podCosts := make([]PodCost, 0, len(snapshot.Pods))
	nsMap := make(map[string]*NamespaceCost)
	nodeMap := make(map[string]*NodeCost)
	labelMap := make(map[string]map[string]float64)

	nsLabels := map[string]map[string]string{}
	for _, ns := range snapshot.Namespaces {
		nsCopy := make(map[string]string, len(ns.Labels))
		for k, v := range ns.Labels {
			nsCopy[k] = v
		}
		nsLabels[ns.Name] = nsCopy
	}

	nodeLabels := map[string]map[string]string{}
	for _, node := range snapshot.Nodes {
		nodeCopy := make(map[string]string, len(node.Labels))
		for k, v := range node.Labels {
			nodeCopy[k] = v
		}
		nodeLabels[node.Name] = nodeCopy
	}

	var clusterCost float64
	var totalCPU, totalMem int64

	for _, pod := range snapshot.Pods {
		cpuReq, memReq := pricing.PodRequestTotals(pod)
		totalCPU += cpuReq
		totalMem += memReq

		key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		podUsage := usage[key]
		cpuUsage := podUsage.CPUUsageMilli
		memUsage := podUsage.MemoryUsageBytes
		if cpuUsage == 0 {
			cpuUsage = cpuReq
		}
		if memUsage == 0 {
			memUsage = memReq
		}
		cost := a.calc.CostFor(cpuUsage, memUsage)

		labels := a.enricher.Merge(nsLabels[pod.Namespace], pod.Labels)
		podCost := PodCost{
			Namespace:          pod.Namespace,
			Pod:                pod.Name,
			Node:               pod.NodeName,
			HourlyCost:         cost,
			CPURequestMilli:    cpuReq,
			MemoryRequestBytes: memReq,
			CPUUsageMilli:      cpuUsage,
			MemoryUsageBytes:   memUsage,
			Labels:             labels,
			ControllerKind:     pod.OwnerKind,
			ControllerName:     pod.OwnerName,
		}
		podCosts = append(podCosts, podCost)

		nsTotals := ensureNamespace(nsMap, pod.Namespace, labels)
		nsTotals.HourlyCost += cost
		nsTotals.CPURequestMilli += cpuReq
		nsTotals.MemoryRequestBytes += memReq

		nodeTotals := ensureNode(nodeMap, pod.NodeName, nodeLabels[pod.NodeName])
		nodeTotals.HourlyCost += cost
		nodeTotals.CPURequestMilli += cpuReq
		nodeTotals.MemoryRequestBytes += memReq

		for k, v := range labels {
			if v == "" {
				continue
			}
			if _, ok := labelMap[k]; !ok {
				labelMap[k] = make(map[string]float64)
			}
			labelMap[k][v] += cost
		}

		clusterCost += cost
	}

	namespaces := make([]NamespaceCost, 0, len(nsMap))
	for _, ns := range nsMap {
		namespaces = append(namespaces, *ns)
	}
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Namespace < namespaces[j].Namespace })

	nodes := make([]NodeCost, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Node < nodes[j].Node })

	labels := make([]LabelCost, 0)
	for key, values := range labelMap {
		for value, cst := range values {
			labels = append(labels, LabelCost{Key: key, Value: value, HourlyCost: cst})
		}
	}
	sort.Slice(labels, func(i, j int) bool {
		if labels[i].Key == labels[j].Key {
			return labels[i].Value < labels[j].Value
		}
		return labels[i].Key < labels[j].Key
	})

	data := AggregatedData{
		GeneratedAt: now,
		Cluster: ClusterCostSummary{
			ClusterName:          snapshot.ClusterName,
			HourlyCost:           clusterCost,
			TotalCPURequestMilli: totalCPU,
			TotalMemoryRequestB:  totalMem,
			PodCount:             len(snapshot.Pods),
			NodeCount:            len(snapshot.Nodes),
			GeneratedAtUnix:      now.Unix(),
		},
		Namespaces: namespaces,
		Pods:       podCosts,
		Nodes:      nodes,
		Labels:     labels,
	}

	a.mu.Lock()
	a.data = data
	a.mu.Unlock()

	a.logger.Debug("aggregated costs", slog.Float64("cluster_cost_hourly", clusterCost))
	return data
}

// Data returns the latest aggregated data.
func (a *CostAggregator) Data() AggregatedData {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data
}

func ensureNamespace(cache map[string]*NamespaceCost, name string, labels map[string]string) *NamespaceCost {
	if ns, ok := cache[name]; ok {
		return ns
	}
	ns := &NamespaceCost{Namespace: name, Labels: labels}
	cache[name] = ns
	return ns
}

func ensureNode(cache map[string]*NodeCost, name string, labels map[string]string) *NodeCost {
	if node, ok := cache[name]; ok {
		return node
	}
	node := &NodeCost{Node: name, Labels: labels}
	cache[name] = node
	return node
}
