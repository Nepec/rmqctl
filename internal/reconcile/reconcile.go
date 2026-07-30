// Package reconcile applies a set of definitions to a RabbitMQ node.
//
// Apply is the package's entry point: given a dry-run flag it either
// prints the definitions that would be created, or declares them against
// a Provisioner, writing one progress line per resource to the given
// io.Writer as it goes.
package reconcile

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/definitions"
)

// Provisioner is the minmal client that Apply needs to
// create resources on a RabbitMQ node.
type Provisioner interface {
	// DeclareQueue creates or updates the queue named name on vhost.
	DeclareQueue(vhost, name string, opts api.QueueDeclareOpts) error
	// PutPolicy creates or overwrites the policy named name on vhost.
	PutPolicy(vhost, name string, policy api.Policy) error
	// DeclareBinding creates the binding described by opts on vhost.
	DeclareBinding(vhost string, opts api.Binding) error
}

// Apply provisions defs on vhost using the Provisioner, in order: queues,
// then policies, then bindings. If dryRun is true, nothing is provisioned
// and defs is instead written out as indented JSON. Apply stops and
// returns the first error encountered.
func Apply(c Provisioner, out io.Writer, vhost string, defs definitions.Definitions, dryRun bool) error {
	if dryRun {
		// TODO: print defs
		d, err := json.MarshalIndent(defs, "", "  ")
		if err != nil {
			return err
		}
		_, _ = out.Write(d)
		return nil
	}

	for _, q := range defs.Queues {
		opts := api.QueueDeclareOpts{
			Durable:    q.Durable,
			AutoDelete: q.AutoDelete,
			Arguments:  q.Arguments,
		}
		err := c.DeclareQueue(vhost, q.Name, opts)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "queue/%s configured\n", q.Name)
	}
	for _, p := range defs.Policies {
		policy := api.Policy{
			Name:       p.Name,
			Pattern:    p.Pattern,
			ApplyTo:    p.ApplyTo,
			Definition: p.Definition,
			Priority:   p.Priority,
		}
		err := c.PutPolicy(vhost, p.Name, policy)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "policy/%s configured\n", p.Name)
	}
	for _, b := range defs.Bindings {
		binding := api.Binding{
			Source:          b.Source,
			Destination:     b.Destination,
			DestinationType: b.DestinationType,
			RoutingKey:      b.RoutingKey,
		}
		err := c.DeclareBinding(vhost, binding)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "binding/%s->%s configured\n", b.Source, b.Destination)
	}
	return nil
}
