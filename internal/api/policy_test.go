package api

import (
	"reflect"
	"testing"
)

func TestMegePolicyDefinitions(t *testing.T) {
	t.Run("conflict preserves current when force=false", func(t *testing.T) {
		current := map[string]any{"dead-letter-exchange": "original-dlx"}
		update := map[string]any{"dead-letter-exchange": "new-dlx"}

		result := MergePolicyDefinitions(current, update, false)

		reflect.DeepEqual(map[string]any{"dead-letter-exchange": "original-dlx"}, result)
	})

	t.Run("conflict overwrites current when force=true", func(t *testing.T) {
		current := map[string]any{"dead-letter-exchange": "original-dlx"}
		update := map[string]any{"dead-letter-exchange": "new-dlx"}

		result := MergePolicyDefinitions(current, update, true)

		reflect.DeepEqual(map[string]any{"dead-letter-exchange": "new-dlx"}, result)
	})

	t.Run("update preserves old keys", func(t *testing.T) {
		current := map[string]any{"dead-letter-exchange": "original-dlx"}
		update := map[string]any{"dead-letter-routing-key": "rk"}

		result := MergePolicyDefinitions(current, update, false)

		reflect.DeepEqual(map[string]any{"dead-letter-exchange": "original-dlx", "dead-letter-routing-key": "rk"}, result)
	})
}
