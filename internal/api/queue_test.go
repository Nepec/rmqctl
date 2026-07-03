package api_test

import (
	"testing"

	"github.com/nepec/rmqctl/internal/api"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestFilterMatches(t *testing.T) {
	testCases := []struct {
		name     string
		filter   *api.QueueFilterOpts
		queue    api.Queue
		expected bool
	}{
		{
			name:     "name contains string",
			filter:   &api.QueueFilterOpts{Contains: "substr"},
			queue:    api.Queue{Name: "contains.substr"},
			expected: true,
		},
		{
			name:     "type quorum matches",
			filter:   &api.QueueFilterOpts{Type: "quorum"},
			queue:    api.Queue{Type: "quorum"},
			expected: true,
		},
		{
			name:     "active queue matches",
			filter:   &api.QueueFilterOpts{Active: boolPtr(true)},
			queue:    api.Queue{Active: true},
			expected: true,
		},
		{
			name:     "empty queue matches",
			filter:   &api.QueueFilterOpts{Empty: boolPtr(true)},
			queue:    api.Queue{Messages: 0},
			expected: true,
		},
		{
			name:     "empty queue does not match",
			filter:   &api.QueueFilterOpts{Empty: boolPtr(true)},
			queue:    api.Queue{Messages: 1},
			expected: false,
		},
		{
			name:     "queue with policy matches",
			filter:   &api.QueueFilterOpts{WithPolicy: boolPtr(true)},
			queue:    api.Queue{PolicyName: "some.policy"},
			expected: true,
		},
		{
			name:     "queue with no policy does not match",
			filter:   &api.QueueFilterOpts{WithPolicy: boolPtr(true)},
			queue:    api.Queue{Name: "no.policy.q"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.Matches(tc.queue)
			if got != tc.expected {
				t.Errorf("got %t, want %t", got, tc.expected)
			}
		})
	}
}
