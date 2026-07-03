// Package manifest defines the shape and parsing rules for a rmqctl resource manifest.
//
// Only specific kinds of manifests exist, each containing a list of specifications which
// detail resource configuration.
package manifest

import (
	"fmt"

	"github.com/nepec/rmqctl/internal/resource"
	"go.yaml.in/yaml/v3"
)

type (
	ResourceKind string
	rawManifest  struct {
		Kind ResourceKind `yaml:"kind"`
		Spec yaml.Node    `yaml:"spec"`
	}
)

type Manifest struct {
	Kind ResourceKind
	Spec resource.SpecList
}

const (
	QueueKind    ResourceKind = "Queue"
	ExchangeKind ResourceKind = "Exchange"
)

var registry = map[ResourceKind]resource.SpecFactory{
	QueueKind:    func() resource.SpecList { return &resource.QueueSpecList{} },
	ExchangeKind: func() resource.SpecList { return &resource.ExchangeSpecList{} },
}

func Parse(data []byte) (Manifest, error) {
	var raw rawManifest
	err := yaml.Unmarshal(data, &raw)
	if err != nil {
		return Manifest{}, fmt.Errorf("parsing manifest: %w", err)
	}

	specFactory, ok := registry[raw.Kind]
	if !ok {
		return Manifest{}, fmt.Errorf("unknown kind: %q", raw.Kind)
	}

	specs := specFactory()
	if err := raw.Spec.Decode(specs); err != nil {
		return Manifest{}, fmt.Errorf("decoding %q spec: %w", raw.Kind, err)
	}

	if err := specs.Validate(); err != nil {
		return Manifest{}, err
	}

	return Manifest{
		Kind: raw.Kind,
		Spec: specs,
	}, nil
}
