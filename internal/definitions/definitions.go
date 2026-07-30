// Package definitions defines DTOs to translate a rmqctl spec
// into a RabbitMQ definition.
//
// A RabbitMQ definition is a JSON map with the same format as a
// definitions file imported/exported directly from a RabbitMQ node.
package definitions

// Definitions is the full set of RabbitMQ resource definitions to be
// applied to a vhost.
type Definitions struct {
	Queues    []Queue    `json:"queues,omitempty"`
	Bindings  []Binding  `json:"bindings,omitempty"`
	Policies  []Policy   `json:"policies,omitempty"`
	Exchanges []Exchange `json:"exchanges,omitempty"`
}

// Queue is a RabbitMQ queue definition.
type Queue struct {
	Name       string         `json:"name"`
	Vhost      string         `json:"vhost"`
	Durable    bool           `json:"durable"`
	AutoDelete bool           `json:"auto_delete,omitempty"`
	Arguments  map[string]any `json:"arguments"`
}

// Binding is a RabbitMQ binding definition, from Source to Destination.
type Binding struct {
	Source          string         `json:"source"`
	Vhost           string         `json:"vhost"`
	Destination     string         `json:"destination,omitempty"`
	DestinationType string         `json:"destination_type"`
	RoutingKey      string         `json:"routing_key,omitempty"`
	Arguments       map[string]any `json:"arguments,omitempty"`
}

// Policy is a RabbitMQ policy definition.
type Policy struct {
	Vhost      string         `json:"vhost"`
	Name       string         `json:"name"`
	Pattern    string         `json:"pattern"`
	ApplyTo    string         `json:"apply_to"`
	Definition map[string]any `json:"definition"`
	Priority   int            `json:"priority"`
}

// Exchange is a RabbitMQ exchange definition.
type Exchange struct {
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Arguments map[string]any `json:"arguments,omitempty"`
}
