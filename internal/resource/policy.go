package resource

import (
	"regexp"

	"github.com/nepec/rmqctl/internal/definitions"
)

// InlinePolicy is a queue's ad-hoc policy definition, declared inline on
// a QueueSpec rather than as a standalone resource.
type InlinePolicy map[string]any

// MarshalDef translates i into a definitions.Policy scoped to queueName
// on vhost, using queueName both as the policy name and as an
// exact-match regex pattern. The returned bool is false when i is empty
// (no policy to include).
func (i InlinePolicy) MarshalDef(vhost, queueName string) (definitions.Policy, bool) {
	if len(i) == 0 {
		return definitions.Policy{}, false
	}
	return definitions.Policy{
		Vhost:      vhost,
		Name:       queueName,
		Pattern:    "^" + regexp.QuoteMeta(queueName) + "$",
		ApplyTo:    "queues",
		Definition: map[string]any(i),
		Priority:   0,
	}, true
}
