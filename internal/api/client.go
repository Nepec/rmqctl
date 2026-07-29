// Package api provides a thin, testable client for administering RabbitMQ
// via its Management HTTP API.
//
// It handles authentication, request building, and response parsing for
// operations on queues, exchanges, vhosts, and policies. The Client type
// is the main entry point.
package api

type ClientFactory func() (RabbitClient, error)

type RabbitClient interface {
	Host() string
	Port() int
	QueueStore
	PolicyStore
	BindingStore
	VhostStore
}

type QueueStore interface {
	ListQueuesIn(vhost string) ([]Queue, error)
	DeclareQueue(vhost, name string, opts QueueDeclareOpts) error
}

type PolicyStore interface {
	ListPoliciesIn(vhost string) (Policies, error)
	GetPolicy(vhost, name string) (*Policy, error)
	PutPolicy(vhost, name string, policy Policy) error
}

type BindingStore interface {
	DeclareBinding(vhost string, opts Binding) error
}

type VhostStore interface {
	ListVhosts() ([]Vhost, error)
}
