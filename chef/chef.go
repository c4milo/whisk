package chef

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Role contains the Chef role run list of recipes that Chefs executes in order.
type Role struct {
	// Name is the name given to the role
	Name string `json:"name"`
	// RunList is the list of roles and/or recipes Chef will run in order.
	RunList []string `json:"run_list"`
}

// NewRole opens and decodes a role file.
func NewRole(path string) (*Role, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed opening role file: %w", err)
	}

	role := new(Role)
	if err := json.Unmarshal(data, role); err != nil {
		return nil, fmt.Errorf("failed decoding role file %q: %w", path, err)
	}

	return role, nil
}

// Cookbook is a Chef Cookbook.
type Cookbook struct {
	CookbookPaths []string
	Name          string            `json:"name"`
	Deps          map[string]string `json:"dependencies"`
}

// LoadDeps loads the cookbook's dependencies, trying first from its metadata.rb,
// if it exists, or its metadata.json file, otherwise.
func (c *Cookbook) LoadDeps() error {
	if c.Name == "" {
		return fmt.Errorf("cookbook name can't be empty")
	}

	if err := c.tryRuby(); err == nil {
		return nil
	}

	return c.tryJSON()
}

// tryJSON reads the cookbook's dependencies for a metadata.json file.
func (c *Cookbook) tryJSON() error {
	metadataPath := filepath.Join(c.Name, "metadata.json")
	var (
		metadata []byte
		err      error
	)
	for _, p := range c.CookbookPaths {
		metadata, err = ioutil.ReadFile(filepath.Join(p, metadataPath))
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("could not find cookbook metadata %q in %q: %w", metadataPath, c.CookbookPaths, err)
	}

	if err := json.Unmarshal(metadata, c); err != nil {
		return fmt.Errorf("failed decoding %q: %w", metadataPath, err)
	}

	return nil
}

// tryRuby reads the cookbook's dependencies from a metadata.rb file.
func (c *Cookbook) tryRuby() error {
	metadataPath := filepath.Join(c.Name, "metadata.rb")
	var (
		metadata []byte
		err      error
	)
	for _, p := range c.CookbookPaths {
		metadata, err = ioutil.ReadFile(filepath.Join(p, metadataPath))
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("could not find cookbook metadata %q in %q: %w", metadataPath, c.CookbookPaths, err)
	}

	f := bytes.NewReader(metadata)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "depends") {
			parts := strings.Fields(line)
			cookbook := strings.Trim(parts[1], `"',`)
			c.Deps[cookbook] = "" // TODO(c4milo): track version if it makes sense.
		}
	}

	return nil
}
