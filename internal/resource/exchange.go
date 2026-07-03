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

type ExchangeType string

type ExchangeSpec struct {
	Name string         `yaml:"name"`
	Type ExchangeType   `yaml:"type"`
	Args map[string]any `yaml:"arguments"`
}
type ExchangeSpecList []ExchangeSpec

func (t ExchangeType) IsValid() bool {
	switch t {
	case ExchangeTypeDirect, ExchangeTypeFanout, ExchangeTypeTopic, "":
		return true
	default:
		return false
	}
}

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

func (e ExchangeSpec) Validate() error {
	if e.Name == "" {
		return ErrInvalidResourceSpec{resource: "exchange", message: "name cannot be empty"}
	}
	return nil
}

func (e ExchangeSpec) MarshalDef(vhost string) (definitions.Exchange, bool) {
	args := make(map[string]any, 0)
	maps.Copy(args, e.Args)

	return definitions.Exchange{
		Name:      e.Name,
		Type:      string(e.Type),
		Arguments: args,
	}, true
}

func (el ExchangeSpecList) Validate() error {
	for _, e := range el {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

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
