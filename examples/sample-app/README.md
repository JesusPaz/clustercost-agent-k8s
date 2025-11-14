# Sample App with Cost Labels

This deployment emulates a simple service with team/environment labels that the agent consumes. Apply it with:

```bash
kubectl apply -f examples/sample-app/deployment.yaml
```

Once pods are running, the agent will expose per-pod and per-namespace costs under `/metrics` and `/api/cost/*`.
