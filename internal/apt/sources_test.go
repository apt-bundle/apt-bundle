package apt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListCustomSources(t *testing.T) {
	tmpDir := t.TempDir()
	sourcesList := filepath.Join(tmpDir, "sources.list")
	sourcesDir := filepath.Join(tmpDir, "sources.list.d")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Empty: no custom sources
	entries, err := ListCustomSources(sourcesList, sourcesDir)
	if err != nil {
		t.Fatalf("ListCustomSources: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	// Write a .list file with one PPA and one deb
	listPath := filepath.Join(sourcesDir, "custom.list")
	content := `# comment
deb http://ppa.launchpad.net/ondrej/php/ubuntu jammy main
deb [arch=amd64] https://download.docker.com/linux/ubuntu jammy stable
`
	if err := os.WriteFile(listPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err = ListCustomSources(sourcesList, sourcesDir)
	if err != nil {
		t.Fatalf("ListCustomSources: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	var ppaFound, debFound bool
	for _, e := range entries {
		if e.Type == "ppa" && strings.Contains(e.AptfileLine, "ppa:ondrej/php") {
			ppaFound = true
		}
		if e.Type == "deb" && strings.Contains(e.AptfileLine, "download.docker.com") {
			debFound = true
		}
	}
	if !ppaFound {
		t.Error("expected ppa entry for ondrej/php")
	}
	if !debFound {
		t.Error("expected deb entry for docker")
	}

	// Write a .sources file (DEB822)
	sourcesPath := filepath.Join(sourcesDir, "test.sources")
	sourcesContent := `Types: deb
URIs: https://example.com/apt
Suites: focal
Components: main
Architectures: amd64
`
	if err := os.WriteFile(sourcesPath, []byte(sourcesContent), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err = ListCustomSources(sourcesList, sourcesDir)
	if err != nil {
		t.Fatalf("ListCustomSources: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries after .sources, got %d", len(entries))
	}
}

func TestListCustomSources_skipsDefault(t *testing.T) {
	tmpDir := t.TempDir()
	sourcesList := filepath.Join(tmpDir, "sources.list")
	sourcesDir := filepath.Join(tmpDir, "sources.list.d")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `deb http://archive.ubuntu.com/ubuntu jammy main
deb http://security.ubuntu.com/ubuntu jammy-security main
deb https://custom.example.com/ubuntu jammy main
`
	if err := os.WriteFile(sourcesList, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := ListCustomSources(sourcesList, sourcesDir)
	if err != nil {
		t.Fatalf("ListCustomSources: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 custom entry (skip defaults), got %d", len(entries))
	}
	if len(entries) > 0 && !strings.Contains(entries[0].AptfileLine, "custom.example.com") {
		t.Errorf("expected custom.example.com entry, got %s", entries[0].AptfileLine)
	}
}
