package resource

import (
	"regexp"

	"github.com/nepec/rmqctl/internal/definitions"
)

type InlinePolicy map[string]any

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
