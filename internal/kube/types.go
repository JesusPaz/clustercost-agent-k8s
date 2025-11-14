package kube

import (
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// PodContainer contains a subset of container resource specifications.
type PodContainer struct {
	Name               string
	CPURequestMilli    int64
	CPULimitMilli      int64
	MemoryRequestBytes int64
	MemoryLimitBytes   int64
}

// Pod represents a simplified pod metadata payload used by the agent.
type Pod struct {
	Namespace  string
	Name       string
	UID        types.UID
	NodeName   string
	Labels     map[string]string
	OwnerKind  string
	OwnerName  string
	Containers []PodContainer
}

// Node contains relevant metadata for pricing decisions.
type Node struct {
	Name         string
	Labels       map[string]string
	InstanceType string
	CapacityCPU  int64 // milli-cores
	CapacityMem  int64 // bytes
}

// Namespace describes kubernetes namespaces with cost labels.
type Namespace struct {
	Name   string
	Labels map[string]string
}

// PodUsage details actual usage metrics collected from the metrics server.
type PodUsage struct {
	CPUUsageMilli    int64
	MemoryUsageBytes int64
}

// ClusterSnapshot is a point-in-time capture of the cluster state relevant to cost.
type ClusterSnapshot struct {
	ClusterName string
	Timestamp   time.Time
	Pods        []Pod
	Namespaces  []Namespace
	Nodes       []Node
}
