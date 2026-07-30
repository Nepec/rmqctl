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

// ResourceKind identifies which kind of resource a manifest declares
// (e.g. "Queue", "Exchange"), determining how its spec section is decoded.
type (
	ResourceKind string
	// rawManifest is the first-pass decode target: it captures Kind so
	// the right resource.SpecList type can be picked before Spec itself
	// is decoded.
	rawManifest struct {
		Kind ResourceKind `yaml:"kind"`
		Spec yaml.Node    `yaml:"spec"`
	}
)

// Manifest is a parsed rmqctl resource manifest: a Kind and the list of
// resource specs decoded for it.
type Manifest struct {
	Kind ResourceKind
	Spec resource.SpecList
}

const (
	QueueKind    ResourceKind = "Queue"
	ExchangeKind ResourceKind = "Exchange"
)

// registry maps each supported ResourceKind to the constructor for its
// resource.SpecList. Adding a new manifest kind means adding an entry here.
var registry = map[ResourceKind]resource.SpecFactory{
	QueueKind:    func() resource.SpecList { return &resource.QueueSpecList{} },
	ExchangeKind: func() resource.SpecList { return &resource.ExchangeSpecList{} },
}

// Parse decodes data as a rmqctl manifest and validates its specs. It
// returns an error if the YAML is malformed, the manifest's kind is
// unrecognized, its spec section doesn't match that kind's shape, or spec
// validation fails.
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
