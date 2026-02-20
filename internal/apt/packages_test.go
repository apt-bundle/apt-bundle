package apt

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestIsPackageInstalled(t *testing.T) {
	t.Run("check dpkg package", func(t *testing.T) {
		if _, err := exec.LookPath("dpkg-query"); err != nil {
			t.Skip("dpkg-query not available, skipping test")
		}

		installed, err := IsPackageInstalled("dpkg")
		if err != nil {
			t.Errorf("IsPackageInstalled(dpkg) returned error: %v", err)
		}

		if !installed {
			t.Log("Note: dpkg not detected as installed, this may be expected on non-Debian systems")
		}
	})

	t.Run("check nonexistent package", func(t *testing.T) {
		if _, err := exec.LookPath("dpkg-query"); err != nil {
			t.Skip("dpkg-query not available, skipping test")
		}

		installed, err := IsPackageInstalled("definitely-not-a-real-package-12345")
		// dpkg-query returns exit 1 for nonexistent packages, so err may be non-nil
		if installed {
			t.Errorf("IsPackageInstalled() = true for nonexistent package, want false (err: %v)", err)
		}
	})

	t.Run("empty package name", func(t *testing.T) {
		if _, err := exec.LookPath("dpkg-query"); err != nil {
			t.Skip("dpkg-query not available, skipping test")
		}

		installed, err := IsPackageInstalled("")
		if err != nil {
			return
		}

		if installed {
			t.Error("IsPackageInstalled('') should return false for empty package name")
		}
	})
}

func TestInstallPackage(t *testing.T) {
	t.Run("install without sudo", func(t *testing.T) {
		err := InstallPackage("test-package-that-does-not-exist")
		if err == nil {
			t.Log("Warning: InstallPackage succeeded unexpectedly (might have sudo)")
		}
	})

	t.Run("empty package name", func(t *testing.T) {
		err := InstallPackage("")
		if err == nil {
			t.Error("InstallPackage('') should fail for empty package name")
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("update without sudo", func(t *testing.T) {
		err := Update()
		if err == nil {
			t.Log("Warning: Update succeeded unexpectedly (might have sudo)")
		}
	})
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"unix newlines", "a\nb\nc", []string{"a", "b", "c"}},
		{"windows newlines", "a\r\nb\r\nc", []string{"a", "b", "c"}},
		{"empty", "", []string{}},
		{"single line", "only", []string{"only"}},
		{"trailing newline", "a\nb\n", []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitLines() got %d lines, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitLines()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkIsPackageInstalled(b *testing.B) {
	if _, err := exec.LookPath("dpkg-query"); err != nil {
		b.Skip("dpkg-query not available, skipping benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IsPackageInstalled("dpkg")
	}
}

func TestIsPackageInstalledWithMock(t *testing.T) {
	defer ResetExecutor()

	t.Run("package is installed", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return []byte(dpkgStatusInstalled), nil
		}
		SetExecutor(mock)

		installed, err := IsPackageInstalled("curl")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !installed {
			t.Error("Expected installed=true, got false")
		}

		if len(mock.OutputCalls) != 1 {
			t.Errorf("Expected 1 output call, got %d", len(mock.OutputCalls))
		}
		if mock.OutputCalls[0][0] != "dpkg-query" {
			t.Errorf("Expected dpkg-query, got %s", mock.OutputCalls[0][0])
		}
	})

	t.Run("package is not installed - different status", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return []byte("deinstall ok config-files"), nil
		}
		SetExecutor(mock)

		installed, err := IsPackageInstalled("removed-package")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if installed {
			t.Error("Expected installed=false, got true")
		}
	})

	t.Run("package query fails - command error", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("dpkg-query: no packages found matching nonexistent")
		}
		SetExecutor(mock)

		installed, err := IsPackageInstalled("nonexistent")
		if err == nil {
			t.Error("Expected error when dpkg-query fails, got nil")
		}
		if installed {
			t.Error("Expected installed=false when command fails, got true")
		}
	})

	t.Run("verifies command arguments", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return []byte(dpkgStatusInstalled), nil
		}
		SetExecutor(mock)

		_, _ = IsPackageInstalled("test-pkg")

		expectedArgs := []string{"dpkg-query", "-W", "-f=${Status}", "test-pkg"}
		if len(mock.OutputCalls[0]) != len(expectedArgs) {
			t.Errorf("Expected %d args, got %d", len(expectedArgs), len(mock.OutputCalls[0]))
		}
		for i, arg := range expectedArgs {
			if mock.OutputCalls[0][i] != arg {
				t.Errorf("Arg %d: expected %s, got %s", i, arg, mock.OutputCalls[0][i])
			}
		}
	})
}

func TestInstallPackageWithMock(t *testing.T) {
	defer ResetExecutor()

	t.Run("successful installation", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		SetExecutor(mock)

		err := InstallPackage("curl")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(mock.RunCalls) != 1 {
			t.Errorf("Expected 1 run call, got %d", len(mock.RunCalls))
		}
		expectedArgs := []string{"apt-get", "install", "-y", "curl"}
		for i, arg := range expectedArgs {
			if mock.RunCalls[0][i] != arg {
				t.Errorf("Arg %d: expected %s, got %s", i, arg, mock.RunCalls[0][i])
			}
		}
	})

	t.Run("installation failure", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return errors.New("E: Unable to locate package nonexistent")
		}
		SetExecutor(mock)

		err := InstallPackage("nonexistent")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "failed to install package nonexistent: E: Unable to locate package nonexistent" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})
}

func TestUpdateWithMock(t *testing.T) {
	defer ResetExecutor()

	t.Run("successful update", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		SetExecutor(mock)

		err := Update()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(mock.RunCalls) != 1 {
			t.Errorf("Expected 1 run call, got %d", len(mock.RunCalls))
		}
		expectedArgs := []string{"apt-get", "update"}
		for i, arg := range expectedArgs {
			if mock.RunCalls[0][i] != arg {
				t.Errorf("Arg %d: expected %s, got %s", i, arg, mock.RunCalls[0][i])
			}
		}
	})

	t.Run("update failure", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return errors.New("E: Could not get lock /var/lib/apt/lists/lock")
		}
		SetExecutor(mock)

		err := Update()
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "failed to update package lists: E: Could not get lock /var/lib/apt/lists/lock" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})
}
