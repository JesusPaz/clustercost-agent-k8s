package pricing

import "clustercost-agent-k8s/internal/kube"

const bytesInGiB = 1024 * 1024 * 1024

// Calculator converts resource consumption to cost using static pricing inputs.
type Calculator struct {
	CPUCoreHourPrice   float64
	MemoryGiBHourPrice float64
}

// NewCalculator returns a calculator using the configuration values.
func NewCalculator(cpuPrice, memPrice float64) *Calculator {
	return &Calculator{CPUCoreHourPrice: cpuPrice, MemoryGiBHourPrice: memPrice}
}

// CostFor requests calculates the hourly cost for the provided resource requests.
func (c *Calculator) CostFor(cpuMilli int64, memoryBytes int64) float64 {
	cores := float64(cpuMilli) / 1000.0
	memoryGiB := float64(memoryBytes) / bytesInGiB
	return cores*c.CPUCoreHourPrice + memoryGiB*c.MemoryGiBHourPrice
}

// PodRequestTotals sums all container requests for a pod.
func PodRequestTotals(pod kube.Pod) (cpuMilli int64, memoryBytes int64) {
	for _, container := range pod.Containers {
		cpuMilli += container.CPURequestMilli
		memoryBytes += container.MemoryRequestBytes
	}
	return
}

// PodUsageCost calculates cost for actual usage metrics.
func (c *Calculator) PodUsageCost(usage kube.PodUsage) float64 {
	return c.CostFor(usage.CPUUsageMilli, usage.MemoryUsageBytes)
}
