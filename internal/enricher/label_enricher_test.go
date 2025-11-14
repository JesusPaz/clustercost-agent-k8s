package enricher

import "testing"

func TestLabelEnricherMergePrecedence(t *testing.T) {
	en := NewLabelEnricher()

	nsLabels := map[string]string{
		"team": "platform",
		"env":  "prod",
	}
	podLabels := map[string]string{
		"team":        "payments",
		"service":     "cart",
		"cost_center": "cc-1001",
	}

	result := en.Merge(nsLabels, podLabels)

	if got := result["team"]; got != "payments" {
		t.Fatalf("expected pod label to win, got %s", got)
	}
	if got := result["env"]; got != "prod" {
		t.Fatalf("expected namespace env retained, got %s", got)
	}
	if got := result["service"]; got != "cart" {
		t.Fatalf("service label missing, got %s", got)
	}
	if got := result["cost_center"]; got != "cc-1001" {
		t.Fatalf("cost_center label missing, got %s", got)
	}
}
