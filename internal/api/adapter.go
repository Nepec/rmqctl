package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

// RabbitHoleClient implements RabbitClient on top of the rabbit-hole
// Management API client.
type RabbitHoleClient struct {
	c *rabbithole.Client
}

// NewRabbitHoleClient builds a RabbitHoleClient targeting the RabbitMQ
// Management API at host:port, authenticating with user/pass.
func NewRabbitHoleClient(host string, port int, user, pass string) (*RabbitHoleClient, error) {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", host, port),
	}
	rh, err := rabbithole.NewClient(u.String(), user, pass)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq %q: %w", u.String(), err)
	}

	return &RabbitHoleClient{c: rh}, nil
}

// Host returns the client's target host, parsed from its endpoint URL.
// It returns "" if the endpoint cannot be parsed.
func (r RabbitHoleClient) Host() string {
	u, err := url.Parse(r.c.Endpoint)
	if err != nil {
		return ""
	}
	return u.Host
}

// Port returns the client's target port, parsed from its endpoint URL.
// It returns 0 if the endpoint cannot be parsed.
func (r RabbitHoleClient) Port() int {
	u, err := url.Parse(r.c.Endpoint)
	if err != nil {
		return 0
	}
	p, err := strconv.Atoi(u.Port())
	if err != nil {
		return 0
	}
	return p
}

// ListVhosts implements VhostStore.
func (r *RabbitHoleClient) ListVhosts() ([]Vhost, error) {
	vs, err := r.c.ListVhosts()
	if err != nil {
		return nil, fmt.Errorf("listing vhosts: %w", err)
	}

	if len(vs) == 0 {
		return []Vhost{}, nil
	}

	vhosts := make([]Vhost, len(vs))
	for i, v := range vs {
		vhosts[i] = Vhost{Name: v.Name}
	}

	return vhosts, nil
}

// ListQueuesIn implements QueueStore.
func (r *RabbitHoleClient) ListQueuesIn(vhost string) ([]Queue, error) {
	qs, err := r.c.ListQueuesIn(vhost)
	if err != nil {
		return nil, fmt.Errorf("listing queues in %q: %w", vhost, err)
	}

	qResp := make([]Queue, len(qs))
	for i, q := range qs {
		qResp[i] = Queue{
			Name:       q.Name,
			Type:       q.Type,
			Vhost:      q.Vhost,
			Messages:   q.Messages,
			Active:     (q.Consumers != 0),
			PolicyName: q.Policy,
		}
	}

	return qResp, nil
}

// DeclareQueue implements QueueStore. opts.Arguments must contain an
// "x-queue-type" entry; DeclareQueue returns an error if it is missing.
func (r *RabbitHoleClient) DeclareQueue(vhost, name string, opts QueueDeclareOpts) error {
	// TODO: improve, too crude. should type be checked here or a default be enforced?
	qt, ok := opts.Arguments["x-queue-type"]
	if !ok {
		return fmt.Errorf("missing queue type")
	}
	settings := rabbithole.QueueSettings{
		Type:       qt.(string), //nolint:errcheck
		Durable:    opts.Durable,
		AutoDelete: opts.AutoDelete,
		Arguments:  opts.Arguments,
	}

	// TODO: check res
	_, err := r.c.DeclareQueue(vhost, name, settings)
	if err != nil {
		return fmt.Errorf("declaring queue %q in %q: %w", name, vhost, err)
	}

	return nil
}

var ErrQueueNotSafeToDelete = errors.New("queue has messages or consumers, not safe to delete")

// DeleteQueue deletes a queue in a vhost. force determines whether to delete
// queues with messages or active consumers.
// Classic queues leverage the 'if-empty' and 'if-active' query params, while
// quorum queues are checked manually.
func (r *RabbitHoleClient) DeleteQueue(vhost, name string, force bool) error {
	opts := rabbithole.QueueDeleteOptions{IfEmpty: !force, IfUnused: !force}
	_, err := r.c.DeleteQueue(vhost, name, opts)
	if err == nil {
		return nil
	}

	if !force && isQuorumSafetyUnsupported(err) {
		return r.deleteQuorumQueueSafely(vhost, name)
	}
	if !force && isClassicQueueSafeToDelete(err) {
		return ErrQueueNotSafeToDelete
	}
	return fmt.Errorf("deleting queue %q in %q: %w", name, vhost, err)
}

