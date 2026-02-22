package apt

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// SourcesDir is where apt sources files are stored
	SourcesDir = "/etc/apt/sources.list.d"
	// SourcesPrefix is the prefix for apt-bundle managed sources files
	SourcesPrefix = "apt-bundle-"
)

// isUbuntu checks if the current system is Ubuntu by reading /etc/os-release
func (m *AptManager) isUbuntu() bool {
	data, err := os.ReadFile(m.OsReleasePath)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "ID=ubuntu") ||
		strings.Contains(content, "ID_LIKE=ubuntu")
}

// AddPPA adds a PPA repository using add-apt-repository
func (m *AptManager) AddPPA(ppa string) error {
	if !m.isUbuntu() {
		fmt.Println("⚠️  Warning: PPAs are designed for Ubuntu. Using on other distros may cause issues.")
	}
	fmt.Printf("Adding PPA: %s\n", ppa)

	if _, err := m.LookPath("add-apt-repository"); err != nil {
		return fmt.Errorf("add-apt-repository not found. Please install software-properties-common")
	}

	if err := m.runCommand("add-apt-repository", "-y", ppa); err != nil {
		return wrapCommandError(err, "add PPA", ppa)
	}

	fmt.Printf("✓ PPA %s added successfully\n", ppa)
	return nil
}

// validateRepoURI ensures the repository URI uses https://. Rejects http://, file://, and other schemes.
func validateRepoURI(repoURI string) error {
	u, err := url.Parse(repoURI)
	if err != nil {
		return fmt.Errorf("invalid repository URI: %w", err)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		return fmt.Errorf("repository URI must use https://, not http:// (rejected for security)")
	case "file":
		return fmt.Errorf("file:// repository URIs are not allowed (rejected for security)")
	case "":
		return fmt.Errorf("invalid repository URI: missing scheme (use https://)")
	default:
		return fmt.Errorf("repository URI scheme %q not allowed; use https://", u.Scheme)
	}
}

// DebRepository represents a parsed deb repository configuration
type DebRepository struct {
	Types         string // "deb" or "deb-src"
	URIs          string
	Suites        string
	Components    string
	Architectures string
	SignedBy      string // Path to the GPG key file
}

// AddDebRepository adds a deb repository in DEB822 format to /etc/apt/sources.list.d/
// keyPath is optional - if provided, it will be used for the Signed-By field
func AddDebRepository(repoLine string, keyPath string) (string, error) {
	fmt.Printf("Adding deb repository: %s\n", repoLine)

	repo, err := parseDebLine(repoLine)
	if err != nil {
		return "", fmt.Errorf("failed to parse deb line: %w", err)
	}

	// Set the Signed-By field if a key path was provided
	if keyPath != "" {
		repo.SignedBy = keyPath
	}

	// Generate filename from repo hash
	hash := sha256.Sum256([]byte(repoLine))
	filename := fmt.Sprintf("%s%x.sources", SourcesPrefix, hash[:8])
	sourcePath := filepath.Join(SourcesDir, filename)

	// Check if source already exists (idempotency)
	if _, err := os.Stat(sourcePath); err == nil {
		fmt.Printf("✓ Repository already configured: %s\n", sourcePath)
		return sourcePath, nil
	}

	// Ensure the sources directory exists
	if err := os.MkdirAll(SourcesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create sources directory: %w", err)
	}

	// Generate DEB822 format content
	content := repo.ToDEB822()

	// Write the sources file
	if err := os.WriteFile(sourcePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write sources file: %w", err)
	}

	fmt.Printf("✓ Repository added: %s\n", sourcePath)
	return sourcePath, nil
}

// parseDebLine parses a traditional deb line into a DebRepository struct
// Format: [options] uri suite [component1] [component2] [...]
// Example: [arch=amd64] https://download.docker.com/linux/ubuntu focal stable
func parseDebLine(line string) (*DebRepository, error) {
	repo := &DebRepository{
		Types: "deb",
	}

	// Remove leading "deb " or "deb-src " if present (check before trim)
	if strings.HasPrefix(line, "deb-src ") {
		repo.Types = "deb-src"
		line = strings.TrimPrefix(line, "deb-src ")
	} else if strings.HasPrefix(line, "deb ") {
		line = strings.TrimPrefix(line, "deb ")
	}

	// Extract options in brackets [key=value key2=value2]
	optionsRegex := regexp.MustCompile(`^\[([^\]]+)\]\s*`)
	if matches := optionsRegex.FindStringSubmatch(line); len(matches) > 1 {
		options := matches[1]
		line = optionsRegex.ReplaceAllString(line, "")

		// Parse arch option
		archRegex := regexp.MustCompile(`arch=([^\s]+)`)
		if archMatches := archRegex.FindStringSubmatch(options); len(archMatches) > 1 {
			repo.Architectures = archMatches[1]
		}
	}

	// Split remaining parts: uri suite [components...]
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid deb line: expected at least URI and suite")
	}

	repo.URIs = parts[0]

	if err := validateRepoURI(repo.URIs); err != nil {
		return nil, err
	}

	repo.Suites = parts[1]

	if len(parts) > 2 {
		repo.Components = strings.Join(parts[2:], " ")
	}

	return repo, nil
}

// ToDEB822 converts the repository to DEB822 format
func (r *DebRepository) ToDEB822() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Types: %s", r.Types))
	lines = append(lines, fmt.Sprintf("URIs: %s", r.URIs))
	lines = append(lines, fmt.Sprintf("Suites: %s", r.Suites))

	if r.Components != "" {
		lines = append(lines, fmt.Sprintf("Components: %s", r.Components))
	}

	if r.Architectures != "" {
		lines = append(lines, fmt.Sprintf("Architectures: %s", r.Architectures))
	}

	if r.SignedBy != "" {
		lines = append(lines, fmt.Sprintf("Signed-By: %s", r.SignedBy))
	}

	return strings.Join(lines, "\n") + "\n"
}

// RemoveDebRepository removes a sources file
func RemoveDebRepository(sourcePath string) error {
	if err := os.Remove(sourcePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove sources file: %w", err)
	}
	return nil
}
