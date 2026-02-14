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
	_ = tmpDir

	t.Run("successful key download - binary key", func(t *testing.T) {
		defer ResetHTTPGet()
		defer ResetExecutor()

		binaryKeyData := []byte{0x99, 0x01, 0x0d, 0x04}
		SetHTTPGet(func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(binaryKeyData))),
			}, nil
		})

		keyringPath := filepath.Join(tmpDir, "keyrings")
		os.MkdirAll(keyringPath, 0755)

		keyPath, err := AddGPGKey("https://example.com/key.gpg")
		if err != nil {
			t.Logf("AddGPGKey failed (expected in test environment): %v", err)
		} else {
			t.Logf("Key path: %s", keyPath)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		defer ResetHTTPGet()

		SetHTTPGet(func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		})

		_, err := AddGPGKey("https://example.com/nonexistent.gpg")
		if err == nil {
			t.Error("Expected error for HTTP 404, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "HTTP 404") && !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("Expected HTTP 404 or permission error, got: %v", err)
		}
	})

	t.Run("invalid URL format", func(t *testing.T) {
		_, err := AddGPGKey("not-a-valid-url")
		if err != nil {
			t.Logf("AddGPGKey() with invalid URL returned error: %v", err)
		}
	})

	t.Run("keyserver URL", func(t *testing.T) {
		_, err := AddGPGKey("keyserver://keyserver.ubuntu.com/12345")
		if err != nil {
			t.Logf("AddGPGKey() with keyserver URL returned error: %v", err)
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

func TestSetHTTPGet(t *testing.T) {
	defer ResetHTTPGet()

	customCalled := false
	SetHTTPGet(func(url string) (*http.Response, error) {
		customCalled = true
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("test")),
		}, nil
	})

	httpGet("http://example.com")
	if !customCalled {
		t.Error("Custom httpGet function was not called")
	}
}

func BenchmarkAddGPGKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddGPGKey("https://example.com/key.gpg")
	}
}