func isQuorumSafetyUnsupported(err error) bool {
	// rabbithole.ErrorResponse bundles all >=400 errors
	var rme rabbithole.ErrorResponse
	if !errors.As(err, &rme) {
		return false // Some other kind of error which should be returned
	}

	return rme.StatusCode == http.StatusBadRequest && strings.Contains(rme.Reason, "not supported by quorum queues")
}

// isClassicQueueSafeToDelete reports whether an err from an attempted deletion
// of a classic queue is due to the queue being non-empty or with active
// consumers.
func isClassicQueueSafeToDelete(err error) bool {
	var rme rabbithole.ErrorResponse
	if !errors.As(err, &rme) {
		return false
	}
	if rme.StatusCode != http.StatusBadRequest {
		return false
	}
	return strings.Contains(rme.Reason, "not empty") || strings.Contains(rme.Reason, "in use")
}

// deleteQuorumQueueSafely replicates the if-empty/if-unused check
// client-side, since quorum queues reject those flags outright
// (rabbitmq-server#10543). It re-fetches the queue immediately before
// deciding, rather than trusting an earlier listing, to keep the race
// window as small as the direct-flag path gets on classic queues.
func (r *RabbitHoleClient) deleteQuorumQueueSafely(vhost, name string) error {
	q, err := r.c.GetQueue(vhost, name)
	if err != nil {
		var rme rabbithole.ErrorResponse
		// Match rabbithole idempotent handling of 404s
		if errors.As(err, &rme) && rme.StatusCode == http.StatusNotFound {
			return nil // queue already gone, not an error
		}
		return fmt.Errorf("checking queue %q in %q before deletion: %w", name, vhost, err)
	}
	if q.Messages > 0 || q.Consumers > 0 {
		return ErrQueueNotSafeToDelete
	}
	_, err = r.c.DeleteQueue(vhost, name)
	return err
}

// ListPoliciesIn implements PolicyStore.
func (r *RabbitHoleClient) ListPoliciesIn(vhost string) (Policies, error) {
	pols, err := r.c.ListPoliciesIn(vhost)
	if err != nil {
		return nil, fmt.Errorf("listing policies in %q: %w", vhost, err)
	}

	if len(pols) == 0 {
		return Policies{}, nil
	}

	policies := make(Policies, len(pols))
	for _, p := range pols {
		policies[p.Name] = Policy{
			Name:       p.Name,
			Vhost:      vhost,
			Pattern:    p.Vhost,
			ApplyTo:    p.ApplyTo,
			Definition: p.Definition,
			Priority:   p.Priority,
		}
	}

	return policies, nil
}

// GetPolicy implements PolicyStore.
func (r *RabbitHoleClient) GetPolicy(vhost, name string) (*Policy, error) {
	pol, err := r.c.GetPolicy(vhost, name)
	if err != nil {
		return nil, fmt.Errorf("getting policy %q in %q: %w", name, vhost, err)
	}

	p := Policy{
		Vhost:      pol.Vhost,
		Name:       pol.Name,
		ApplyTo:    pol.ApplyTo,
		Pattern:    pol.Pattern,
		Priority:   pol.Priority,
		Definition: map[string]any(pol.Definition),
	}
	return &p, nil
}

// PutPolicy implements PolicyStore.
func (r *RabbitHoleClient) PutPolicy(vhost, name string, policy Policy) error {
	newPolicy := rabbithole.Policy{
		Vhost:      vhost,
		Pattern:    policy.Pattern,
		ApplyTo:    policy.ApplyTo,
		Name:       name,
		Priority:   policy.Priority,
		Definition: policy.Definition,
	}

	res, err := r.c.PutPolicy(vhost, name, newPolicy)
	if err != nil {
		return fmt.Errorf("putting policy: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	switch res.StatusCode {
	case http.StatusCreated:
		slog.Debug("creating new policy", "vhost", vhost, "policy", policy.Name)
	case http.StatusNoContent:
		slog.Debug("policy updated", "vhost", vhost, "policy", policy.Name)
	default:
		return fmt.Errorf("could not update policy: %s", res.Status)
	}

	return nil
}

// DeclareBinding implements BindingStore.
func (r *RabbitHoleClient) DeclareBinding(vhost string, opts Binding) error {
	b := rabbithole.BindingInfo{
		Source:          opts.Source,
		Vhost:           vhost,
		Destination:     opts.Destination,
		DestinationType: opts.DestinationType,
		RoutingKey:      opts.RoutingKey,
	}
	_, err := r.c.DeclareBinding(vhost, b)
	if err != nil {
		return fmt.Errorf("declaring binding in %q: %w", vhost, err)
	}
	return nil
}
