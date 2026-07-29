//go:build integration

package integration

import (
	"bytes"
	"testing"

	"github.com/nepec/rmqctl/internal/cli"
)

func runApply(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer
	cmd := cli.NewApplyCommand(realClientFactory)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestApplyCommand(t *testing.T) {
	c := newClient(t)
	t.Run("given a queue manifest, it creates queues, bindings and policies", func(t *testing.T) {
		vhost := "test-apply"
		setupVhost(t, c, vhost)
		setupExchange(t, c, vhost, "events")

		out, err := runApply(t, "apply",
			"--vhosts", vhost,
			"--file", "testdata/manifests/queues.yaml",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("dry-run does not actually apply anything", func(t *testing.T) {
		vhost := "empty"
		setupVhost(t, c, vhost)
		out, err := runApply(t, "apply",
			"--vhosts", vhost,
			"--file", "testdata/manifests/queues.yaml",
			"--dry-run",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("idempotency", func(t *testing.T) {
		vhost := "test-idempotent"
		setupVhost(t, c, vhost)
		setupExchange(t, c, vhost, "events")

		out, err := runApply(t, "apply",
			"--vhosts", vhost,
			"--file", "testdata/manifests/queues.yaml",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}
		// 2nd run
		out, err = runApply(t, "apply",
			"--vhosts", vhost,
			"--file", "testdata/manifests/queues.yaml",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("fails for a missing manifest", func(t *testing.T) {
		_, err := runApply(t, "apply",
			"--file", "testdata/manifests/missing.yaml",
		)
		if err == nil {
			t.Fatal("expected file not found error, got nil")
		}
	})
}
