package apt

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// KeyringDir is where apt-bundle stores GPG keys (scoped to repos, not globally trusted)
	KeyringDir = "/etc/apt/keyrings"
	// KeyPrefix is the prefix for apt-bundle managed key files
	KeyPrefix = "apt-bundle-"
)

// validateKeyURL ensures the URL uses https://. Rejects http://, file://, and other schemes.
func validateKeyURL(keyURL string) error {
	u, err := url.Parse(keyURL)
	if err != nil {
		return fmt.Errorf("invalid key URL: %w", err)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		return fmt.Errorf("key URL must use https://, not http:// (rejected for security)")
	case "file":
		return fmt.Errorf("file:// key URLs are not allowed (rejected for security)")
	case "":
		return fmt.Errorf("invalid key URL: missing scheme (use https://)")
	default:
		return fmt.Errorf("key URL scheme %q not allowed; use https://", u.Scheme)
	}
}

// KeyPathForURL returns the path where AddGPGKey would store the key for the given URL
func KeyPathForURL(keyURL string) string {
	hash := sha256.Sum256([]byte(keyURL))
	filename := fmt.Sprintf("%s%x.gpg", KeyPrefix, hash[:8])
	return filepath.Join(KeyringDir, filename)
}

// keyHTTPClient is used for key downloads; has timeout to avoid hanging
var keyHTTPClient = &http.Client{Timeout: 30 * time.Second}

// httpGet is the function used to make HTTP requests (overridable for testing)
var httpGet = keyHTTPClient.Get

// AddGPGKey downloads and adds a GPG key from a URL
// Returns the path to the saved key file for use with Signed-By in DEB822 format
// Only https:// URLs are allowed; http:// and file:// are rejected for security.
func AddGPGKey(keyURL string) (string, error) {
	if err := validateKeyURL(keyURL); err != nil {
		return "", err
	}

	fmt.Printf("Adding GPG key from: %s\n", keyURL)

	hash := sha256.Sum256([]byte(keyURL))
	filename := fmt.Sprintf("%s%x.gpg", KeyPrefix, hash[:8])
	keyPath := filepath.Join(KeyringDir, filename)

	// Check if key already exists (idempotency)
	if _, err := os.Stat(keyPath); err == nil {
		fmt.Printf("✓ GPG key already exists: %s\n", keyPath)
		return keyPath, nil
	}

	// Ensure the keyring directory exists
	if err := os.MkdirAll(KeyringDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create keyring directory: %w", err)
	}

	// Download the key
	resp, err := httpGet(keyURL)
	if err != nil {
		return "", fmt.Errorf("failed to download key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download key: HTTP %d", resp.StatusCode)
	}

	keyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read key data: %w", err)
	}

	// Check if the key is ASCII armored and needs dearmoring
	if isArmoredKey(keyData) {
		keyData, err = dearmorKey(keyData)
		if err != nil {
			return "", fmt.Errorf("failed to dearmor key: %w", err)
		}
	}

	// Write the key to the keyring directory
	if err := os.WriteFile(keyPath, keyData, 0644); err != nil {
		return "", fmt.Errorf("failed to write key file: %w", err)
	}

	fmt.Printf("✓ GPG key saved to: %s\n", keyPath)
	return keyPath, nil
}

// isArmoredKey checks if the key data is ASCII armored
func isArmoredKey(data []byte) bool {
	return strings.Contains(string(data), "-----BEGIN PGP PUBLIC KEY BLOCK-----")
}

// dearmorKey converts an ASCII armored key to binary format using gpg --dearmor
func dearmorKey(data []byte) ([]byte, error) {
	// Create a temp file for the armored key
	tmpFile, err := os.CreateTemp("", "apt-bundle-key-*.asc")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, err
	}
	tmpFile.Close()

	// Create a temp file for the dearmored output
	outFile, err := os.CreateTemp("", "apt-bundle-key-*.gpg")
	if err != nil {
		return nil, err
	}
	outPath := outFile.Name()
	outFile.Close()
	defer os.Remove(outPath)

	// Run gpg --dearmor
	if err := runCommand("gpg", "--dearmor", "-o", outPath, tmpFile.Name()); err != nil {
		return nil, fmt.Errorf("gpg --dearmor failed: %w", err)
	}

	// Read the dearmored key
	return os.ReadFile(outPath)
}

// RemoveGPGKey removes a GPG key file
func RemoveGPGKey(keyPath string) error {
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove key: %w", err)
	}
	return nil
}

// SetHTTPGet sets the HTTP get function (for testing only)
func SetHTTPGet(f func(string) (*http.Response, error)) {
	httpGet = f
}

// ResetHTTPGet resets the HTTP get function to default (for testing only)
func ResetHTTPGet() {
	httpGet = keyHTTPClient.Get
}
