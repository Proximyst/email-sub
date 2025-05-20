package batching_test

import (
	"testing"

	"github.com/proximyst/email-sub/pkg/batching"
	"github.com/stretchr/testify/require"
)

func TestBatching(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     []int
		batchSize int
		expected  [][]int
	}{
		"nil": {
			input:     nil,
			batchSize: 1,
			expected:  nil,
		},
		"empty": {
			input:     []int{},
			batchSize: 1,
			expected:  nil,
		},
		"single": {
			input:     []int{1},
			batchSize: 1,
			expected:  [][]int{{1}},
		},
		"single with larger size": {
			input:     []int{1},
			batchSize: 2,
			expected:  [][]int{{1}},
		},
		"multiple in unit batches": {
			input:     []int{1, 2, 3},
			batchSize: 1,
			expected:  [][]int{{1}, {2}, {3}},
		},
		"multiple in larger, even batches": {
			input:     []int{1, 2, 3, 4},
			batchSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}},
		},
		"multiple in larger, odd batches": {
			input:     []int{1, 2, 3, 4, 5},
			batchSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}, {5}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := batching.Batch(test.input, test.batchSize)
			require.Equal(t, test.expected, actual)
		})
	}
}
