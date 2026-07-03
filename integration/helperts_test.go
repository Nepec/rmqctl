package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	approvals "github.com/approvals/go-approval-tests"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

const (
	rmqHost = "localhost"
	rmqPort = 15672
	rmqUser = "guest"
	rmqPass = "guest"
)

func newClient(t *testing.T) *rabbithole.Client {
	t.Helper()

	rmqURL := fmt.Sprintf("http://%s:%d", rmqHost, rmqPort)
	c, err := rabbithole.NewClient(rmqURL, rmqUser, rmqPass)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return c
}

func setupVhost(t *testing.T, c *rabbithole.Client, vhost string) {
	t.Helper()

	_, err := c.PutVhost(vhost, rabbithole.VhostSettings{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Cleanup(func() {
		_, err := c.DeleteVhost(vhost)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func setupQueue(t *testing.T, c *rabbithole.Client, vhost, name string) {
	t.Helper()

	_, err := c.DeclareQueue(vhost, name, rabbithole.QueueSettings{Durable: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func waitForPolicy(t *testing.T, c *rabbithole.Client, vhost, queueName, policyName string) {
	t.Helper()

	timeout := 6 * time.Second
	interval := 100 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for {
		info, err := c.GetQueue(vhost, queueName)
		if err == nil && info.Policy == policyName {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("policy %q not applied to queue %q within %v", policyName, queueName, timeout)
		}

		time.Sleep(interval)
	}
}

func fetchDefinitions(t *testing.T, vhost string) map[string]any {
	t.Helper()

	rmqURL := fmt.Sprintf("http://%s:%d", rmqHost, rmqPort)
	url := fmt.Sprintf("%s/api/definitions/%s", rmqURL, vhost)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(rmqUser, rmqPass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var def map[string]any
	err = json.Unmarshal(body, &def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return scrubVolatileFields(def)
}

func scrubVolatileFields(def map[string]any) map[string]any {
	delete(def, "rabbitmq_version")
	delete(def, "product_version")
	delete(def, "erlang_version")

	if queues, ok := def["queues"].([]any); ok {
		for _, q := range queues {
			if qm, ok := q.(map[string]any); ok {
				delete(qm, "message_stats")
				delete(qm, "messages")
			}
		}
	}

	return def
}

func assertGolden(t *testing.T, actual map[string]any) {
	t.Helper()

	actualBytes, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	approvals.VerifyString(t, string(actualBytes))
}
