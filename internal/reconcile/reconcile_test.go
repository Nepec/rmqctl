package reconcile

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/definitions"
)

type mockProvisioner struct {
	declaredQueues   []string
	declaredBindings []api.Binding
	declaredPolicies []string
	err              error
}

func (p *mockProvisioner) DeclareQueue(vhost, name string, opts api.QueueDeclareOpts) error {
	p.declaredQueues = append(p.declaredQueues, name)
	return p.err
}

func (p *mockProvisioner) PutPolicy(vhost, name string, policy api.Policy) error {
	p.declaredPolicies = append(p.declaredPolicies, policy.Name)
	return p.err
}

func (p *mockProvisioner) DeclareBinding(vhost string, opts api.Binding) error {
	p.declaredBindings = append(p.declaredBindings, opts)
	return p.err
}

func TestApply(t *testing.T) {
	defs := definitions.Definitions{
		Queues: []definitions.Queue{
			{Name: "q1", Durable: true, AutoDelete: false, Arguments: map[string]any{"x-queue-type": "quorum"}},
			{Name: "q2", Durable: true, AutoDelete: false, Arguments: map[string]any{"x-queue-type": "classic"}},
		},
		Policies: []definitions.Policy{
			{Name: "q1", Pattern: ".*"},
		},
		Bindings: []definitions.Binding{
			{Source: "q1", Destination: "events", DestinationType: "exchange", RoutingKey: "rk"},
		},
	}

	t.Run("applies all resources", func(t *testing.T) {
		mock := &mockProvisioner{}
		err := Apply(mock, io.Discard, "test", defs, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff([]string{"q1", "q2"}, mock.declaredQueues); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]string{"q1"}, mock.declaredPolicies); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("stops on first error", func(t *testing.T) {
		mock := &mockProvisioner{err: errors.New("unexpected")}
		err := Apply(mock, io.Discard, "test", defs, false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if len(mock.declaredQueues) != 1 {
			t.Fatalf("expected only 1 declared queue before err, got %d", len(mock.declaredQueues))
		}
	})

	t.Run("dryRun skips provisioninig", func(t *testing.T) {
		mock := &mockProvisioner{}
		err := Apply(mock, io.Discard, "test", defs, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(mock.declaredQueues) != 0 {
			t.Errorf("expected no declared queues, got %d", len(mock.declaredQueues))
		}
		if len(mock.declaredPolicies) != 0 {
			t.Errorf("expected no declared policies, got %d", len(mock.declaredPolicies))
		}
		if len(mock.declaredBindings) != 0 {
			t.Errorf("expected no declared bindings, got %d", len(mock.declaredBindings))
		}
	})
}
