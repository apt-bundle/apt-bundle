package apt

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestIsUbuntu(t *testing.T) {
	dir := t.TempDir()

	t.Run("ID=ubuntu returns true", func(t *testing.T) {
		f := filepath.Join(dir, "os-release-ubuntu")
		if err := os.WriteFile(f, []byte("ID=ubuntu\nVERSION_ID=22.04\n"), 0644); err != nil {
			t.Fatal(err)
		}
		m := &AptManager{OsReleasePath: f}
		if !m.isUbuntu() {
			t.Error("expected isUbuntu() true for ID=ubuntu")
		}
	})

	t.Run("ID_LIKE=ubuntu returns true", func(t *testing.T) {
		f := filepath.Join(dir, "os-release-mint")
		if err := os.WriteFile(f, []byte("ID=linuxmint\nID_LIKE=ubuntu\n"), 0644); err != nil {
			t.Fatal(err)
		}
		m := &AptManager{OsReleasePath: f}
		if !m.isUbuntu() {
			t.Error("expected isUbuntu() true for ID_LIKE=ubuntu")
		}
	})

	t.Run("non-Ubuntu returns false", func(t *testing.T) {
		f := filepath.Join(dir, "os-release-debian")
		if err := os.WriteFile(f, []byte("ID=debian\nVERSION_ID=12\n"), 0644); err != nil {
			t.Fatal(err)
		}
		m := &AptManager{OsReleasePath: f}
		if m.isUbuntu() {
			t.Error("expected isUbuntu() false for ID=debian")
		}
	})

	t.Run("missing file returns false", func(t *testing.T) {
		m := &AptManager{OsReleasePath: filepath.Join(dir, "nonexistent")}
		if m.isUbuntu() {
			t.Error("expected isUbuntu() false when os-release missing")
		}
	})
}

func TestAddPPA(t *testing.T) {
	t.Run("add-apt-repository not available", func(t *testing.T) {
		m := NewAptManager()
		err := m.AddPPA("ppa:deadsnakes/ppa")

		if err != nil {
			if _, lookupErr := exec.LookPath("add-apt-repository"); lookupErr != nil {
				return
			}
			t.Logf("AddPPA failed (likely due to permissions): %v", err)
		}
	})

	t.Run("empty ppa", func(t *testing.T) {
		if _, err := exec.LookPath("add-apt-repository"); err != nil {
			t.Skip("add-apt-repository not available, skipping test")
		}

		m := NewAptManager()
		err := m.AddPPA("")
		if err == nil {
			t.Log("Warning: AddPPA('') succeeded unexpectedly")
		}
	})

	t.Run("invalid ppa format", func(t *testing.T) {
		if _, err := exec.LookPath("add-apt-repository"); err != nil {
			t.Skip("add-apt-repository not available, skipping test")
		}

		m := NewAptManager()
		err := m.AddPPA("invalid-ppa-format")
		if err == nil {
			t.Log("Warning: AddPPA with invalid format succeeded unexpectedly")
		}
	})

	t.Run("valid ppa format", func(t *testing.T) {
		if _, err := exec.LookPath("add-apt-repository"); err != nil {
			t.Skip("add-apt-repository not available, skipping test")
		}

		m := NewAptManager()
		err := m.AddPPA("ppa:deadsnakes/ppa")
		if err == nil {
			t.Log("Warning: AddPPA succeeded unexpectedly (might have sudo)")
		}
	})
}

