package collector

import (
	"context"
	"fmt"
	"log/slog"

	"clustercost-agent-k8s/internal/kube"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetricsCollector retrieves usage metrics from the metrics.k8s.io API.
type MetricsCollector struct {
	client *kube.Client
	logger *slog.Logger
}

// NewMetricsCollector returns a configured collector.
func NewMetricsCollector(client *kube.Client, logger *slog.Logger) *MetricsCollector {
	return &MetricsCollector{client: client, logger: logger}
}

// CollectPodMetrics returns usage metrics keyed by namespace/pod name.
func (c *MetricsCollector) CollectPodMetrics(ctx context.Context) (map[string]kube.PodUsage, error) {
	result := make(map[string]kube.PodUsage)

	metricsList, err := c.client.Metrics.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pod metrics: %w", err)
	}

	for _, m := range metricsList.Items {
		var cpuMilli int64
		var memBytes int64
		for _, container := range m.Containers {
			cpuMilli += container.Usage.Cpu().MilliValue()
			memBytes += container.Usage.Memory().Value()
		}
		key := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
		result[key] = kube.PodUsage{CPUUsageMilli: cpuMilli, MemoryUsageBytes: memBytes}
	}

	return result, nil
}
