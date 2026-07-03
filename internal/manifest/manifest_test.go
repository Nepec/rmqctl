package manifest

import (
	"errors"
	"os"
	"testing"

	"github.com/nepec/rmqctl/internal/resource"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name      string
		file      string
		expected  ResourceKind
		expectErr bool
	}{
		{
			name:      "returns a QueueSpec from a queue manifest",
			file:      "../../testdata/queues.yaml",
			expected:  QueueKind,
			expectErr: false,
		},
		{
			name: "returns an ExchangeSpec from an exchange manifest",

			file:      "../../testdata/exchanges.yaml",
			expected:  ExchangeKind,
			expectErr: false,
		},
		{
			name:      "should throw ErrInvalidResourceType for invalid queue type",
			file:      "../../testdata/invalid_type_queue.yaml",
			expected:  QueueKind,
			expectErr: true,
		},
		{
			name:      "should throw ErrInvalidResourceType for invalid exchange type",
			file:      "../../testdata/invalid_type_exchange.yaml",
			expected:  ExchangeKind,
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := os.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("unexpected error while reading test file: %v", err)
			}

			man, err := Parse(raw)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var target resource.ErrInvalidResourceType
				if !errors.As(err, &target) {
					t.Fatalf("wanted error of type %q, got %T", "ErrInvalidResourceType", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if man.Kind != tc.expected {
				t.Fatalf("want kind of %q, got %q", tc.expected, man.Kind)
			}
			if man.Spec == nil {
				t.Fatalf("parsed spec should not be nil")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name string
		file string
	}{
		{
			name: "each queue spec should have a name",
			file: "../../testdata/invalid_name_queue.yaml",
		},
		{
			name: "each exchange spec should have a name",
			file: "../../testdata/invalid_name_exchange.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := os.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("unexpected error while reading test file: %v", err)
			}
			_, err = Parse(raw)
			if err == nil {
				t.Fatalf("expected error, got nil: %v", err)
			}
			var target resource.ErrInvalidResourceSpec
			if !errors.As(err, &target) {
				t.Fatalf("wanted error of type %q, got %T", "ErrInvalidResourceSpec", err)
			}
		})
	}
}
