package resource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInlinePolicyMarshalDef(t *testing.T) {
	t.Run("policy must meet inline requirements", func(t *testing.T) {
		queueName := "test.queue+"
		patter := `^test\.queue\+$`
		applyTo := "queues"
		policyDef := map[string]any{"dead-letter-exchange": "dead-letters"}

		ps := InlinePolicy{"dead-letter-exchange": "dead-letters"}

		got, _ := ps.MarshalDef("test.vhost", queueName)

		if got.Name != queueName {
			t.Errorf("name must be same as associated queue: want %q, got %q", queueName, got.Name)
		}
		if got.Pattern != patter {
			t.Errorf("pattern must be exact regex match of queue name, with proper escapes. want %q, got %q", patter, got.Pattern)
		}
		if got.ApplyTo != applyTo {
			t.Errorf("must apply to queues, got %q", got.ApplyTo)
		}

		if diff := cmp.Diff(policyDef, got.Definition); diff != "" {
			t.Errorf("defintions mismatch (-want +got):\n%s", diff)
		}
	})
}
