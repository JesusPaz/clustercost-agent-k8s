package exporter

import (
	"encoding/json"
	"net/http"
	"time"

	"clustercost-agent-k8s/internal/aggregator"
)

// HTTPAPI exposes REST handlers used by local dashboards.
type HTTPAPI struct {
	aggregator *aggregator.CostAggregator
}

// NewHTTPAPI builds a HTTPAPI helper.
func NewHTTPAPI(agg *aggregator.CostAggregator) *HTTPAPI {
	return &HTTPAPI{aggregator: agg}
}

// Register wires endpoints into the provided mux.
func (h *HTTPAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", h.health)
	mux.HandleFunc("/api/cost/summary", h.summary)
	mux.HandleFunc("/api/cost/namespaces", h.namespaces)
	mux.HandleFunc("/api/cost/pods", h.pods)
}

func (h *HTTPAPI) health(w http.ResponseWriter, r *http.Request) {
	data := h.aggregator.Data()
	respondJSON(w, HealthStatus{
		Status:      "ok",
		GeneratedAt: data.GeneratedAt,
		Cluster:     data.Cluster,
	})
}

func (h *HTTPAPI) summary(w http.ResponseWriter, r *http.Request) {
	data := h.aggregator.Data()
	payload := map[string]any{
		"cluster": data.Cluster,
		"labels":  data.Labels,
	}
	respondJSON(w, payload)
}

func (h *HTTPAPI) namespaces(w http.ResponseWriter, r *http.Request) {
	data := h.aggregator.Data()
	respondJSON(w, data.Namespaces)
}

func (h *HTTPAPI) pods(w http.ResponseWriter, r *http.Request) {
	data := h.aggregator.Data()
	respondJSON(w, data.Pods)
}

func respondJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}

// HealthStatus describes readiness values.
type HealthStatus struct {
	Status      string                        `json:"status"`
	GeneratedAt time.Time                     `json:"generatedAt"`
	Cluster     aggregator.ClusterCostSummary `json:"cluster"`
}
