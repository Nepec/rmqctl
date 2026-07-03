package integration

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/mergepolicy"
)

func TestMain(m *testing.M) {
	approvals.UseFolder("testdata")
	slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	os.Exit(m.Run())
}

func TestMergePolicyAction(t *testing.T) {
	c := newClient(t)
	adapter, err := api.NewRabbitHoleAdapter(rmqHost, rmqPort, rmqUser, rmqPass)
	if err != nil {
		t.Fatalf("unexpeted error: %v", err)
	}

	t.Run("creates new policy when queue has none", func(t *testing.T) {
		vhost := "test-merge-no-existing"
		setupVhost(t, c, vhost)
		setupQueue(t, c, vhost, "q1")
		setupQueue(t, c, vhost, "q2")

		err = mergepolicy.MergeQueuePolicyAction(io.Discard, adapter,
			[]string{vhost},
			"testdata/definitions/create_policy_no_existing.json",
			&api.QueueFilterOpts{},
			false,
			false,
		)
		if err != nil {
			t.Fatalf("unexpeted error: %v", err)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("preserves existing policy when force=false", func(t *testing.T) {
		vhost := "test-merge-preserve"
		setupVhost(t, c, vhost)
		setupQueue(t, c, vhost, "q1")

		res, err := c.PutPolicy(vhost, "q1", rabbithole.Policy{
			Name:       "q1",
			Pattern:    "^q1$",
			Definition: rabbithole.PolicyDefinition{"dead-letter-exchange": "dlx", "dead-letter-routing-key": "original-rk"},
			ApplyTo:    "queues",
		})
		if err != nil {
			t.Fatalf("unexpeted error: %v", err)
		}
		defer res.Body.Close()
		waitForPolicy(t, c, vhost, "q1", "q1")

		if res.StatusCode != http.StatusCreated {
			t.Fatalf("could not create policy for testing: %s", res.Status)
		}

		err = mergepolicy.MergeQueuePolicyAction(
			io.Discard, adapter,
			[]string{vhost},
			"testdata/definitions/merge_policy_update.json",
			&api.QueueFilterOpts{},
			false,
			false,
		)
		if err != nil {
			t.Fatalf("unexpeted error: %v", err)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})

	t.Run("overwrite existing policy when force=true", func(t *testing.T) {
		vhost := " test-merge-overwrite"
		setupVhost(t, c, vhost)
		setupQueue(t, c, vhost, "q1")

		res, err := c.PutPolicy(vhost, "q1", rabbithole.Policy{
			Pattern:    "^q1$",
			Definition: rabbithole.PolicyDefinition{"dead-letter-exchange": "dlx", "dead-letter-routing-key": "original-rk"},
			ApplyTo:    "queues",
		})
		if err != nil {
			t.Fatalf("unexpeted error: %v", err)
		}
		defer res.Body.Close()
		waitForPolicy(t, c, vhost, "q1", "q1")

		if res.StatusCode != http.StatusCreated {
			t.Fatalf("could not create policy for testing: %s", res.Status)
		}

		err = mergepolicy.MergeQueuePolicyAction(
			io.Discard, adapter,
			[]string{vhost},
			"testdata/definitions/merge_policy_overwrite.json",
			&api.QueueFilterOpts{},
			false,
			true,
		)
		if err != nil {
			t.Fatalf("unexpeted error: %v", err)
		}

		actual := fetchDefinitions(t, vhost)
		assertGolden(t, actual)
	})
}