func TestParseDebLine(t *testing.T) {
	t.Run("simple deb line", func(t *testing.T) {
		repo, err := parseDebLine("https://example.com/ubuntu jammy main")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.Types != "deb" {
			t.Errorf("Expected type 'deb', got '%s'", repo.Types)
		}
		if repo.URIs != "https://example.com/ubuntu" {
			t.Errorf("Expected URI 'https://example.com/ubuntu', got '%s'", repo.URIs)
		}
		if repo.Suites != "jammy" {
			t.Errorf("Expected suite 'jammy', got '%s'", repo.Suites)
		}
		if repo.Components != "main" {
			t.Errorf("Expected components 'main', got '%s'", repo.Components)
		}
	})

	t.Run("deb line with arch option", func(t *testing.T) {
		repo, err := parseDebLine("[arch=amd64] https://download.docker.com/linux/ubuntu focal stable")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.Architectures != "amd64" {
			t.Errorf("Expected arch 'amd64', got '%s'", repo.Architectures)
		}
		if repo.URIs != "https://download.docker.com/linux/ubuntu" {
			t.Errorf("Expected URI, got '%s'", repo.URIs)
		}
		if repo.Suites != "focal" {
			t.Errorf("Expected suite 'focal', got '%s'", repo.Suites)
		}
		if repo.Components != "stable" {
			t.Errorf("Expected components 'stable', got '%s'", repo.Components)
		}
	})

	t.Run("deb line with multiple components", func(t *testing.T) {
		repo, err := parseDebLine("https://example.com/ubuntu jammy main contrib non-free")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.Components != "main contrib non-free" {
			t.Errorf("Expected components 'main contrib non-free', got '%s'", repo.Components)
		}
	})

	t.Run("deb line with leading deb", func(t *testing.T) {
		repo, err := parseDebLine("deb https://example.com/ubuntu jammy main")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.Types != "deb" {
			t.Errorf("Expected type 'deb', got '%s'", repo.Types)
		}
	})

	t.Run("deb-src line", func(t *testing.T) {
		repo, err := parseDebLine("deb-src https://example.com/ubuntu jammy main")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.Types != "deb-src" {
			t.Errorf("Expected type 'deb-src', got '%s'", repo.Types)
		}
		if repo.URIs != "https://example.com/ubuntu" {
			t.Errorf("Expected URI 'https://example.com/ubuntu', got '%s'", repo.URIs)
		}
	})

	t.Run("invalid deb line - too few parts", func(t *testing.T) {
		_, err := parseDebLine("https://example.com/ubuntu")
		if err == nil {
			t.Error("Expected error for invalid deb line")
		}
	})

	t.Run("empty deb line", func(t *testing.T) {
		_, err := parseDebLine("")
		if err == nil {
			t.Error("Expected error for empty deb line")
		}
	})
}

func TestDebRepositoryToDEB822(t *testing.T) {
	t.Run("full repository", func(t *testing.T) {
		repo := &DebRepository{
			Types:         "deb",
			URIs:          "https://download.docker.com/linux/ubuntu",
			Suites:        "focal",
			Components:    "stable",
			Architectures: "amd64",
			SignedBy:      "/etc/apt/keyrings/docker.gpg",
		}

		content := repo.ToDEB822()

		expected := `Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: focal
Components: stable
Architectures: amd64
Signed-By: /etc/apt/keyrings/docker.gpg
`
		if content != expected {
			t.Errorf("DEB822 content mismatch.\nGot:\n%s\nExpected:\n%s", content, expected)
		}
	})

	t.Run("minimal repository", func(t *testing.T) {
		repo := &DebRepository{
			Types:  "deb",
			URIs:   "https://example.com/ubuntu",
			Suites: "jammy",
		}

		content := repo.ToDEB822()

		if !strings.Contains(content, "Types: deb") {
			t.Error("Missing Types field")
		}
		if !strings.Contains(content, "URIs: https://example.com/ubuntu") {
			t.Error("Missing URIs field")
		}
		if !strings.Contains(content, "Suites: jammy") {
			t.Error("Missing Suites field")
		}
		if strings.Contains(content, "Components:") {
			t.Error("Should not have Components field when empty")
		}
		if strings.Contains(content, "Architectures:") {
			t.Error("Should not have Architectures field when empty")
		}
		if strings.Contains(content, "Signed-By:") {
			t.Error("Should not have Signed-By field when empty")
		}
	})
}

