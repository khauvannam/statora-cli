package switcher

import "statora-cli/internal/resolver"

// Action describes what will change for a single item.
type Action struct {
	Kind    string // "php", "composer", "ext-enable", "ext-disable"
	Name    string // version string or extension name
	Current string // current value (empty if not set)
	Next    string // target value
}

// Plan holds all the changes that `statora switch` will apply.
type Plan struct {
	Resolution resolver.Resolution
	Actions    []Action
}

// HasChanges returns true if there is at least one action.
func (p *Plan) HasChanges() bool {
	return len(p.Actions) > 0
}
