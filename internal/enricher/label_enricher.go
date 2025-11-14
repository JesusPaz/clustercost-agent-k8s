package enricher

// LabelEnricher merges label information with a deterministic precedence order.
type LabelEnricher struct{}

var trackedKeys = []string{"team", "service", "env", "client", "cost_center"}

// NewLabelEnricher returns a no-op enricher placeholder for future expansion.
func NewLabelEnricher() *LabelEnricher {
	return &LabelEnricher{}
}

// Merge returns a filtered map using the provided sources. Later sources win.
func (e *LabelEnricher) Merge(sources ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, key := range trackedKeys {
		for i := len(sources) - 1; i >= 0; i-- {
			if sources[i] == nil {
				continue
			}
			if val, ok := sources[i][key]; ok && val != "" {
				result[key] = val
				break
			}
		}
	}
	return result
}

// Keys returns the label keys handled by the enricher.
func (e *LabelEnricher) Keys() []string {
	cp := make([]string, len(trackedKeys))
	copy(cp, trackedKeys)
	return cp
}
