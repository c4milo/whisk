package scc

import (
	"fmt"
	"sort"
)

// Tarjan implements https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm
type Tarjan struct {
	// G is the graph's adjencency list
	G map[string][]string
	// index is the unique index assigned to each vertex, and used to identify the root of the strongly connected subgraph.
	index map[string]int
	// lowlink is Tarjan's mechanism to identify unique strongly connected subgraphs.
	lowlink map[int]int
	// stack is where nodes are stored until a unique strongly connected subgraph is identified.
	stack []string
	// onStack whether or not a vertex is on the stack.
	onStack map[int]bool
	// currentIndex is the latest index assigned to a vertex.
	currentIndex int
	// sccs holds the subgraphs of unique strongly connected components found.
	sccs [][]string
}

// NewTarjan initializes and returns a Tarjan's algorithm instance for finding strongly
// connected components in a graph. It takes linear time in the number of vertices and edges: O(V+E)
func NewTarjan(g map[string][]string) *Tarjan {
	return &Tarjan{
		G:       g,
		index:   make(map[string]int),
		lowlink: make(map[int]int),
		onStack: make(map[int]bool),
	}
}

// Find identifies and returns strongly connected components.
// Strongly connected components are unique and have the following properties:
//
// - Reflexive: There is a trivial path of length zero from any vertex to itself.
// - Symmetric: If there is a path from u to v, the same edges form a path from v to u.
// - Transitive: If there is a path from u to v and a path from v to w, the two paths may be concatenated together to form a path from u to w.
func (t *Tarjan) Find() ([][]string, error) {
	if t.G == nil {
		return nil, fmt.Errorf("no graph found")
	}

	t.currentIndex = 1

	for _, v := range sortKeys(t.G) {
		if t.index[v] == 0 {
			t.search(v)
		}
	}

	return t.sccs, nil
}

// search recursively walks the graph finding unique strongly connected components.
func (t *Tarjan) search(v string) {
	t.currentIndex++
	t.index[v] = t.currentIndex
	t.lowlink[t.index[v]] = t.currentIndex
	t.stack = append(t.stack, v)
	t.onStack[t.index[v]] = true

	for _, w := range t.G[v] {
		if _, ok := t.index[w]; !ok {
			t.search(w)
			t.lowlink[t.index[v]] = min(t.lowlink[t.index[v]], t.lowlink[t.index[w]])
		} else if t.index[w] < t.index[v] {
			if t.onStack[t.index[w]] {
				t.lowlink[t.index[v]] = min(t.lowlink[t.index[v]], t.index[w])
			}
		}
	}

	if t.index[v] == t.lowlink[t.index[v]] {
		var vertices []string
		i := len(t.stack) - 1
		for {
			u := t.stack[i]
			t.onStack[t.index[u]] = false
			t.lowlink[t.index[u]] = t.index[v]

			vertices = append(vertices, u)
			if u == v {
				break
			}
			i--
		}
		t.stack = t.stack[:i] // pop v
		t.sccs = append(t.sccs, vertices)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
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
