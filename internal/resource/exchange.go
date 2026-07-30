package resource

import (
	"maps"

	"github.com/nepec/rmqctl/internal/definitions"
	"go.yaml.in/yaml/v3"
)

const (
	ExchangeTypeDirect ExchangeType = "direct"
	ExchangeTypeFanout ExchangeType = "fanout"
	ExchangeTypeTopic  ExchangeType = "topic"
)

// ExchangeType is a RabbitMQ exchange type (direct, fanout, or topic).
type ExchangeType string

// ExchangeSpec declares an exchange.
type ExchangeSpec struct {
	Name string         `yaml:"name"`
	Type ExchangeType   `yaml:"type"`
	Args map[string]any `yaml:"arguments"`
}

// ExchangeSpecList is a manifest's list of exchange specs. It implements
// SpecList.
type ExchangeSpecList []ExchangeSpec

// IsValid reports whether t is a known exchange type. An empty
// ExchangeType is treated as valid, deferring to the vhost's default.
func (t ExchangeType) IsValid() bool {
	switch t {
	case ExchangeTypeDirect, ExchangeTypeFanout, ExchangeTypeTopic, "":
		return true
	default:
		return false
	}
}

// UnmarshalYAML decodes a YAML scalar into an ExchangeType, rejecting
// values that fail IsValid with an ErrInvalidResourceType.
func (t *ExchangeType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	et := ExchangeType(s)
	if !et.IsValid() {
		return ErrInvalidResourceType{resource: "exchange", requestedType: s}
	}

	*t = et

	return nil
}

// Validate reports whether an ExchangeSpec is well-formed.
// An exchange spec requires a non-empty Name.
func (e ExchangeSpec) Validate() error {
	if e.Name == "" {
		return ErrInvalidResourceSpec{resource: "exchange", message: "name cannot be empty"}
	}
	return nil
}

// MarshalDef translates an ExchangeSpec into a definitions.Exchange.
// The returned bool reports whether the exchange should be included
// in the output; for ExchangeSpec it is always true.
func (e ExchangeSpec) MarshalDef(vhost string) (definitions.Exchange, bool) {
	args := make(map[string]any, 0)
	maps.Copy(args, e.Args)

	return definitions.Exchange{
		Name:      e.Name,
		Type:      string(e.Type),
		Arguments: args,
	}, true
}

// Validate reports whether every spec in the list is well-formed.
func (el ExchangeSpecList) Validate() error {
	for _, e := range el {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MarshalDefs translates every spec in the list into definitions.
func (el ExchangeSpecList) MarshalDefs(vhost string) definitions.Definitions {
	var exchanges []definitions.Exchange

	for _, es := range el {
		if ed, ok := es.MarshalDef(vhost); ok {
			exchanges = append(exchanges, ed)
		}
	}
	return definitions.Definitions{
		Exchanges: exchanges,
	}
}
