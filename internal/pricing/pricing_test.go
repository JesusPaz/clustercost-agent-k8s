package pricing

import (
	"testing"

	"clustercost-agent-k8s/internal/kube"
)

func TestCalculatorCostFor(t *testing.T) {
	calc := NewCalculator(0.1, 0.01)
	cost := calc.CostFor(500, 2*bytesInGiB)

	expected := 0.05 + 0.02
	if cost != expected {
		t.Fatalf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestPodRequestTotals(t *testing.T) {
	pod := kube.Pod{
		Containers: []kube.PodContainer{
			{CPURequestMilli: 100, MemoryRequestBytes: 256},
			{CPURequestMilli: 200, MemoryRequestBytes: 512},
		},
	}
	cpu, mem := PodRequestTotals(pod)
	if cpu != 300 || mem != 768 {
		t.Fatalf("unexpected totals cpu=%d mem=%d", cpu, mem)
	}
}

func TestPodUsageCostUsesUsageValues(t *testing.T) {
	calc := NewCalculator(1, 1)
	usage := kube.PodUsage{CPUUsageMilli: 200, MemoryUsageBytes: bytesInGiB}
	if cost := calc.PodUsageCost(usage); cost != 1.2 {
		t.Fatalf("expected 1.2, got %.2f", cost)
	}
}