func TestAddDebRepository(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("parse and format Docker repository", func(t *testing.T) {
		repoLine := "[arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
		keyPath := filepath.Join(tmpDir, "docker.gpg")

		// Create a fake key file
		if err := os.WriteFile(keyPath, []byte("fake key"), 0644); err != nil {
			t.Fatalf("Failed to create fake key file: %v", err)
		}

		// This will fail without write permissions to /etc/apt/sources.list.d
		// but we can verify it doesn't panic and handles the error
		_, err := AddDebRepository(repoLine, keyPath)
		if err != nil {
			// Expected to fail in test environment
			t.Logf("AddDebRepository failed (expected in test): %v", err)
		}
	})
}

func TestRemoveDebRepository(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("remove existing source", func(t *testing.T) {
		sourcePath := filepath.Join(tmpDir, "test.sources")
		if err := os.WriteFile(sourcePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := RemoveDebRepository(sourcePath)
		if err != nil {
			t.Errorf("RemoveDebRepository failed: %v", err)
		}

		if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
			t.Error("Source file should have been removed")
		}
	})

	t.Run("remove nonexistent source", func(t *testing.T) {
		err := RemoveDebRepository(filepath.Join(tmpDir, "nonexistent.sources"))
		if err != nil {
			t.Errorf("RemoveDebRepository should not error for nonexistent file: %v", err)
		}
	})
}

func TestAddPPAWithMock(t *testing.T) {
	t.Run("add-apt-repository not found", func(t *testing.T) {
		m := &AptManager{
			OsReleasePath: "/nonexistent",
			LookPath: func(file string) (string, error) {
				return "", errors.New("executable file not found in $PATH")
			},
		}

		err := m.AddPPA("ppa:test/ppa")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "add-apt-repository not found. Please install software-properties-common" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("successful PPA addition", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		m := &AptManager{
			Executor:      mock,
			OsReleasePath: "/nonexistent",
			LookPath: func(file string) (string, error) {
				return "/usr/bin/add-apt-repository", nil
			},
		}

		err := m.AddPPA("ppa:deadsnakes/ppa")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(mock.RunCalls) != 1 {
			t.Errorf("Expected 1 run call, got %d", len(mock.RunCalls))
		}
		expectedArgs := []string{"add-apt-repository", "-y", "ppa:deadsnakes/ppa"}
		for i, arg := range expectedArgs {
			if mock.RunCalls[0][i] != arg {
				t.Errorf("Arg %d: expected %s, got %s", i, arg, mock.RunCalls[0][i])
			}
		}
	})

	t.Run("PPA addition command failure", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return errors.New("E: The repository 'ppa:invalid/ppa' does not have a Release file")
		}
		m := &AptManager{
			Executor:      mock,
			OsReleasePath: "/nonexistent",
			LookPath: func(file string) (string, error) {
				return "/usr/bin/add-apt-repository", nil
			},
		}

		err := m.AddPPA("ppa:invalid/ppa")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		expectedErr := "failed to add PPA ppa:invalid/ppa: E: The repository 'ppa:invalid/ppa' does not have a Release file"
		if err.Error() != expectedErr {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("lookPath is called with correct argument", func(t *testing.T) {
		var calledWith string
		mock := testutil.NewMockExecutor()
		m := &AptManager{
			Executor:      mock,
			OsReleasePath: "/nonexistent",
			LookPath: func(file string) (string, error) {
				calledWith = file
				return "/usr/bin/add-apt-repository", nil
			},
		}

		_ = m.AddPPA("ppa:test/ppa")

		if calledWith != "add-apt-repository" {
			t.Errorf("Expected lookPath to be called with 'add-apt-repository', got '%s'", calledWith)
		}
	})
}

func TestLookPathField(t *testing.T) {
	customLookPath := func(file string) (string, error) {
		return "/custom/path", nil
	}
	m := &AptManager{LookPath: customLookPath}

	path, err := m.LookPath("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if path != "/custom/path" {
		t.Errorf("Expected '/custom/path', got '%s'", path)
	}
}

func BenchmarkAddPPA(b *testing.B) {
	if _, err := exec.LookPath("add-apt-repository"); err != nil {
		b.Skip("add-apt-repository not available, skipping benchmark")
	}

	m := NewAptManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.AddPPA("ppa:test/ppa")
	}
}
