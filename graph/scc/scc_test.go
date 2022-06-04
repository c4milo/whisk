package scc

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestTarjanFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		g        map[string][]string
		expected [][]string
	}{
		{
			"it should identify 3 strongly connected components",
			map[string][]string{
				"1": {"2", "5", "8"},
				"2": {"3", "7", "9"},
				"3": {"1", "2", "4", "6"},
				"4": {"5"},
				"5": {"2"},
				"6": {"4"},
				"7": {},
				"8": {"9"},
				"9": {"8"},
			},
			[][]string{
				{"7"},
				{"8", "9"},
				{"6", "5", "4", "3", "2", "1"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			sccs, err := NewTarjan(tt.g).Find()
			c.Assert(err, qt.IsNil, qt.Commentf("err should be nil: %v", err))
			c.Assert(sccs, qt.DeepEquals, tt.expected)
		})
	}
}
