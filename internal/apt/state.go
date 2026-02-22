package apt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

const (
	// StateDir is the directory where apt-bundle stores its state
	StateDir = "/var/lib/apt-bundle"
	// StateFile is the filename for the state file
	StateFile = "state.json"
	// StateVersion is the current version of the state file format
	StateVersion = 1
)

// State represents the apt-bundle managed state
type State struct {
	Version      int      `json:"version"`
	Packages     []string `json:"packages"`
	Repositories []string `json:"repositories"`
	Keys         []string `json:"keys"`
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Version:      StateVersion,
		Packages:     []string{},
		Repositories: []string{},
		Keys:         []string{},
	}
}

// LoadState loads the state from disk, or returns a new state if none exists
func (m *AptManager) LoadState() (*State, error) {
	path := m.StatePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewState(), nil
		}
		return nil, fmt.Errorf("load state from %s: %w", path, err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("load state from %s: %w", path, err)
	}

	return &state, nil
}

// SaveState persists the state to disk
func (m *AptManager) SaveState(s *State) error {
	path := m.StatePath()

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write state file %s: %w", path, err)
	}
	return nil
}

// AddPackage adds a package to the state if not already present
func (s *State) AddPackage(pkg string) bool {
	if slices.Contains(s.Packages, pkg) {
		return false
	}
	s.Packages = append(s.Packages, pkg)
	return true
}

// RemovePackage removes a package from the state
func (s *State) RemovePackage(pkg string) bool {
	idx := slices.Index(s.Packages, pkg)
	if idx == -1 {
		return false
	}
	s.Packages = slices.Delete(s.Packages, idx, idx+1)
	return true
}

// HasPackage checks if a package is tracked in the state
func (s *State) HasPackage(pkg string) bool {
	return slices.Contains(s.Packages, pkg)
}

// AddRepository adds a repository to the state if not already present
func (s *State) AddRepository(repo string) bool {
	if slices.Contains(s.Repositories, repo) {
		return false
	}
	s.Repositories = append(s.Repositories, repo)
	return true
}

// RemoveRepository removes a repository from the state
func (s *State) RemoveRepository(repo string) bool {
	idx := slices.Index(s.Repositories, repo)
	if idx == -1 {
		return false
	}
	s.Repositories = slices.Delete(s.Repositories, idx, idx+1)
	return true
}

// HasRepository checks if a repository is tracked in the state
func (s *State) HasRepository(repo string) bool {
	return slices.Contains(s.Repositories, repo)
}

// AddKey adds a key to the state if not already present
func (s *State) AddKey(key string) bool {
	if slices.Contains(s.Keys, key) {
		return false
	}
	s.Keys = append(s.Keys, key)
	return true
}

// RemoveKey removes a key from the state
func (s *State) RemoveKey(key string) bool {
	idx := slices.Index(s.Keys, key)
	if idx == -1 {
		return false
	}
	s.Keys = slices.Delete(s.Keys, idx, idx+1)
	return true
}

// HasKey checks if a key is tracked in the state
func (s *State) HasKey(key string) bool {
	return slices.Contains(s.Keys, key)
}

// GetPackagesNotIn returns packages in state that are not in the given list
func (s *State) GetPackagesNotIn(packages []string) []string {
	var result []string
	for _, pkg := range s.Packages {
		if !slices.Contains(packages, pkg) {
			result = append(result, pkg)
		}
	}
	return result
}
