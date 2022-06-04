package whisk

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"slack/whisk/chef"
	"slack/whisk/graph/cycle"
	"slack/whisk/graph/scc"

	"github.com/xlab/treeprint"
)

// graphviz dot template
const dotTpl = `
digraph g {
	bgcolor = "#ffffff"
	splines = ortho
	overlap = true
	newrank = true

	node [
		shape = rectangle,
		width = 0.25,
		color = "#323538",
		fillcolor = white,
		style = "filled, solid",
		fontcolor = "#323538",
		fontsize = 8,
	]

	edge [
		penwidth = 0.50,
		color = "#323538",
		arrowhead = "vee"
	]
	{{ range $index, $value := .Sccs }}
	subgraph cluster_sccs{{ $index }} {
		style = "filled, solid";
		color = "#F2C744";
		label = "Strongly Connected Subgraph {{ $index }}";

		{{ range $value -}}
		{{ replace . "-" "_" }} [color = "#F2C744"];
		{{ end }}
	}
	{{ end -}}

	{{- range $vertex, $edges := .G -}}
	{{ replace $vertex "-" "_" }} -> { {{ range $edges }} {{ replace . "-" "_" }} {{ end }} }
	{{ end }}
}
`

// Handler is, guess what, the whisk's handler! It loads Chef's dependency graph and
// finds strongly connected components as well as distinct cycles. Offering multiple
// output formats to display the information.
type Handler struct {
	// cookbookPaths is the list of directories where cookbooks are kept.
	cookbookPaths []string
	// rolesPath is the directory where the roles are stored.
	rolesPath string
	// rolesIndex keeps the entire roles population in memory for lookups.
	// This is because role names don't necessarily match their file names.
	rolesIndex map[string]*chef.Role
	// sccs holds the subgraphs of unique strongly connected components found.
	sccs [][]string
	// graph contains the unmodified directed graph, as found in roles and cookbooks.
	graph map[string][]string
	// cycles contains the distinct cycles found in the dependency graph.
	cycles [][]string
}

// NewHandler creates a new whisk handler instance.
func NewHandler(cookbooks []string, rolesPath string) *Handler {
	return &Handler{
		cookbookPaths: cookbooks,
		rolesPath:     rolesPath,
		graph:         make(map[string][]string),
		rolesIndex:    make(map[string]*chef.Role),
	}
}

// loadRoles preemptively loads and decodes role files. This is
// so we can do role name lookup since roles names and their file names
// don't have to match.
func (h *Handler) loadRoles() error {
	fn := func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(path, ".json") && h.rolesPath != path {
			return fs.SkipDir
		}

		// If there was any error stat()ing path, return it.
		if err != nil {
			return err
		}

		// Do not attempt to index the roles base dir.
		if path == h.rolesPath {
			return nil
		}

		role, err := chef.NewRole(path)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		h.rolesIndex[role.Name] = role

		return nil
	}

	if err := filepath.WalkDir(h.rolesPath, fn); err != nil {
		return fmt.Errorf("failed loading roles: %w", err)
	}

	return nil
}

