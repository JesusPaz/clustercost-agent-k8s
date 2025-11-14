package aggregator

import (
	"testing"
	"time"

	"clustercost-agent-k8s/internal/enricher"
	"clustercost-agent-k8s/internal/kube"
	"clustercost-agent-k8s/internal/pricing"

	"log/slog"
	"os"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestCostAggregatorAggregate(t *testing.T) {
	calc := pricing.NewCalculator(1, 0.5) // simple prices for assertions
	agg := NewCostAggregator(calc, enricher.NewLabelEnricher(), newTestLogger())

	snapshot := kube.ClusterSnapshot{
		ClusterName: "dev",
		Timestamp:   time.Unix(123, 0),
		Namespaces: []kube.Namespace{
			{Name: "payments", Labels: map[string]string{"team": "finops", "env": "prod"}},
		},
		Nodes: []kube.Node{
			{Name: "node-a", Labels: map[string]string{"topology.kubernetes.io/zone": "us-east-1a"}},
		},
		Pods: []kube.Pod{
			{
				Namespace: "payments",
				Name:      "cart-abc",
				NodeName:  "node-a",
				Labels:    map[string]string{"service": "cart"},
				Containers: []kube.PodContainer{
					{CPURequestMilli: 200, MemoryRequestBytes: 512 * 1024 * 1024},
				},
			},
		},
	}

	usage := map[string]kube.PodUsage{
		"payments/cart-abc": {CPUUsageMilli: 300, MemoryUsageBytes: 512 * 1024 * 1024},
	}

	data := agg.Aggregate(snapshot, usage)

	if len(data.Pods) != 1 {
		t.Fatalf("expected 1 pod result, got %d", len(data.Pods))
	}

	pod := data.Pods[0]
	expectedCost := calc.CostFor(300, 512*1024*1024)
	if pod.HourlyCost != expectedCost {
		t.Fatalf("expected pod cost %.2f, got %.2f", expectedCost, pod.HourlyCost)
	}

	if data.Cluster.HourlyCost != expectedCost {
		t.Fatalf("cluster hourly cost mismatch: %.2f vs %.2f", data.Cluster.HourlyCost, expectedCost)
	}

	if len(data.Labels) == 0 {
		t.Fatalf("expected label aggregations")
	}

	foundTeam := false
	for _, label := range data.Labels {
		if label.Key == "team" && label.Value == "finops" {
			foundTeam = true
			if label.HourlyCost != expectedCost {
				t.Fatalf("expected team cost %.2f got %.2f", expectedCost, label.HourlyCost)
			}
		}
	}
	if !foundTeam {
		t.Fatalf("team label aggregation missing")
	}
}
