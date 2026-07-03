// Package reconcile applies a set of definitions to a RabbitMQ node.
package reconcile

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/definitions"
)

type Provisioner interface {
	DeclareQueue(vhost, name string, opts api.QueueDeclareOpts) error
	PutPolicy(vhost, name string, policy api.Policy) error
	DeclareBinding(vhost string, opts api.Binding) error
}

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
