package resource

import (
	"maps"

	"github.com/nepec/rmqctl/internal/definitions"
	"go.yaml.in/yaml/v3"
)

// QueueType is a RabbitMQ queue type (quorum, classic, or stream).
type QueueType string

// QueueSpec declares a queue and, inline, the bindings and policy that
// should exist alongside it once applied.
type QueueSpec struct {
	Name         string         `yaml:"name"`
	Type         QueueType      `yaml:"type"`
	Args         map[string]any `yaml:"arguments"`
	Bindings     []QueueBinding `yaml:"bindings"`
	InlinePolicy InlinePolicy   `yaml:"inlinePolicy"`
}

// QueueSpecList is a manifest's list of queue specs. It implements SpecList.
type QueueSpecList []QueueSpec

const (
	QueueTypeQuorum  QueueType = "quorum"
	QueueTypeClassic QueueType = "classic"
	QueueTypeStream  QueueType = "stream"
)

// IsValid reports whether t is a known queue type. An empty QueueType is
// treated as valid, deferring to the vhost's default queue type.
func (t QueueType) IsValid() bool {
	switch t {
	// TODO: queueType may be empty. In such case the vhosts default queue-type should be applied
	case QueueTypeQuorum, QueueTypeClassic, QueueTypeStream, "":
		return true
	default:
		return false
	}
}

// UnmarshalYAML decodes a YAML scalar into t, rejecting values that fail
// IsValid with an ErrInvalidResourceType.
func (t *QueueType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	qt := QueueType(s)
	if !qt.IsValid() {
		return ErrInvalidResourceType{resource: "queue", requestedType: s}
	}

	*t = qt

	return nil
}

// TODO: make UnmarshalYAML method for QueueSpec. A spec with empty name is not a valid spec, both semantiaclly and structurally.
// Should fail fast during decoding

// Validate reports whether q is well-formed. A queue spec requires a
// non-empty Name.
func (q QueueSpec) Validate() error {
	if q.Name == "" {
		return ErrInvalidResourceSpec{resource: "queue", message: "name cannot be empty"}
	}
	return nil
}

// MarshalDef translates q into a definitions.Queue scoped to vhost. The
// returned bool reports whether the queue should be included in the
// output; for QueueSpec it is always true. The queue is always declared
// durable and non-auto-delete, and q.Type is injected into Arguments as
// "x-queue-type".
func (q QueueSpec) MarshalDef(vhost string) (definitions.Queue, bool) {
	args := make(map[string]any, 0)
	maps.Copy(args, q.Args)
	args["x-queue-type"] = string(q.Type)

	return definitions.Queue{
		Name:       q.Name,
		Vhost:      vhost,
		Durable:    true,
		AutoDelete: false,
		Arguments:  args,
	}, true
}

// Validate reports whether every spec in ql is well-formed.
func (ql QueueSpecList) Validate() error {
	for _, q := range ql {
		if err := q.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MarshalDefs translates every spec in ql — along with each queue's
// inline bindings and policy — into definitions scoped to vhost.
func (ql QueueSpecList) MarshalDefs(vhost string) definitions.Definitions {
	var queues []definitions.Queue
	var bindings []definitions.Binding
	var policies []definitions.Policy

	for _, qs := range ql {
		if qd, ok := qs.MarshalDef(vhost); ok {
			queues = append(queues, qd)
		}

		for _, b := range qs.Bindings {
			if bd, ok := b.MarshalDef(vhost, qs.Name); ok {
				bindings = append(bindings, bd)
			}
		}

		if pd, ok := qs.InlinePolicy.MarshalDef(vhost, qs.Name); ok {
			policies = append(policies, pd)
		}
	}

	return definitions.Definitions{
		Queues:   queues,
		Bindings: bindings,
		Policies: policies,
	}
}
