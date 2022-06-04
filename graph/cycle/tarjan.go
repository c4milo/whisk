package cycle

import (
	"fmt"
	"sort"
)

// Tarjan implements Enumeration of the Elementary Circuits of a Directed Graph:
// https://ecommons.cornell.edu/bitstream/handle/1813/5941/72-145.pdf
type Tarjan struct {
	// G is the source graph.
	G map[string][]string
	// index are the unique identifiers of the vertices.
	index map[string]int
	// marked determines whether a vertex has been walked or not during a single path exploration rooted in s.
	marked map[int]bool
	// removed keeps track of vertices that have been processed for cycles already, to avoid duplicates.
	removed map[int]map[int]bool
	// markedStack keeps track of the marked vertices.
	markedStack []int
	// pointStack keeps track of the current path being explored, cycles are taken from here.
	pointStack []string
	// cycles is the list of distinct cycles found.
	cycles [][]string
}

// NewTarjan initializes and returns a Tarjan algorithm instance for enumerating
// distinct cycles in a graph.
func NewTarjan(g map[string][]string) *Tarjan {
	return &Tarjan{
		G:       g,
		marked:  make(map[int]bool),
		removed: make(map[int]map[int]bool),
		index:   make(map[string]int),
	}
}

// Find enumerates and returns the list of distinct cycles in the graph.
func (t *Tarjan) Find() ([][]string, error) {
	if t.G == nil {
		return nil, fmt.Errorf("no graph found")
	}

	vertices := sortKeys(t.G)
	for i, v := range vertices {
		i++ // makes debugging easier
		t.index[v] = i
	}

	for _, start := range vertices {
		t.search(start, start)

		for _, u := range t.markedStack {
			t.marked[u] = false
		}
		t.markedStack = nil
	}

	return t.cycles, nil
}

// search recursively walks individual paths in the graph, looking for distinct cycles.
func (t *Tarjan) search(s, v string) bool {
	found := false
	t.marked[t.index[v]] = true
	t.pointStack = append(t.pointStack, v)
	t.markedStack = append(t.markedStack, t.index[v])

	for _, w := range t.G[v] {
		// Skip finding cyclic permutations of the same cycle.
		if _, ok := t.removed[t.index[v]][t.index[w]]; ok {
			continue
		}

		switch {
		case t.index[w] < t.index[s]:
			// edge w was previously explored, add it to the removed list of v, to skip it,
			// as we keep recursing down the graph. This is to avoid finding duplicated cycles.
			if _, ok := t.removed[t.index[v]]; !ok {
				t.removed[t.index[v]] = make(map[int]bool)
			}
			t.removed[t.index[v]][t.index[w]] = true
		case t.index[w] == t.index[s]:
			// A cycle has been found.
			newCycle := make([]string, len(t.pointStack))
			copy(newCycle, t.pointStack)
			newCycle = append(newCycle, s)
			t.cycles = append(t.cycles, newCycle)
			found = true
		case !t.marked[t.index[w]]:
			// We haven't seen this vertex yet for s' current path; so, we keep walking the graph.
			g := t.search(s, w)
			found = found || g
		}
	}

	// If a cycle is found we need to unmark visited vertices so we can walk them again
	// while searching for the next cycle rooted in s.
	if found {
		j := len(t.markedStack) - 1
		for {
			u := t.markedStack[j]
			t.markedStack = t.markedStack[:j] // pops
			t.marked[u] = false
			if u == t.index[v] {
				break
			}
			j--
		}
	}

	// pops v from the pointStack since we are done exploring paths in its edges.
	t.pointStack = t.pointStack[:len(t.pointStack)-1]

	return found
}

// sortKeys sorts the map's keys alphabetically.
func sortKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
