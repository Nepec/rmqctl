package api

import (
	"fmt"
	"maps"
)

type Policy struct {
	Vhost, Pattern, ApplyTo, Name string
	Priority                      int
	Definition                    map[string]any
}
type Policies map[string]Policy

// NewStandardQueuePolicy defines a new RabbitMQ policy scoped to queues only.
// The queue's name is used both for the policy name and the exact-matched regex pattern (ex "^name$")
// The definition is provided as input.
func NewStandardQueuePolicy(vhost, name string, definition map[string]any) Policy {
	return Policy{
		Vhost:      vhost,
		Pattern:    fmt.Sprintf("^%s$", name),
		ApplyTo:    "queues",
		Name:       name,
		Priority:   0,
		Definition: definition,
	}
}

// MergePolicyDefinitions merges two RabbitMQ policy definition maps into a new map.
// Keys present in only one map are always included.
// When the same key exists in both, the force flag determines precedence:
// force=false - current values are preserved
// force = true - update valued overwrite current
// Neither maps are modified.
func MergePolicyDefinitions(current, update map[string]any, force bool) map[string]any {
	// Since maps.Copy overwirtes on copy, if base=update and overwirite=current
	// curren values will be overwritten onto the updated ones, making sure the result
	// preserves the original values
	base, overwrite := update, current
	if force {
		// if base=current and overwrite=update, then the update will effectively overwrite current values
		base, overwrite = current, update
	}

	merged := make(map[string]any, len(base)+len(overwrite))
	maps.Copy(merged, base)
	maps.Copy(merged, overwrite)

	return merged
}
