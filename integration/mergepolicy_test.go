//go:build integration

package integration

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/nepec/rmqctl/internal/cli"
)

func TestMain(m *testing.M) {
	approvals.UseFolder("testdata")
	slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	os.Exit(m.Run())
}

func runMerge(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer
	cmd := cli.NewMergeCommand(realClientFactory)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestMergeQueuePolicy(t *testing.T) {
	c := newClient(t)

	t.Run("creates new policy when queue has none", func(t *testing.T) {
		vhost := "test-merge-no-existing"
		setupVhost(t, c, vhost)
		setupQueue(t, c, vhost, "q1")
		setupQueue(t, c, vhost, "q2")

		out, err := runMerge(t, "policy",
			"--apply-to", "queues",
			"--vhosts", vhost,
			"--file", "testdata/definitions/create_policy_no_existing.json",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("preserves existing policy definitions when force=false", func(t *testing.T) {
		vhost := "test-merge-preserve"
		setupVhost(t, c, vhost)
		setupQueue(t, c, vhost, "q1")
		setupPolicy(t, c, vhost, "q1", rabbithole.PolicyDefinition{
			"dead-letter-exchange":    "dlx",
			"dead-letter-routing-key": "original-rk",
		})

		out, err := runMerge(t, "policy",
			"--apply-to", "queues",
			"--vhosts", vhost,
			"--file", "testdata/definitions/merge_policy_update.json",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("overwrite existing policy when force=true", func(t *testing.T) {
		vhost := " test-merge-overwrite"
		setupVhost(t, c, vhost)
		setupQueue(t, c, vhost, "q1")
		setupPolicy(t, c, vhost, "q1", rabbithole.PolicyDefinition{
			"dead-letter-exchange":    "dlx",
			"dead-letter-routing-key": "original-rk",
		})

		out, err := runMerge(t, "policy",
			"--apply-to", "queues",
			"--vhosts", vhost,
			"--file", "testdata/definitions/merge_policy_overwrite.json",
			"--force",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v\ncmd output:\n%s", err, out)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("fails for invalid applyTo resource type", func(t *testing.T) {
		_, err := runMerge(t, "policy", "--apply-to", "bogus")
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("fails for invalid queue type", func(t *testing.T) {
		_, err := runMerge(t, "policy", "--type", "bogus")
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("fails for non-existent vhost", func(t *testing.T) {
		_, err := runMerge(t, "policy", "--vhosts", "non-existent")
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("fails for non-existent definitions file", func(t *testing.T) {
		_, err := runMerge(t, "policy", "--file", "testdata/definitions/does-not-exists.json")
		if err == nil {
			t.Fatal("expected file not found error, got nil")
		}
	})
}
