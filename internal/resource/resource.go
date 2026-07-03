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

type SpecFactory func() SpecList

type (
	ErrInvalidResourceType struct {
		resource      string
		requestedType string
	}
	ErrInvalidResourceSpec struct {
		resource string
		message  string
	}
)

type SpecList interface {
	Validate() error
	MarshalDefs(vhostr string) definitions.Definitions
}

func (e ErrInvalidResourceType) Error() string {
	return fmt.Sprintf("invalid type %q for %q resource", e.requestedType, e.resource)
}

func (e ErrInvalidResourceSpec) Error() string {
	return fmt.Sprintf("invalid spec for %q: %s", e.resource, e.message)
}
