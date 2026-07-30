// Package api provides a thin, testable client for administering RabbitMQ
// via its Management HTTP API.
//
// It handles authentication, request building, and response parsing for
// operations on queues, exchanges, vhosts, and policies. RabbitHoleClient
// is the concrete implementation of the RabbitClient interface, which the
// rest of the codebase depends on so a fake can be substituted in tests.
package api

// ClientFactory builds a RabbitClient, deferring connection setup to the
// caller, often a cli command.
type ClientFactory func() (RabbitClient, error)

// RabbitClient is the set of RabbitMQ Management API operations rmqctl
// depends on.
type RabbitClient interface {
	// Host returns the RabbitMQ management API host this client talks to.
	Host() string
	// Port returns the RabbitMQ management API port this client talks to.
	Port() int
	QueueStore
	PolicyStore
	BindingStore
	VhostStore
}

// QueueStore lists and declares queues on a vhost.
type QueueStore interface {
	// ListQueuesIn returns every queue defined on vhost.
	ListQueuesIn(vhost string) ([]Queue, error)
	// DeclareQueue creates or updates the queue named name on vhost.
	DeclareQueue(vhost, name string, opts QueueDeclareOpts) error
}

// PolicyStore lists, reads, and writes policies on a vhost.
type PolicyStore interface {
	// ListPoliciesIn returns every policy defined on vhost, keyed by name.
	ListPoliciesIn(vhost string) (Policies, error)
	// GetPolicy returns the policy named name on vhost.
	GetPolicy(vhost, name string) (*Policy, error)
	// PutPolicy creates or overwrites the policy named name on vhost.
	PutPolicy(vhost, name string, policy Policy) error
}

// BindingStore declares bindings between RabbitMQ resources.
type BindingStore interface {
	// DeclareBinding creates the binding described by opts on vhost.
	DeclareBinding(vhost string, opts Binding) error
}

// VhostStore lists virtual hosts.
type VhostStore interface {
	// ListVhosts returns every virtual host known to the node.
	ListVhosts() ([]Vhost, error)
}
