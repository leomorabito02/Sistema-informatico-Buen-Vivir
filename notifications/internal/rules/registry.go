package rules

import "sync"

// Registry holds the mapping between event_type and domain notification rules.
type Registry struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

// NewRegistry initializes a thread-safe rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]Rule),
	}
}

// Register adds a rule to the registry.
func (r *Registry) Register(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[rule.EventType()] = rule
}

// Get retrieves a rule by event_type.
func (r *Registry) Get(eventType string) (Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, found := r.rules[eventType]
	return rule, found
}

// EventTypes returns a slice of all registered event types.
func (r *Registry) EventTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.rules))
	for t := range r.rules {
		types = append(types, t)
	}
	return types
}

