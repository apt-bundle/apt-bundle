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

	t.Run("ID_LIKE with multiple values returns true", func(t *testing.T) {
		f := filepath.Join(dir, "os-release-pop")
		if err := os.WriteFile(f, []byte("ID=pop\nID_LIKE=\"ubuntu debian\"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		m := &AptManager{OsReleasePath: f}
		if !m.isUbuntu() {
			t.Error("expected isUbuntu() true for ID_LIKE containing ubuntu")
		}
	})

	t.Run("PRETTY_NAME containing ubuntu does not false-match", func(t *testing.T) {
		f := filepath.Join(dir, "os-release-pretty")
		if err := os.WriteFile(f, []byte("PRETTY_NAME=\"Something ubuntu-based\"\nID=custom\n"), 0644); err != nil {
			t.Fatal(err)
		}
		m := &AptManager{OsReleasePath: f}
		if m.isUbuntu() {
			t.Error("expected isUbuntu() false when only PRETTY_NAME contains ubuntu")
		}
	})

	t.Run("ID=ubuntu-derived does not match", func(t *testing.T) {
		f := filepath.Join(dir, "os-release-derived")
		if err := os.WriteFile(f, []byte("ID=ubuntu-derived\n"), 0644); err != nil {
			t.Fatal(err)
		}
		m := &AptManager{OsReleasePath: f}
		if m.isUbuntu() {
			t.Error("expected isUbuntu() false for ID=ubuntu-derived")
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

	t.Run("rejects http URI", func(t *testing.T) {
		_, err := parseDebLine("http://example.com/ubuntu jammy main")
		if err == nil {
			t.Fatal("Expected error for http:// URI")
		}
		if !strings.Contains(err.Error(), "not http://") {
			t.Errorf("Expected http rejection message, got: %s", err.Error())
		}
	})

	t.Run("rejects file URI", func(t *testing.T) {
		_, err := parseDebLine("file:///local/repo jammy main")
		if err == nil {
			t.Fatal("Expected error for file:// URI")
		}
		if !strings.Contains(err.Error(), "file://") {
			t.Errorf("Expected file:// rejection message, got: %s", err.Error())
		}
	})

	t.Run("rejects ftp URI", func(t *testing.T) {
		_, err := parseDebLine("ftp://example.com/ubuntu jammy main")
		if err == nil {
			t.Fatal("Expected error for ftp:// URI")
		}
		if !strings.Contains(err.Error(), "not allowed") {
			t.Errorf("Expected scheme rejection message, got: %s", err.Error())
		}
	})

	t.Run("rejects missing scheme", func(t *testing.T) {
		_, err := parseDebLine("example.com/ubuntu jammy main")
		if err == nil {
			t.Fatal("Expected error for URI without scheme")
		}
		if !strings.Contains(err.Error(), "missing scheme") {
			t.Errorf("Expected missing scheme message, got: %s", err.Error())
		}
	})

	t.Run("accepts https URI", func(t *testing.T) {
		repo, err := parseDebLine("https://example.com/ubuntu jammy main")
		if err != nil {
			t.Fatalf("Expected no error for https URI, got: %v", err)
		}
		if repo.URIs != "https://example.com/ubuntu" {
			t.Errorf("Expected URI 'https://example.com/ubuntu', got '%s'", repo.URIs)
		}
	})

	t.Run("deb line with signed-by option", func(t *testing.T) {
		repo, err := parseDebLine("[signed-by=/etc/apt/keyrings/test.gpg] https://example.com/repo stable main")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.SignedBy != "/etc/apt/keyrings/test.gpg" {
			t.Errorf("Expected SignedBy '/etc/apt/keyrings/test.gpg', got '%s'", repo.SignedBy)
		}
		if repo.Architectures != "" {
			t.Errorf("Expected empty Architectures, got '%s'", repo.Architectures)
		}
	})

	t.Run("deb line with arch and signed-by options", func(t *testing.T) {
		repo, err := parseDebLine("[arch=amd64 signed-by=/etc/apt/keyrings/githubcli.gpg] https://cli.github.com/packages stable main")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.Architectures != "amd64" {
			t.Errorf("Expected Architectures 'amd64', got '%s'", repo.Architectures)
		}
		if repo.SignedBy != "/etc/apt/keyrings/githubcli.gpg" {
			t.Errorf("Expected SignedBy '/etc/apt/keyrings/githubcli.gpg', got '%s'", repo.SignedBy)
		}
		if repo.URIs != "https://cli.github.com/packages" {
			t.Errorf("Expected URI 'https://cli.github.com/packages', got '%s'", repo.URIs)
		}
	})

	t.Run("deb line with signed-by but no arch", func(t *testing.T) {
		repo, err := parseDebLine("[signed-by=/etc/apt/keyrings/myrepo.gpg] https://example.com/repo focal main")
		if err != nil {
			t.Fatalf("parseDebLine failed: %v", err)
		}
		if repo.SignedBy != "/etc/apt/keyrings/myrepo.gpg" {
			t.Errorf("Expected SignedBy '/etc/apt/keyrings/myrepo.gpg', got '%s'", repo.SignedBy)
		}
		if repo.Architectures != "" {
			t.Errorf("Expected empty Architectures, got '%s'", repo.Architectures)
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

		m := &AptManager{
			SourcesDir:    tmpDir,
			SourcesPrefix: "apt-bundle-",
		}

		_, err := m.AddDebRepository(repoLine, keyPath)
		if err != nil {
			// Confirm the failure is a filesystem error, not a parse error
			if strings.Contains(err.Error(), "failed to parse deb line") {
				t.Errorf("Expected filesystem error (not parse error), got: %v", err)
			}
			t.Logf("AddDebRepository failed at filesystem step as expected: %v", err)
		}
	})

	t.Run("rejects invalid repo line", func(t *testing.T) {
		m := &AptManager{SourcesDir: tmpDir, SourcesPrefix: "apt-bundle-"}
		_, err := m.AddDebRepository("not-a-valid-repo", "")
		if err == nil {
			t.Error("Expected error for invalid repo line")
		}
	})

	t.Run("rejects http URI", func(t *testing.T) {
		m := &AptManager{SourcesDir: tmpDir, SourcesPrefix: "apt-bundle-"}
		_, err := m.AddDebRepository("http://example.com/ubuntu jammy main", "")
		if err == nil {
			t.Fatal("Expected error for http:// URI")
		}
		if !strings.Contains(err.Error(), "https") {
			t.Errorf("Expected https rejection message, got: %v", err)
		}
	})

	t.Run("explicit signed-by overrides keyPath", func(t *testing.T) {
		repoLine := "[arch=amd64 signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main"
		keyPath := filepath.Join(tmpDir, "implicit.gpg")

		if err := os.WriteFile(keyPath, []byte("fake key"), 0644); err != nil {
			t.Fatalf("Failed to create fake key file: %v", err)
		}

		m := &AptManager{
			SourcesDir:    tmpDir,
			SourcesPrefix: "apt-bundle-explicit-",
		}

		sourcePath, err := m.AddDebRepository(repoLine, keyPath)
		if err != nil {
			t.Fatalf("AddDebRepository failed: %v", err)
		}

		content, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("Failed to read sources file: %v", err)
		}

		if !strings.Contains(string(content), "Signed-By: /etc/apt/keyrings/githubcli-archive-keyring.gpg") {
			t.Errorf("Expected explicit signed-by path in sources file, got:\n%s", content)
		}
		if strings.Contains(string(content), keyPath) {
			t.Errorf("keyPath should NOT appear in sources file when deb line has signed-by=, got:\n%s", content)
		}
	})
}

func TestRemoveDebRepository(t *testing.T) {
	tmpDir := t.TempDir()
	m := &AptManager{}

	t.Run("remove existing source", func(t *testing.T) {
		sourcePath := filepath.Join(tmpDir, "test.sources")
		if err := os.WriteFile(sourcePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := m.RemoveDebRepository(sourcePath)
		if err != nil {
			t.Errorf("RemoveDebRepository failed: %v", err)
		}

		if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
			t.Error("Source file should have been removed")
		}
	})

	t.Run("remove nonexistent source", func(t *testing.T) {
		err := m.RemoveDebRepository(filepath.Join(tmpDir, "nonexistent.sources"))
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
	// Use a fake os-release so isUbuntu() doesn't read the real file
	dir := b.TempDir()
	f := filepath.Join(dir, "os-release")
	_ = os.WriteFile(f, []byte("ID=ubuntu\n"), 0644)

	mock := testutil.NewMockExecutor()
	mock.RunFunc = func(name string, args ...string) error {
		return nil
	}

	m := &AptManager{
		Executor:      mock,
		OsReleasePath: f,
		LookPath: func(file string) (string, error) {
			return "/usr/bin/add-apt-repository", nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.AddPPA("ppa:test/ppa")
	}
}