// WalkRole traverses a role's run list using depth-first search and load Chef's
// dependency graph into memory to work with it.
func (h *Handler) WalkRole(name string, tree treeprint.Tree) error {
	if len(h.rolesIndex) == 0 {
		if err := h.loadRoles(); err != nil {
			return err
		}
	}

	role, ok := h.rolesIndex[name]
	if !ok {
		return fmt.Errorf("role %s doesn't exist", name)
	}

	for _, dep := range role.RunList {
		switch {
		case strings.HasPrefix(dep, "role[") && strings.HasSuffix(dep, "]"):
			// making role[foobar] into foobar.json
			name := dep[5 : len(dep)-1]
			metaName := fmt.Sprintf("%s.json", name)

			if err := h.WalkRole(name, tree.AddBranch(metaName)); err != nil {
				return fmt.Errorf("failed walking run_list: %w", err)
			}

		case strings.HasPrefix(dep, "recipe[") && strings.HasSuffix(dep, "]"):
			recipe := dep[7 : len(dep)-1]
			cookbook := strings.Split(recipe, "::")[0]

			if err := h.walkCookbook(cookbook, tree.AddBranch(cookbook)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid entry in role's run_list: %q", dep)
		}
	}

	return nil
}

// sortKeys sorts the map's keys alphabetically.
func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// walkCookbook recursively walks the cookbook's dependencies and loads
// them into a graph adjency list.
func (h *Handler) walkCookbook(name string, tree treeprint.Tree) error {
	// initializes the vertex to not miss it, in case it has no neihgbors.
	h.graph[name] = []string{}

	// Initialize struct to decode cookbooks metadata into.
	cookbook := &chef.Cookbook{
		CookbookPaths: h.cookbookPaths,
		Name:          name,
		Deps:          make(map[string]string),
	}

	if err := cookbook.LoadDeps(); err != nil {
		return fmt.Errorf("unable to load %q dependencies: %w", name, err)
	}

	for _, dep := range sortKeys(cookbook.Deps) {
		h.graph[name] = append(h.graph[name], dep)

		// if we haven't seen `dep`, we walk its dependencies as well (DFS).
		if _, ok := h.graph[dep]; !ok {
			if err := h.walkCookbook(dep, tree.AddBranch(dep)); err != nil {
				return err
			}
		}
	}

	return nil
}

// FindSCCs finds strongly connected components in the dependency graph.
func (h *Handler) FindSCCs() error {
	t := scc.NewTarjan(h.graph)

	sccs, err := t.Find()
	if err != nil {
		return fmt.Errorf("failed finding strongly connected components: %w", err)
	}

	// We are only interested in strongly connected subgraphs greater than 1
	for _, c := range sccs {
		if len(c) > 1 {
			h.sccs = append(h.sccs, c)
		}
	}

	return nil
}

// FindCycles finds distinct cycles in the graph.
func (h *Handler) FindCycles() error {
	t := cycle.NewTarjan(h.graph)

	cycles, err := t.Find()
	if err != nil {
		return fmt.Errorf("failed finding cycles: %w", err)
	}
	h.cycles = cycles

	return nil
}

// Result defines the struct to return back to callers using the different output formats.
type Result struct {
	// G is the digraph of the role
	G map[string][]string `json:"g"`
	// Sccs are the strongly connected cookbooks found.
	Sccs [][]string `json:"sccs"`
	// Cycles contains all the distinct cycles found in the digraph.
	Cycles [][]string `json:"cycles"`
}

// Result returns the dependency analysis results.
func (h *Handler) Result() Result {
	return Result{
		G:      h.graph,
		Sccs:   h.sccs,
		Cycles: h.cycles,
	}
}

// ASCII encodes the dependency graph using unicode ¯\_(ツ)_/¯
func (h *Handler) ASCII(tree treeprint.Tree, w io.Writer) {
	fmt.Fprintf(w, "%s\n", tree.String())

	totalSCC := len(h.sccs)
	fmt.Fprintf(w, "\n⚠️  Strongly Connected Components (topologically sorted): %d\n\n", totalSCC)
	if totalSCC == 0 {
		fmt.Fprintf(w, "None! 🍻 🎉 \n\n")
	}

	for i, c := range h.sccs {
		i++
		scc := strings.Join(c, ", ")
		fmt.Fprintf(w, "%d. %s\n", i, scc)
	}

	totalCycles := len(h.cycles)
	fmt.Fprintf(w, "\n\n🌀 Cycles: %d\n\n", totalCycles)
	if totalCycles == 0 {
		fmt.Fprintf(w, "None! 🍻 🎉 \n\n")
	}

	for i, c := range h.cycles {
		i++
		scc := strings.Join(c, ", ")
		fmt.Fprintf(w, "%d. %s\n", i, scc)
	}
}

// dotOutput encodes the dependency graph to graphviz's dot format.
func (h *Handler) DOT(w io.Writer) error {
	funcMap := template.FuncMap{
		// replace is used to make node names valid Dot identifiers.
		"replace": strings.ReplaceAll,
	}

	tpl, err := template.New("graph").Funcs(funcMap).Parse(dotTpl)
	if err != nil {
		return fmt.Errorf("failed parsing dot template: %w", err)
	}

	if err := tpl.Execute(w, h.Result()); err != nil {
		return fmt.Errorf("failed executing dot template: %w", err)
	}

	return nil
}

// jsonOutput encodes the dependency graph to JSON.
func (h *Handler) JSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(h.Result())
}
