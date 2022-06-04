package cycle

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
			"it should identify 6 distinct cycles",
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
				{"1", "2", "3", "1"},
				{"1", "5", "2", "3", "1"},
				{"2", "3", "2"},
				{"2", "3", "4", "5", "2"},
				{"2", "3", "6", "4", "5", "2"},
				{"8", "9", "8"},
			},
		},
		{
			"it should identify 10 distinct cycles",
			map[string][]string{
				"0": {"1"},
				"1": {"4", "6", "7"},
				"2": {"4", "6", "7"},
				"3": {"4", "6", "7"},
				"4": {"2", "3"},
				"5": {"2", "3"},
				"6": {"5", "8"},
				"7": {"5", "8"},
				"8": {""},
				"9": {""},
			},
			[][]string{
				{"2", "4", "2"},
				{"2", "4", "3", "6", "5", "2"},
				{"2", "4", "3", "7", "5", "2"},
				{"2", "6", "5", "2"},
				{"2", "6", "5", "3", "4", "2"},
				{"2", "7", "5", "2"},
				{"2", "7", "5", "3", "4", "2"},
				{"3", "4", "3"},
				{"3", "6", "5", "3"},
				{"3", "7", "5", "3"},
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
