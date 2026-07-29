//go:build integration

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
	"github.com/nepec/rmqctl/internal/api"
)

const (
	rmqHost = "localhost"
	rmqPort = 15672
	rmqUser = "guest"
	rmqPass = "guest"
)

func realClientFactory() (api.RabbitClient, error) {
	return api.NewRabbitHoleClient(rmqHost, rmqPort, rmqUser, rmqPass)
}

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

func setupExchange(t *testing.T, c *rabbithole.Client, vhost, name string) {
	t.Helper()

	res, err := c.DeclareExchange(vhost, name, rabbithole.ExchangeSettings{
		Type:    "topic",
		Durable: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("could not create exchange for testing: %s", res.Status)
	}

	waitForExchange(t, c, vhost, name)
}

func setupPolicy(t *testing.T, c *rabbithole.Client, vhost, name string, def rabbithole.PolicyDefinition) {
	t.Helper()

	res, err := c.PutPolicy(vhost, name, rabbithole.Policy{
		Name:       name,
		Pattern:    "^" + name + "$",
		Definition: def,
		ApplyTo:    "queues",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("could not create policy for testing: %s", res.Status)
	}
	waitForPolicy(t, c, vhost, name, name)
}

func waitForExchange(t *testing.T, c *rabbithole.Client, vhost, name string) {
	t.Helper()

	timeout := 6 * time.Second
	interval := 100 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for {
		info, err := c.GetExchange(vhost, name)
		if err == nil && info.Name == name {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("exchange %q has not been created within the timeout %v", name, timeout)
		}

		time.Sleep(interval)
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
	delete(def, "rabbit_version")
	delete(def, "rabbitmq_version")
	delete(def, "product_version")
	delete(def, "erlang_version")
	delete(def, "rabbitmq_definition_format")

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
