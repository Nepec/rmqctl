// Package resource defines the Specs of various rmqctl resources.
//
// Specs contain details on how to configure the relative resources.
// Resources are meant to be the logical objects defined in RabbitMQ i.e.
// queues, policies, users and so on.
// Each spec has a way to validate itself and to translate itself into a
// RabbitMQ definition.
package resource

import (
	"fmt"

	"github.com/nepec/rmqctl/internal/definitions"
)

// SpecFactory builds an empty SpecList for a manifest kind. It's the
// constructor registered per-kind so a manifest can be decoded into the
// right concrete spec type.
type SpecFactory func() SpecList

type (
	// ErrInvalidResourceType reports that a resource was declared with a
	// type it doesn't recognize (e.g. an unknown queue or exchange type).
	ErrInvalidResourceType struct {
		resource      string
		requestedType string
	}
	// ErrInvalidResourceSpec reports that a resource spec failed
	// validation, e.g. a required field was left empty.
	ErrInvalidResourceSpec struct {
		resource string
		message  string
	}
)

// SpecList is a validated, homogeneous list of resource specs decoded
// from a manifest (e.g. QueueSpecList, ExchangeSpecList).
type SpecList interface {
	// Validate reports whether every spec in the list is well-formed.
	Validate() error
	// MarshalDefs translates every spec in the list into RabbitMQ
	// definitions scoped to vhost.
	MarshalDefs(vhost string) definitions.Definitions
}

func (e ErrInvalidResourceType) Error() string {
	return fmt.Sprintf("invalid type %q for %q resource", e.requestedType, e.resource)
}

func (e ErrInvalidResourceSpec) Error() string {
	return fmt.Sprintf("invalid spec for %q: %s", e.resource, e.message)
}
