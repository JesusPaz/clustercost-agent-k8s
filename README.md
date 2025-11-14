# ClusterCost Kubernetes Agent

`clustercost-agent-k8s` is the open-source Kubernetes cost visibility agent that powers the ClusterCost platform. It runs once per cluster, collects allocation signals (pods, namespaces, nodes, services), correlates usage metrics, and exposes real-time hourly cost estimates via Prometheus metrics and a JSON API for the local dashboard.

## Highlights

- **Cluster-wide visibility** – CPU/memory requests and usage mapped to pods, namespaces, nodes, controllers, and cluster totals.
- **Label-aware allocation** – understands `team`, `service`, `env`, `client`, and `cost_center` labels across namespaces and workloads.
- **Prometheus + JSON API** – scrape `/metrics` or call `/api/cost/*` for summaries, per-namespace, or per-pod views.
- **Config-driven pricing** – supply CPU and memory hourly prices per region/provider via ConfigMap, env vars, or flags.
- **Secure-by-default** – read-only RBAC, no outbound calls, minimal resource footprint.

## Architecture Overview

```
Kubernetes API + Metrics API --> Collectors --> Enricher --> Aggregator --> Prometheus / HTTP API
```

The agent periodically performs a scrape (default 30s):

1. **Collectors** gather pods, namespaces, nodes, and usage metrics.
2. **Label Enricher** merges namespace and workload labels for consistent allocation keys.
3. **Aggregator** derives hourly cost per entity and stores the latest dataset.
4. **Exporters** project the data via `/metrics`, `/api/health`, `/api/cost/summary`, `/api/cost/namespaces`, and `/api/cost/pods`.

## Quickstart with Helm

```bash
helm repo add clustercost https://charts.clustercost.io
helm repo update
helm install clustercost-agent clustercost/clustercost-agent-k8s \
  --namespace clustercost --create-namespace \
  --set pricing.cpuHourPrice=0.046 \
  --set pricing.memoryGibHourPrice=0.005 \
  --set clusterName=my-prod-cluster
```

For local development, use the bundled chart:

```bash
helm install clustercost-agent deployment/helm \
  --namespace clustercost --create-namespace
```

## Configuration

| Source | Notes |
| --- | --- |
| **Flags** | `--listen-addr`, `--scrape-interval`, `--cluster-name`, `--cpu-price`, `--memory-price`, etc. |
| **Environment** | `CLUSTERCOST_LOG_LEVEL`, `CLUSTERCOST_LISTEN_ADDR`, `CLUSTERCOST_SCRAPE_INTERVAL`, `CLUSTERCOST_CPU_HOUR_PRICE`, `CLUSTERCOST_MEMORY_GIB_HOUR_PRICE`, `CLUSTERCOST_PROVIDER`, `CLUSTERCOST_REGION`, `CLUSTERCOST_CLUSTER_NAME`. |
| **ConfigMap YAML** | Mounted file referenced via `CLUSTERCOST_CONFIG_FILE` or `--config`. Keys mirror the struct names in `internal/config`. |

Example ConfigMap snippet:

```yaml
pricing:
  provider: aws
  region: us-east-1
  cpuHourPrice: 0.046
  memoryGibHourPrice: 0.005
clusterName: prod-us-east-1
scrapeIntervalSeconds: 45
```

## Prometheus Metrics

Key metric families exported:

- `clustercost_pod_cost_hourly{namespace,pod,node,team,service,env,client,cluster_name}`
- `clustercost_namespace_cost_hourly{namespace,team,service,env,client,cluster_name}`
- `clustercost_node_cost_hourly{node,cluster_name}`
- `clustercost_cluster_cost_hourly{cluster_name}`

Additional internal gauges capture usage and owner metadata, enabling dashboards, alerting, or recording rules.

## HTTP JSON API

- `GET /api/health` – readiness + cluster summary.
- `GET /api/cost/summary` – cluster totals plus aggregated label spend.
- `GET /api/cost/namespaces` – namespaces with hourly cost and request totals.
- `GET /api/cost/pods` – pods enriched with node placement and controller fields.

## Security & RBAC

The chart provisions a dedicated ServiceAccount with:

- `get`, `list`, `watch` on pods, namespaces, deployments, services, and nodes.
- `get`, `list` on metrics.k8s.io resources.
- No write operations, no exec, and no permissions outside the cluster.

Only cluster-local APIs are contacted; there are **no outbound network calls**. TLS termination is left to the cluster ingress stack if required.

## Pricing Model

The MVP relies on static pricing numbers supplied via configuration. The default is tuned for generic AWS on-demand nodes. Override CPU and memory prices per cluster to match your blend (e.g., spot, savings plans). Future versions can plug into the ClusterCost Cloud for live price feeds.

## Roadmap

- Deeper CSV/Parquet export for offline analysis.
- Built-in rightsizing signals.
- Automatic AWS pricing sync and EKS/Fargate differentiation.
- Optional WASM policy hooks for custom allocation logic.
- Extended APIs for services/deployments and label taxonomies.

## Local Development

```bash
# Build binary
make build # or go build ./cmd/agent

# Run locally against current kube context
CLUSTERCOST_KUBECONFIG=$HOME/.kube/config \
CLUSTERCOST_CPU_HOUR_PRICE=0.04 \
CLUSTERCOST_MEMORY_GIB_HOUR_PRICE=0.004 \
go run ./cmd/agent
```

See `examples/kind` and `examples/sample-app` for quick local playgrounds.

## Container Image

Build a minimal distroless-based image using the multi-stage `Dockerfile`:

```bash
docker build -t ghcr.io/you/clustercost-agent-k8s:dev .
docker run --rm -p 8080:8080 \
  -e CLUSTERCOST_CLUSTER_NAME=dev-cluster \
  ghcr.io/you/clustercost-agent-k8s:dev
```
