package grep

import (
    "testing"
)

func TestGrepBuffer(t *testing.T) {
	tt := []struct {
		name     string
		size     int
		elements []string
		expected []string
	}{
		{
			name:     "same number of elements as buffer size",
			size:     3,
			elements: []string{"line1", "line2", "line3"},
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "more elements than buffer size",
			size:     3,
			elements: []string{"line1", "line2", "line3", "line4", "line5"},
			expected: []string{"line3", "line4", "line5"},
		},
		{
			name:     "empty buffer",
			size:     0,
			elements: []string{"line1", "line2", "line3"},
			expected: []string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			buffer := NewGrepBuffer(tc.size)
			for _, e := range tc.elements {
				buffer.Push(e)
			}

			got := buffer.Dump()
			if len(got) != len(tc.expected) {
				t.Errorf("Expected length %d but got %d", len(tc.expected), len(got))
			}
			for i, v := range tc.expected {
				if got[i] != v {
					t.Errorf("Expected %s but got %s", v, got[i])
				}
			}
		})
	}
}
