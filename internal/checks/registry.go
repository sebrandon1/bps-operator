package checks

import "sync"

var (
	mu       sync.RWMutex
	registry []CheckInfo
)

// Register adds a check to the global registry.
func Register(info CheckInfo) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, info)
}

// All returns all registered checks.
func All() []CheckInfo {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]CheckInfo, len(registry))
	copy(out, registry)
	return out
}

// Filtered returns checks matching the given names. If names is empty, returns all.
func Filtered(names []string) []CheckInfo {
	if len(names) == 0 {
		return All()
	}
	allowed := make(map[string]bool, len(names))
	for _, n := range names {
		allowed[n] = true
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []CheckInfo
	for _, c := range registry {
		if allowed[c.Name] {
			out = append(out, c)
		}
	}
	return out
}
