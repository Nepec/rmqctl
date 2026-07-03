package resource

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/definitions"
)

type FakeRMQClient struct {
	resources map[string][]any

	declareQueue func(map[string][]any, string, string, api.QueueDeclareOpts) (*http.Response, error)
}

func (f FakeRMQClient) ListQueuesIn(vhost string) ([]api.Queue, error) {
	return nil, nil
}

func (f FakeRMQClient) ListVhosts() ([]api.Vhost, error) {
	return nil, nil
}

func (f FakeRMQClient) ListPoliciesIn(vhost string) ([]api.Policy, error) {
	return nil, nil
}

func (f FakeRMQClient) GetPolicy(vhost, name string) (*api.Policy, error) {
	return nil, nil
}

func (f FakeRMQClient) PutPolicy(vhost, name string, policy api.Policy) (*http.Response, error) {
	return nil, nil
}

func (f FakeRMQClient) DeclareQueue(vhost, name string, opts api.QueueDeclareOpts) (*http.Response, error) {
	f.declareQueue(f.resources, vhost, name, opts)

	return &http.Response{}, nil
}

func NewFakeClient() FakeRMQClient {
	resources := make(map[string][]any)
	resources["queue"] = make([]any, 0)
	resources["exchanges"] = make([]any, 0)
	resources["bindings"] = make([]any, 0)
	resources["policies"] = make([]any, 0)

	return FakeRMQClient{resources: resources}
}

func TestMarshalDefinition(t *testing.T) {
	testCases := []struct {
		name string
		spec QueueSpec
		want definitions.Queue
	}{
		{
			name: "queue definition must contain queue x-queue-type arg",
			spec: QueueSpec{Name: "q", Type: QueueTypeQuorum},
			want: definitions.Queue{
				Name:       "q",
				Vhost:      "vhost.test",
				Durable:    true,
				AutoDelete: false,
				Arguments: map[string]any{
					"x-queue-type": string(QueueTypeQuorum),
				},
			},
		},
		{
			name: "spec args must be present in definition",
			spec: QueueSpec{Name: "q", Type: QueueTypeQuorum, Args: map[string]any{"x-message-ttl": 30}},
			want: definitions.Queue{
				Name:       "q",
				Vhost:      "vhost.test",
				Durable:    true,
				AutoDelete: false,
				Arguments: map[string]any{
					"x-queue-type":  string(QueueTypeQuorum),
					"x-message-ttl": 30,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.spec.MarshalDef("vhost.test")

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("defintions mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarshalDefs(t *testing.T) {
	testCases := []struct {
		name  string
		specs QueueSpecList
		vhost string
		want  definitions.Definitions
	}{
		{
			name: "aggregates multiple specs",
			specs: QueueSpecList{
				{Name: "queue.a", Type: QueueTypeQuorum, Bindings: []QueueBinding{{Exchange: "events", RoutingKey: "a"}}, InlinePolicy: InlinePolicy{"dead-letter-exchange": "exchA"}},
				{Name: "queue.b", Type: QueueTypeClassic, Bindings: []QueueBinding{{Exchange: "events", RoutingKey: "b"}}, InlinePolicy: InlinePolicy{"dead-letter-exchange": "exchB"}},
			},
			vhost: "test.vhost",
			want: definitions.Definitions{
				Queues: []definitions.Queue{
					{
						Name:       "queue.a",
						Vhost:      "test.vhost",
						Durable:    true,
						AutoDelete: false,
						Arguments: map[string]any{
							"x-queue-type": string(QueueTypeQuorum),
						},
					},
					{
						Name:       "queue.b",
						Vhost:      "test.vhost",
						Durable:    true,
						AutoDelete: false,
						Arguments: map[string]any{
							"x-queue-type": string(QueueTypeClassic),
						},
					},
				},
				Bindings: []definitions.Binding{
					{
						Source:          "events",
						Vhost:           "test.vhost",
						Destination:     "queue.a",
						DestinationType: "queue",
						RoutingKey:      "a",
					},
					{
						Source:          "events",
						Vhost:           "test.vhost",
						Destination:     "queue.b",
						DestinationType: "queue",
						RoutingKey:      "b",
					},
				},
				Policies: []definitions.Policy{
					{
						Vhost:   "test.vhost",
						Name:    "queue.a",
						Pattern: `^queue\.a$`,
						ApplyTo: "queues",
						Definition: map[string]any{
							"dead-letter-exchange": "exchA",
						},
						Priority: 0,
					},
					{
						Vhost:   "test.vhost",
						Name:    "queue.b",
						Pattern: `^queue\.b$`,
						ApplyTo: "queues",
						Definition: map[string]any{
							"dead-letter-exchange": "exchB",
						},
						Priority: 0,
					},
				},
			},
		},
		{
			name:  "contains queues and bindings",
			specs: QueueSpecList{{Name: "queue.only.bindings", Type: QueueTypeClassic, Bindings: []QueueBinding{{Exchange: "events", RoutingKey: "bindings"}}}},
			vhost: "test.vhost",
			want: definitions.Definitions{
				Queues: []definitions.Queue{
					{
						Name:       "queue.only.bindings",
						Vhost:      "test.vhost",
						Durable:    true,
						AutoDelete: false,
						Arguments: map[string]any{
							"x-queue-type": string(QueueTypeClassic),
						},
					},
				},
				Bindings: []definitions.Binding{
					{
						Source:          "events",
						Vhost:           "test.vhost",
						Destination:     "queue.only.bindings",
						DestinationType: "queue",
						RoutingKey:      "bindings",
					},
				},
			},
		},
		{
			name:  "contains queues and policies",
			specs: QueueSpecList{{Name: "queue.only.policy", Type: QueueTypeQuorum, InlinePolicy: InlinePolicy{"dead-letter-exchange": "dead-letters"}}},
			vhost: "test.vhost",
			want: definitions.Definitions{
				Queues: []definitions.Queue{
					{
						Name:       "queue.only.policy",
						Vhost:      "test.vhost",
						Durable:    true,
						AutoDelete: false,
						Arguments: map[string]any{
							"x-queue-type": string(QueueTypeQuorum),
						},
					},
				},
				Policies: []definitions.Policy{
					{
						Vhost:   "test.vhost",
						Name:    "queue.only.policy",
						Pattern: `^queue\.only\.policy$`,
						ApplyTo: "queues",
						Definition: map[string]any{
							"dead-letter-exchange": "dead-letters",
						},
						Priority: 0,
					},
				},
			},
		},
		{
			name:  "no definitions on an empty spec",
			specs: QueueSpecList{},
			vhost: "test.vhost",
			want:  definitions.Definitions{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defs := tc.specs.MarshalDefs(tc.vhost)
			if diff := cmp.Diff(tc.want, defs); diff != "" {
				t.Errorf("defintions mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
