package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"slack/whisk"
	"slack/whisk/chef"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
	"golang.org/x/sync/errgroup"
)

const statsTmpl = `
Closest Matches:
{{ range $metric, $stat := . }}
  {{ $metric }}: {{ $stat.Max }}
  role: {{ $stat.Role }}
  found: {{ $stat.Value }}
{{ end }}
`

var lintCmd = &cobra.Command{
	Use:   "lint [flags] <roles_dir>",
	Short: "Lints all Chef roles dependencies to make sure a minimum quality bar is held",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("chef roles base directory required")
		}

		return nil
	},
	RunE: lint,
}

// Command line flags for linting rules supported.
var (
	maxCycles          uint
	maxSCCs            uint
	maxCookbooksPerSCC uint
)

// init Initializes command line flags supported.
func init() {
	flagSet := lintCmd.Flags()
	flagSet.UintVar(&maxCycles, "max-cycles", 0, "maximum number of distinct circular dependencies accepted")
	flagSet.UintVar(&maxSCCs, "max-sccs", 0, "maximum number of unique strongly connected components")
	flagSet.UintVar(&maxCookbooksPerSCC, "max-cookbooks-per-scc", 0, "maximum number of cookbooks per strongly connected component")
}

// closestMatch is used to give people context on successful linting results, in case they are using
// the lint comnand and CI output as feedback loop to validate their fixes.
type closestMatch struct {
	// role is the role's name.
	Role string
	// metric is the name of the metric.
	Metric string
	// value is the score of the role on the specific metric.
	Value int
	// max is the maximum value allowed by the involved linting rule.
	Max int
}

// linter defines a simple linter for Chef roles and cookbooks.
type linter struct {
	cookbookPath   string
	rolesDir       string
	eg             *errgroup.Group
	roles          int
	closestMatches map[string]*closestMatch

	// rules
	maxCycles          uint
	maxSCCs            uint
	maxCookbooksPerSCC uint
}

// lint is a Cobra function handler for the lint subcommand.
func lint(cmd *cobra.Command, args []string) error {
	rolesDir := args[0]

	l := &linter{
		cookbookPath: cookbookPath, // persistent flag defined in root.go
		rolesDir:     rolesDir,
		eg:           new(errgroup.Group),
		roles:        0,
		closestMatches: map[string]*closestMatch{
			"max-cycles": {
				Metric: "max-cycles",
				Max:    int(maxCycles),
			},
			"max-sccs": {
				Metric: "max-sccs",
				Max:    int(maxSCCs),
			},
			"max-cookbooks-per-scc": {
				Metric: "max-cookbooks-per-scc",
				Max:    int(maxCookbooksPerSCC),
			},
		},

		// rules
		maxCycles:          maxCycles,
		maxSCCs:            maxSCCs,
		maxCookbooksPerSCC: maxCookbooksPerSCC,
	}

	if err := l.lintRoles(); err != nil {
		return fmt.Errorf("linting errors were found. \n\n %w", err)
	}

	fmt.Fprintf(os.Stderr, "No threshold was reached! 🍻 \n")

	t := template.Must(template.New("stats").Parse(statsTmpl))
	if err := t.Execute(os.Stderr, l.closestMatches); err != nil {
		return fmt.Errorf("failed rendering stats: %w", err)
	}

	return nil
}

// lintRoles walks Chef's roles directory and analyzes the digraph of every role found, returning
// on the first role not meeting linting thresholds.
func (l *linter) lintRoles() error {
	if err := filepath.WalkDir(l.rolesDir, l.walkDirFunc); err != nil {
		return fmt.Errorf("failed walking %q dir: %w", l.rolesDir, err)
	}

	fmt.Fprintf(os.Stderr, "\nLinting %d Chef roles...\n\n", l.roles)

	return l.eg.Wait()
}

// walkDirFunc runs for each role JSON file found and launches a goroutine
// to do dependency analysis.
func (l *linter) walkDirFunc(path string, d fs.DirEntry, err error) error {
	if !strings.HasSuffix(path, ".json") && l.rolesDir != path {
		return fs.SkipDir
	}

	// If there was any error stat()ing path, return it.
	if err != nil {
		return err
	}

	// Do not attempt to lint the roles base dir.
	if path == l.rolesDir {
		return nil
	}

	l.roles++
	l.eg.Go(func() error {
		return l.lint(path)
	})

	return nil
}

// lint runs linting for a single role.
func (l *linter) lint(rolePath string) error {
	handler := whisk.NewHandler(strings.Split(l.cookbookPath, ","), l.rolesDir)

	role, err := chef.NewRole(rolePath)
	if err != nil {
		return fmt.Errorf("failed loading role %q: %w", rolePath, err)
	}

	if err := handler.WalkRole(role.Name, treeprint.New()); err != nil {
		return fmt.Errorf("%s: %w", role.Name, err)
	}

	if err := handler.FindSCCs(); err != nil {
		return fmt.Errorf("%s: failed to find strongly connected components: %w", role.Name, err)
	}

	if err := handler.FindCycles(); err != nil {
		return fmt.Errorf("%s: failed to enumerate distinct cyles: %w", role.Name, err)
	}

	r := handler.Result()
	var lr *multierror.Error

	cyclesFound := len(r.Cycles)
	if cyclesFound > int(l.maxCycles) {
		lr = multierror.Append(lr, fmt.Errorf("%s: %d cycles found. Max threshold: %d", role.Name, cyclesFound, l.maxCycles))
	}

	sccsFound := len(r.Sccs)
	if sccsFound > int(l.maxSCCs) {
		lr = multierror.Append(lr, fmt.Errorf("%s: %d sccs found. Max threshold: %d", role.Name, sccsFound, l.maxSCCs))
	}

	for i, scc := range r.Sccs {
		cookbooksFound := len(scc)
		if cookbooksFound > int(l.maxCookbooksPerSCC) {
			lr = multierror.Append(lr, fmt.Errorf("%s: %d cookbooks found in scc %d. Max threshold: %d", role.Name, cookbooksFound, i, l.maxCookbooksPerSCC))
		}

		if m := l.closestMatches["max-cookbooks-per-scc"]; m.Value < cookbooksFound {
			m.Value = cookbooksFound
			m.Role = role.Name
		}
	}

	if m := l.closestMatches["max-cycles"]; m.Value < cyclesFound {
		m.Value = cyclesFound
		m.Role = role.Name
	}

	if m := l.closestMatches["max-sccs"]; m.Value < sccsFound {
		m.Value = sccsFound
		m.Role = role.Name
	}

	return lr.ErrorOrNil()
}
