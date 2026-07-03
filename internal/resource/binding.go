package resource

import "github.com/nepec/rmqctl/internal/definitions"

type QueueBinding struct {
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routingKey"`
}

func (b QueueBinding) MarshalDef(vhost, queueName string) (definitions.Binding, bool) {
	return definitions.Binding{
		Source:          b.Exchange,
		Vhost:           vhost,
		Destination:     queueName,
		DestinationType: "queue",
		RoutingKey:      b.RoutingKey,
	}, true
}
