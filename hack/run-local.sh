#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"
export CLUSTERCOST_KUBECONFIG="${CLUSTERCOST_KUBECONFIG:-$HOME/.kube/config}"
export CLUSTERCOST_CPU_HOUR_PRICE="${CLUSTERCOST_CPU_HOUR_PRICE:-0.04}"
export CLUSTERCOST_MEMORY_GIB_HOUR_PRICE="${CLUSTERCOST_MEMORY_GIB_HOUR_PRICE:-0.004}"
export CLUSTERCOST_CLUSTER_NAME="${CLUSTERCOST_CLUSTER_NAME:-dev-cluster}"
export CLUSTERCOST_LOG_LEVEL="${CLUSTERCOST_LOG_LEVEL:-debug}"

validate_access() {
	if ! command -v kubectl >/dev/null 2>&1; then
		echo "kubectl not found in PATH; please install kubectl to validate cluster access." >&2
		return 1
	fi

	if [ ! -f "$CLUSTERCOST_KUBECONFIG" ]; then
		echo "Kubeconfig '$CLUSTERCOST_KUBECONFIG' does not exist." >&2
		return 1
	fi

	echo "Validating cluster access with kubectl (5s timeout, default namespace)..."
	if ! kubectl --kubeconfig "$CLUSTERCOST_KUBECONFIG" get pods -n default --chunk-size=1 --request-timeout=5s -o name >/dev/null 2>&1; then
		echo "WARNING: kubectl could not list pods with the provided kubeconfig." >&2
		echo "Check that your AWS credentials are active (e.g., 'aws sso login' or aws-vault), that" >&2
		echo "the kubeconfig context points to an accessible cluster, and that your IAM/RBAC role" >&2
		echo "has list/watch permissions. The agent will likely fail until this is resolved." >&2
	else
		echo "kubectl validation succeeded."
	fi
}

validate_access

go run "$ROOT_DIR/cmd/agent"
