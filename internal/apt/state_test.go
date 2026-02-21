package apt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewState(t *testing.T) {
	state := NewState()

	if state.Version != StateVersion {
		t.Errorf("Expected version %d, got %d", StateVersion, state.Version)
	}
	if len(state.Packages) != 0 {
		t.Errorf("Expected empty packages, got %v", state.Packages)
	}
	if len(state.Repositories) != 0 {
		t.Errorf("Expected empty repositories, got %v", state.Repositories)
	}
	if len(state.Keys) != 0 {
		t.Errorf("Expected empty keys, got %v", state.Keys)
	}
}

func TestStatePackages(t *testing.T) {
	state := NewState()

	t.Run("add package", func(t *testing.T) {
		added := state.AddPackage("vim")
		if !added {
			t.Error("Expected package to be added")
		}
		if !state.HasPackage("vim") {
			t.Error("Expected HasPackage to return true")
		}
	})

	t.Run("add duplicate package", func(t *testing.T) {
		added := state.AddPackage("vim")
		if added {
			t.Error("Expected duplicate package not to be added")
		}
		if len(state.Packages) != 1 {
			t.Errorf("Expected 1 package, got %d", len(state.Packages))
		}
	})

	t.Run("add another package", func(t *testing.T) {
		added := state.AddPackage("curl")
		if !added {
			t.Error("Expected package to be added")
		}
		if len(state.Packages) != 2 {
			t.Errorf("Expected 2 packages, got %d", len(state.Packages))
		}
	})

	t.Run("remove package", func(t *testing.T) {
		removed := state.RemovePackage("vim")
		if !removed {
			t.Error("Expected package to be removed")
		}
		if state.HasPackage("vim") {
			t.Error("Expected HasPackage to return false")
		}
		if len(state.Packages) != 1 {
			t.Errorf("Expected 1 package, got %d", len(state.Packages))
		}
	})

	t.Run("remove nonexistent package", func(t *testing.T) {
		removed := state.RemovePackage("nonexistent")
		if removed {
			t.Error("Expected remove to return false for nonexistent package")
		}
	})
}

func TestStateRepositories(t *testing.T) {
	state := NewState()

	t.Run("add repository", func(t *testing.T) {
		added := state.AddRepository("docker.sources")
		if !added {
			t.Error("Expected repository to be added")
		}
		if !state.HasRepository("docker.sources") {
			t.Error("Expected HasRepository to return true")
		}
	})

	t.Run("add duplicate repository", func(t *testing.T) {
		added := state.AddRepository("docker.sources")
		if added {
			t.Error("Expected duplicate repository not to be added")
		}
	})

	t.Run("remove repository", func(t *testing.T) {
		removed := state.RemoveRepository("docker.sources")
		if !removed {
			t.Error("Expected repository to be removed")
		}
		if state.HasRepository("docker.sources") {
			t.Error("Expected HasRepository to return false")
		}
	})
}

func TestStateKeys(t *testing.T) {
	state := NewState()

	t.Run("add key", func(t *testing.T) {
		added := state.AddKey("docker.gpg")
		if !added {
			t.Error("Expected key to be added")
		}
		if !state.HasKey("docker.gpg") {
			t.Error("Expected HasKey to return true")
		}
	})

	t.Run("add duplicate key", func(t *testing.T) {
		added := state.AddKey("docker.gpg")
		if added {
			t.Error("Expected duplicate key not to be added")
		}
	})

	t.Run("remove key", func(t *testing.T) {
		removed := state.RemoveKey("docker.gpg")
		if !removed {
			t.Error("Expected key to be removed")
		}
		if state.HasKey("docker.gpg") {
			t.Error("Expected HasKey to return false")
		}
	})
}

func TestGetPackagesNotIn(t *testing.T) {
	state := NewState()
	state.AddPackage("vim")
	state.AddPackage("curl")
	state.AddPackage("git")
	state.AddPackage("htop")

	t.Run("some packages not in list", func(t *testing.T) {
		notIn := state.GetPackagesNotIn([]string{"vim", "git"})
		if len(notIn) != 2 {
			t.Errorf("Expected 2 packages not in list, got %d", len(notIn))
		}
		// Should contain curl and htop
		found := make(map[string]bool)
		for _, pkg := range notIn {
			found[pkg] = true
		}
		if !found["curl"] || !found["htop"] {
			t.Errorf("Expected curl and htop, got %v", notIn)
		}
	})

	t.Run("all packages in list", func(t *testing.T) {
		notIn := state.GetPackagesNotIn([]string{"vim", "curl", "git", "htop"})
		if len(notIn) != 0 {
			t.Errorf("Expected 0 packages not in list, got %d", len(notIn))
		}
	})

	t.Run("no packages in list", func(t *testing.T) {
		notIn := state.GetPackagesNotIn([]string{})
		if len(notIn) != 4 {
			t.Errorf("Expected 4 packages not in list, got %d", len(notIn))
		}
	})
}

func newTestManager(t *testing.T) *AptManager {
	t.Helper()
	tmpDir := t.TempDir()
	testStatePath := filepath.Join(tmpDir, "state.json")
	return &AptManager{
		StatePath: func() string { return testStatePath },
	}
}

func TestStatePersistence(t *testing.T) {
	m := newTestManager(t)
	testStatePath := m.StatePath()

	t.Run("save and load state", func(t *testing.T) {
		state := NewState()
		state.AddPackage("vim")
		state.AddPackage("curl")
		state.AddRepository("docker.sources")
		state.AddKey("docker.gpg")

		err := m.SaveState(state)
		if err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(testStatePath); os.IsNotExist(err) {
			t.Fatal("State file was not created")
		}

		// Load the state
		loaded, err := m.LoadState()
		if err != nil {
			t.Fatalf("Failed to load state: %v", err)
		}

		// Verify loaded state matches
		if loaded.Version != StateVersion {
			t.Errorf("Expected version %d, got %d", StateVersion, loaded.Version)
		}
		if len(loaded.Packages) != 2 {
			t.Errorf("Expected 2 packages, got %d", len(loaded.Packages))
		}
		if !loaded.HasPackage("vim") || !loaded.HasPackage("curl") {
			t.Error("Expected loaded state to have vim and curl")
		}
		if !loaded.HasRepository("docker.sources") {
			t.Error("Expected loaded state to have docker.sources repository")
		}
		if !loaded.HasKey("docker.gpg") {
			t.Error("Expected loaded state to have docker.gpg key")
		}
	})
}

func TestLoadStateNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	testStatePath := filepath.Join(tmpDir, "nonexistent", "state.json")
	m := &AptManager{
		StatePath: func() string { return testStatePath },
	}

	state, err := m.LoadState()
	if err != nil {
		t.Fatalf("Expected no error for nonexistent state file, got %v", err)
	}

	if state.Version != StateVersion {
		t.Errorf("Expected version %d, got %d", StateVersion, state.Version)
	}
	if len(state.Packages) != 0 {
		t.Errorf("Expected empty packages, got %v", state.Packages)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Use a nested path that doesn't exist
	testStatePath := filepath.Join(tmpDir, "nested", "dir", "state.json")
	m := &AptManager{
		StatePath: func() string { return testStatePath },
	}

	state := NewState()
	state.AddPackage("vim")

	err := m.SaveState(state)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testStatePath); os.IsNotExist(err) {
		t.Fatal("State file was not created in nested directory")
	}
}
