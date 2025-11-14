# Run ClusterCost Agent on kind

1. Create the kind cluster:
   ```bash
   kind create cluster --config examples/kind/cluster.yaml
   ```
2. Install metrics-server (required for usage signals):
   ```bash
   kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
   ```
3. Deploy the sample workload:
   ```bash
   kubectl apply -f examples/sample-app/deployment.yaml
   ```
4. Install the agent chart from the repo root:
   ```bash
   helm install clustercost-agent deployment/helm \
     --namespace clustercost --create-namespace \
     --set clusterName=kind-dev \
     --set pricing.cpuHourPrice=0.025 \
     --set pricing.memoryGibHourPrice=0.003
   ```
5. Port-forward the service if you want to inspect the JSON API:
   ```bash
   kubectl -n clustercost port-forward svc/clustercost-agent-clustercost-agent-k8s 8080:8080
   curl localhost:8080/api/cost/summary | jq
   ```
