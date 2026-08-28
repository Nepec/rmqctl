package queue

import (
	"bytes"
	"errors"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/api/apitest"
	"github.com/nepec/rmqctl/internal/cli/sharedopts"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestDeleteQueues(t *testing.T) {
	seedQueues := []api.Queue{
		{Name: "ok"},
		{Name: "fail"},
		{Name: "busy"},
	}
	client := apitest.NewInMemoryClient().WithQueues(seedQueues)
	client.ListQueuesInFunc = func(vhost string) ([]api.Queue, error) {
		qs := make([]api.Queue, len(client.Queues))
		for _, q := range client.Queues {
			qs = append(qs, q)
		}

		return qs, nil
	}
	client.DeleteQueueFunc = func(vhost, name string, force bool) error {
		if strings.Contains(name, "fail") {
			return errors.New("boom")
		}
		delete(client.Queues, name)
		return nil
	}

	t.Run("happy path", func(t *testing.T) {
		var out bytes.Buffer
		deleted, _ := deleteQueues(client, &out, "/", []api.Queue{{Name: "ok"}}, false)

		if deleted != 1 {
			t.Errorf("got %d deleted, want 1", deleted)
		}
		if _, exists := client.Queues["ok"]; exists {
			t.Errorf("'ok' queue should have been deleted")
		}
		if !strings.Contains(out.String(), `deleted queue "ok"`) {
			t.Errorf("expected output to report the deleted queue, got %q", out.String())
		}
	})

	t.Run("generic error should not mention force", func(t *testing.T) {
		var out bytes.Buffer
		_, failed := deleteQueues(client, &out, "/", []api.Queue{{Name: "fail"}}, false)

		if failed != 1 {
			t.Errorf("got %d failed, want 1", failed)
		}
		if _, exists := client.Queues["fail"]; !exists {
			t.Errorf("'fail' queue should still be present after a failed delete")
		}

		if strings.Contains(out.String(), "--force") {
			t.Errorf("unrelated failure should not hint at --force, got %q", out.String())
		}
	})

	t.Run("not safe to delete error hints force", func(t *testing.T) {
		client.DeleteQueueFunc = func(vhost, name string, force bool) error {
			return api.ErrQueueNotSafeToDelete
		}

		var out bytes.Buffer
		_, failed := deleteQueues(client, &out, "/", []api.Queue{{Name: "busy"}}, false)

		if failed != 1 {
			t.Errorf("got %d failed, want 1", failed)
		}
		if !strings.Contains(out.String(), "--force") {
			t.Errorf("expected output to hint at --force, got %q", out.String())
		}
	})
}

func TestDeleteByFilter(t *testing.T) {
	testCases := []struct {
		name        string
		filter      *api.QueueFilterOpts
		deleteOpts  *sharedopts.DeleteOptions
		seedQueues  []api.Queue
		remaining   []string
		wantSummary string
		wantAbsent  string
	}{
		{
			name:       "deletes matched queues",
			filter:     &api.QueueFilterOpts{Contains: "test"},
			deleteOpts: &sharedopts.DeleteOptions{},
			seedQueues: []api.Queue{
				{Name: "test"},
				{Name: "test2"},
			},
			remaining:   nil,
			wantSummary: "Total: 2 deleted, 0 failed",
		},
		{
			name:       "does not delete protected queues without force",
			deleteOpts: &sharedopts.DeleteOptions{},
			seedQueues: []api.Queue{
				{Name: "test", Messages: 5},
				{Name: "test2", Active: true},
			},
			remaining:   []string{"test", "test2"},
			wantSummary: "Total: 0 deleted, 2 failed",
		},
		{
			name:       "does delete protected queues with force",
			deleteOpts: &sharedopts.DeleteOptions{Force: true},
			seedQueues: []api.Queue{
				{Name: "test"},
				{Name: "test2", Active: true},
			},
			remaining:   nil,
			wantSummary: "Total: 2 deleted, 0 failed",
		},
		{
			name:       "dry-run performs no deletions",
			deleteOpts: &sharedopts.DeleteOptions{DryRun: true},
			seedQueues: []api.Queue{
				{Name: "test"},
				{Name: "test2"},
			},
			remaining:   []string{"test", "test2"},
			wantSummary: "(dry run) Total: 2 would be deleted, 0 failed",
		},
		{
			name:       "filter narrows the matched set, force does not widen it beyond the filter",
			filter:     &api.QueueFilterOpts{Empty: boolPtr(false)},
			deleteOpts: &sharedopts.DeleteOptions{Force: true},
			seedQueues: []api.Queue{
				{Name: "empty"},
				{Name: "busy1", Messages: 5},
				{Name: "busy2", Messages: 3},
			},
			remaining:   []string{"empty"},
			wantSummary: "Total: 2 deleted, 0 failed",
		},
		{
			name:       "filtering does not bypass the safety check, force is still required per matched queue",
			filter:     &api.QueueFilterOpts{Empty: boolPtr(false)},
			deleteOpts: &sharedopts.DeleteOptions{},
			seedQueues: []api.Queue{
				{Name: "empty"},
				{Name: "busy1", Messages: 5},
				{Name: "busy2", Messages: 3},
			},
			remaining:   []string{"empty", "busy1", "busy2"},
			wantSummary: "Total: 0 deleted, 2 failed",
			wantAbsent:  `"empty"`,
		},
		{
			name:       "filter matching nothing deletes nothing",
			filter:     &api.QueueFilterOpts{Contains: "nonexistent"},
			deleteOpts: &sharedopts.DeleteOptions{Force: true},
			seedQueues: []api.Queue{
				{Name: "test"},
				{Name: "test2"},
			},
			remaining:   []string{"test", "test2"},
			wantSummary: "Total: 0 deleted, 0 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := apitest.NewInMemoryClient().WithQueues(tc.seedQueues)

			client.ListQueuesInFunc = func(vhost string) ([]api.Queue, error) {
				qs := make([]api.Queue, 0, len(client.Queues))
				for _, q := range client.Queues {
					qs = append(qs, q)
				}

				return qs, nil
			}
			client.DeleteQueueFunc = func(vhost, name string, force bool) error {
				q := client.Queues[name]
				if !force && (q.Messages > 0 || q.Active) {
					return api.ErrQueueNotSafeToDelete
				}
				delete(client.Queues, name)
				return nil
			}

			var out bytes.Buffer
			err := deleteByFilter(&out, client, []string{"/"}, tc.filter, tc.deleteOpts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out.String(), tc.wantSummary) {
				t.Errorf("expected output to contain %q, got %q", tc.wantSummary, out.String())
			}
			if tc.wantAbsent != "" && strings.Contains(out.String(), tc.wantAbsent) {
				t.Errorf("expected output to not contain %q, got %q", tc.wantAbsent, out.String())
			}

			var got []string
			for name := range client.Queues {
				got = append(got, name)
			}
			sort.Strings(got)
			want := append([]string(nil), tc.remaining...)
			sort.Strings(want)

			if !slices.Equal(got, want) {
				t.Fatalf("got queues %v, want %v", got, want)
			}
		})
	}
}

func TestDeleteByName(t *testing.T) {
	t.Run("dry-run performs no deletions", func(t *testing.T) {
		seedQueues := []api.Queue{{Name: "still-here"}}
		deleteOpts := &sharedopts.DeleteOptions{DryRun: true}
		client := apitest.NewInMemoryClient().WithQueues(seedQueues)

		client.ListQueuesInFunc = func(vhost string) ([]api.Queue, error) {
			qs := make([]api.Queue, 0, len(client.Queues))
			for _, q := range client.Queues {
				qs = append(qs, q)
			}

			return qs, nil
		}
		client.DeleteQueueFunc = func(vhost, name string, force bool) error {
			q, ok := client.Queues[name]
			if !ok {
				return api.ErrQueueNotFound
			}
			if !force && (q.Messages > 0 || q.Active) {
				return api.ErrQueueNotSafeToDelete
			}
			delete(client.Queues, name)
			return nil
		}

		var out bytes.Buffer
		err := deleteByName(&out, client, []string{"/"}, "", deleteOpts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(seedQueues) != len(client.Queues) {
			t.Errorf("want %d queues, got %d", len(seedQueues), len(client.Queues))
		}
	})
}
