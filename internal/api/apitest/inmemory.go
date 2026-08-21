// Package apitest provides a configurable fake implementing
// api.RabbitClient, for unit tests that exercise CLI command logic
// without a live RabbitMQ node.
package apitest

import "github.com/nepec/rmqctl/internal/api"

// InMemoryClient is an api.RabbitClient test double. Each XxxFunc field
// stands in for the like-named interface method; a test sets only the
// ones the code under test actually calls. Calling DeclareQueue,
// ListPoliciesIn, GetPolicy, PutPolicy, DeclareBinding, or ListVhosts
// with its func field unset panics, so a dependency the test forgot to
// configure fails loudly instead of silently returning a zero value.
// Host and Port default to "" and 0 instead of panicking, since they're
// usually incidental (error-message context) rather than behavior under
// test.
type InMemoryClient struct {
	Queues map[string]api.Queue

	HostFunc func() string
	PortFunc func() int

	ListQueuesInFunc func(vhost string) ([]api.Queue, error)
	DeclareQueueFunc func(vhost, name string, opts api.QueueDeclareOpts) error
	DeleteQueueFunc  func(vhost, name string, force bool) error

	ListPoliciesInFunc func(vhost string) (api.Policies, error)
	GetPolicyFunc      func(vhost, name string) (*api.Policy, error)
	PutPolicyFunc      func(vhost, name string, policy api.Policy) error

	DeclareBindingFunc func(vhost string, opts api.Binding) error

	ListVhostsFunc func() ([]api.Vhost, error)
}

func NewInMemoryClient() *InMemoryClient {
	return &InMemoryClient{}
}

// WithQueues adds a seed of queues to the struct
func (i *InMemoryClient) WithQueues(queues []api.Queue) *InMemoryClient {
	qm := make(map[string]api.Queue, len(queues))
	for _, q := range queues {
		qm[q.Name] = q
	}
	i.Queues = qm

	return i
}

// compile-time check that FakeClient stays in sync with api.RabbitClient.
var _ api.RabbitClient = (*InMemoryClient)(nil)

func (i *InMemoryClient) Host() string {
	if i.HostFunc == nil {
		return ""
	}
	return i.HostFunc()
}

func (i *InMemoryClient) Port() int {
	if i.PortFunc == nil {
		return 0
	}
	return i.PortFunc()
}

func (i *InMemoryClient) ListQueuesIn(vhost string) ([]api.Queue, error) {
	if i.ListQueuesInFunc == nil {
		panic("InMemoryClient: ListQueuesIn not implemented")
	}
	return i.ListQueuesInFunc(vhost)
}

func (i *InMemoryClient) DeclareQueue(vhost, name string, opts api.QueueDeclareOpts) error {
	if i.DeclareQueueFunc == nil {
		panic("InMemoryClient: DeclareQueueFunc not set")
	}
	return i.DeclareQueueFunc(vhost, name, opts)
}

func (i *InMemoryClient) DeleteQueue(vhost, name string, force bool) error {
	if i.DeleteQueueFunc == nil {
		panic("InMemoryClient: DeleteQueue not implemented")
	}
	return i.DeleteQueueFunc(vhost, name, force)
}

func (i *InMemoryClient) ListPoliciesIn(vhost string) (api.Policies, error) {
	if i.ListPoliciesInFunc == nil {
		panic("InMemoryClient: ListPoliciesInFunc not set")
	}
	return i.ListPoliciesInFunc(vhost)
}

func (i *InMemoryClient) GetPolicy(vhost, name string) (*api.Policy, error) {
	if i.GetPolicyFunc == nil {
		panic("InMemoryClient: GetPolicyFunc not set")
	}
	return i.GetPolicyFunc(vhost, name)
}

func (i *InMemoryClient) PutPolicy(vhost, name string, policy api.Policy) error {
	if i.PutPolicyFunc == nil {
		panic("InMemoryClient: PutPolicyFunc not set")
	}
	return i.PutPolicyFunc(vhost, name, policy)
}

func (i *InMemoryClient) DeclareBinding(vhost string, opts api.Binding) error {
	if i.DeclareBindingFunc == nil {
		panic("InMemoryClient: DeclareBindingFunc not set")
	}
	return i.DeclareBindingFunc(vhost, opts)
}

func (i *InMemoryClient) ListVhosts() ([]api.Vhost, error) {
	if i.ListVhostsFunc == nil {
		panic("InMemoryClient: ListVhostsFunc not set")
	}
	return i.ListVhostsFunc()
}
