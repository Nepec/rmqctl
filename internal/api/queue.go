package api

import (
	"strings"
)

// Queue is a snapshot of a RabbitMQ queue's state, as returned by the
// Management API.
type Queue struct {
	Name, Type, Vhost, PolicyName string
	Messages                      int
	Active                        bool
}

// QueueDeclareOpts holds the settings used to declare or update a queue.
type QueueDeclareOpts struct {
	Durable    bool
	AutoDelete bool
	Arguments  map[string]any
}

// QueueFilterOpts narrows a queue listing. String fields match by
// substring/equality when non-empty; the pointer fields are three-valued
// (nil means "no constraint") so callers can distinguish "unset" from
// "must be false".
type QueueFilterOpts struct {
	Contains, Type            string
	Active, Empty, WithPolicy *bool
}

// Matches determines if a queue satisfies all criteria in the filter.
// Nil or zero-valued fields are ignored and treated as no constraint.
func (f *QueueFilterOpts) Matches(q Queue) bool {
	if f.Contains != "" && !strings.Contains(q.Name, f.Contains) {
		return false
	}

	if f.Type != "" && q.Type != f.Type {
		return false
	}

	if f.Active != nil && q.Active != *f.Active {
		return false
	}

	if f.Empty != nil {
		isQEmpty := (q.Messages == 0)
		if isQEmpty != *f.Empty {
			return false
		}
	}

	if f.WithPolicy != nil && q.PolicyName == "" {
		return false
	}

	return true
}

// FilterQueues returns only the queues that match all criteria in QueueFilterOpts.
// If no options are passed, it returns the queues unchanged.
func FilterQueues(queues []Queue, opts *QueueFilterOpts) []Queue {
	if opts == nil {
		return queues
	}

	filtered := make([]Queue, 0, len(queues))
	for _, q := range queues {
		if opts.Matches(q) {
			filtered = append(filtered, q)
		}
	}

	return filtered
}
