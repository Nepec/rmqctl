package resource

import "github.com/nepec/rmqctl/internal/definitions"

// QueueBinding declares a binding from an exchange to the queue it's
// nested under in a QueueSpec.
type QueueBinding struct {
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routingKey"`
}

// MarshalDef translates b into a definitions.Binding from b.Exchange to
// the queue queueName on vhost. The returned bool reports whether the
// binding should be included in the output; for QueueBinding it is
// always true.
func (b QueueBinding) MarshalDef(vhost, queueName string) (definitions.Binding, bool) {
	return definitions.Binding{
		Source:          b.Exchange,
		Vhost:           vhost,
		Destination:     queueName,
		DestinationType: "queue",
		RoutingKey:      b.RoutingKey,
	}, true
}
