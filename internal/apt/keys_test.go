package apt

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddGPGKey(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("successful key download - binary key", func(t *testing.T) {
		keyringPath := filepath.Join(tmpDir, "keyrings")
		if err := os.MkdirAll(keyringPath, 0755); err != nil {
			t.Fatalf("Failed to create keyring dir: %v", err)
		}

		binaryKeyData := []byte{0x99, 0x01, 0x0d, 0x04}
		m := &AptManager{
			KeyringDir: keyringPath,
			HTTPGet: func(url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(binaryKeyData))),
				}, nil
			},
		}

		keyPath, err := m.AddGPGKey("https://example.com/key.gpg")
		if err != nil {
			t.Fatalf("AddGPGKey failed: %v", err)
		}
		if keyPath == "" {
			t.Error("Expected non-empty key path")
		}
		if _, err := os.Stat(keyPath); err != nil {
			t.Errorf("Key file should exist at %s: %v", keyPath, err)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		m := &AptManager{
			KeyringDir: tmpDir,
			HTTPGet: func(url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		_, err := m.AddGPGKey("https://example.com/nonexistent.gpg")
		if err == nil {
			t.Error("Expected error for HTTP 404, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "HTTP 404") && !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("Expected HTTP 404 or permission error, got: %v", err)
		}
	})

	t.Run("reject http://", func(t *testing.T) {
		m := NewAptManager()
		_, err := m.AddGPGKey("http://example.com/key.gpg")
		if err == nil {
			t.Error("Expected error for http:// URL")
		}
		if err != nil && !strings.Contains(err.Error(), "https") {
			t.Errorf("Expected error to mention https, got: %v", err)
		}
	})

	t.Run("reject file://", func(t *testing.T) {
		m := NewAptManager()
		_, err := m.AddGPGKey("file:///etc/passwd")
		if err == nil {
			t.Error("Expected error for file:// URL")
		}
		if err != nil && !strings.Contains(err.Error(), "file") {
			t.Errorf("Expected error to mention file, got: %v", err)
		}
	})

	t.Run("reject invalid URL", func(t *testing.T) {
		m := NewAptManager()
		_, err := m.AddGPGKey("not-a-valid-url")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("reject keyserver://", func(t *testing.T) {
		m := NewAptManager()
		_, err := m.AddGPGKey("keyserver://keyserver.ubuntu.com/12345")
		if err == nil {
			t.Error("Expected error for keyserver:// URL")
		}
	})
}

func TestIsArmoredKey(t *testing.T) {
	t.Run("armored key", func(t *testing.T) {
		armoredKey := []byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----
mQINBF...
-----END PGP PUBLIC KEY BLOCK-----`)
		if !isArmoredKey(armoredKey) {
			t.Error("Expected armored key to be detected")
		}
	})

	t.Run("binary key", func(t *testing.T) {
		binaryKey := []byte{0x99, 0x01, 0x0d, 0x04, 0x5f, 0x5e}
		if isArmoredKey(binaryKey) {
			t.Error("Expected binary key not to be detected as armored")
		}
	})
}

func TestRemoveGPGKey(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("remove existing key", func(t *testing.T) {
		keyPath := filepath.Join(tmpDir, "test.gpg")
		if err := os.WriteFile(keyPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := RemoveGPGKey(keyPath)
		if err != nil {
			t.Errorf("RemoveGPGKey failed: %v", err)
		}

		if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
			t.Error("Key file should have been removed")
		}
	})

	t.Run("remove nonexistent key", func(t *testing.T) {
		err := RemoveGPGKey(filepath.Join(tmpDir, "nonexistent.gpg"))
		if err != nil {
			t.Errorf("RemoveGPGKey should not error for nonexistent file: %v", err)
		}
	})
}

func TestKeyPathForURL(t *testing.T) {
	m := &AptManager{KeyringDir: "/test/keyrings"}
	path := m.KeyPathForURL("https://example.com/key.gpg")
	if path == "" {
		t.Error("Expected non-empty path")
	}
	if !strings.HasPrefix(path, "/test/keyrings/") {
		t.Errorf("Expected path to start with /test/keyrings/, got %s", path)
	}
}

func BenchmarkAddGPGKey(b *testing.B) {
	keyringPath := b.TempDir()

	binaryKeyData := []byte{0x99, 0x01, 0x0d, 0x04}
	m := &AptManager{
		KeyringDir: keyringPath,
		HTTPGet: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(binaryKeyData))),
			}, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Remove key file each iteration so AddGPGKey doesn't short-circuit on idempotency check
		_ = os.RemoveAll(keyringPath)
		_ = os.MkdirAll(keyringPath, 0755)
		_, _ = m.AddGPGKey("https://example.com/key.gpg")
	}
}
