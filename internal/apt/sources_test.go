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

func TestParseDebLineToSource(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantType string
		wantLine string
		wantOK   bool
	}{
		{
			name:     "deb line",
			line:     "deb https://example.com/ubuntu jammy main",
			wantType: "deb",
			wantLine: "deb https://example.com/ubuntu jammy main",
			wantOK:   true,
		},
		{
			name:     "deb line with arch",
			line:     "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable",
			wantType: "deb",
			wantLine: "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable",
			wantOK:   true,
		},
		{
			name:     "ppa line",
			line:     "deb http://ppa.launchpad.net/ondrej/php/ubuntu jammy main",
			wantType: "ppa",
			wantLine: "ppa ppa:ondrej/php",
			wantOK:   true,
		},
		{
			name:   "comment",
			line:   "# deb http://example.com/ubuntu jammy main",
			wantOK: false,
		},
		{
			name:   "empty",
			line:   "",
			wantOK: false,
		},
		{
			name:   "too few parts",
			line:   "deb https://example.com",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseDebLineToSource(tt.line)
			if ok != tt.wantOK {
				t.Errorf("parseDebLineToSource() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if tt.wantOK {
				if got.Type != tt.wantType {
					t.Errorf("parseDebLineToSource() type = %v, want %v", got.Type, tt.wantType)
				}
				if got.AptfileLine != tt.wantLine {
					t.Errorf("parseDebLineToSource() AptfileLine = %v, want %v", got.AptfileLine, tt.wantLine)
				}
			}
		})
	}
}

func TestReadDEB822File(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantLine string
		wantOK   bool
	}{
		{
			name: "minimal deb",
			content: `Types: deb
URIs: https://example.com/apt
Suites: focal
`,
			wantLine: "deb https://example.com/apt focal",
			wantOK:   true,
		},
		{
			name: "full with arch and components",
			content: `Types: deb
URIs: https://example.com/apt
Suites: focal
Components: main
Architectures: amd64
`,
			wantLine: "deb [arch=amd64] https://example.com/apt focal main",
			wantOK:   true,
		},
		{
			name: "deb-src",
			content: `Types: deb-src
URIs: https://example.com/apt
Suites: jammy
`,
			wantLine: "deb-src https://example.com/apt jammy",
			wantOK:   true,
		},
		{
			name: "missing uris",
			content: `Types: deb
Suites: focal
`,
			wantOK: false,
		},
		{
			name: "missing suites",
			content: `Types: deb
URIs: https://example.com/apt
`,
			wantOK: false,
		},
		{
			name: "default uri skipped",
			content: `Types: deb
URIs: http://archive.ubuntu.com/ubuntu
Suites: jammy
`,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.sources")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			entries, err := readDEB822File(path)
			if err != nil {
				t.Fatalf("readDEB822File: %v", err)
			}
			if tt.wantOK {
				if len(entries) != 1 {
					t.Errorf("readDEB822File() got %d entries, want 1", len(entries))
					return
				}
				if entries[0].AptfileLine != tt.wantLine {
					t.Errorf("readDEB822File() AptfileLine = %v, want %v", entries[0].AptfileLine, tt.wantLine)
				}
			} else {
				if len(entries) != 0 {
					t.Errorf("readDEB822File() got %d entries, want 0", len(entries))
				}
			}
		})
	}
}

func TestParseDEB822StanzaSignedBy(t *testing.T) {
	t.Run("Signed-By with companion URL file", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "apt-bundle-abc123.gpg")
		urlFilePath := keyPath + ".url"
		const keyURL = "https://repo.charm.sh/apt/gpg.key"

		if err := os.WriteFile(keyPath, []byte("fake key"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(urlFilePath, []byte(keyURL), 0644); err != nil {
			t.Fatal(err)
		}

		content := "Types: deb\nURIs: https://repo.charm.sh/apt\nSuites: *\nComponents: *\nSigned-By: " + keyPath + "\n"
		path := filepath.Join(tmpDir, "charm.sources")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		entries, err := readDEB822File(path)
		if err != nil {
			t.Fatalf("readDEB822File: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].KeyURL != keyURL {
			t.Errorf("KeyURL = %q, want %q", entries[0].KeyURL, keyURL)
		}
	})

	t.Run("Signed-By without companion URL file", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := "Types: deb\nURIs: https://example.com/apt\nSuites: focal\nSigned-By: /etc/apt/keyrings/no-url-file.gpg\n"
		path := filepath.Join(tmpDir, "test.sources")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		entries, err := readDEB822File(path)
		if err != nil {
			t.Fatalf("readDEB822File: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].KeyURL != "" {
			t.Errorf("KeyURL should be empty when URL file absent, got %q", entries[0].KeyURL)
		}
	})

	t.Run("no Signed-By field", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := "Types: deb\nURIs: https://example.com/apt\nSuites: focal\nComponents: main\n"
		path := filepath.Join(tmpDir, "test.sources")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		entries, err := readDEB822File(path)
		if err != nil {
			t.Fatalf("readDEB822File: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].KeyURL != "" {
			t.Errorf("KeyURL should be empty when no Signed-By, got %q", entries[0].KeyURL)
		}
	})
}

func TestReadDEB822FileMultiStanza(t *testing.T) {
	content := `Types: deb
URIs: https://repo-a.example.com/apt
Suites: focal
Components: main

Types: deb
URIs: https://repo-b.example.com/apt
Suites: jammy
Components: stable
Architectures: amd64
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.sources")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := readDEB822File(path)
	if err != nil {
		t.Fatalf("readDEB822File: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for multi-stanza file, got %d", len(entries))
	}
	if entries[0].AptfileLine != "deb https://repo-a.example.com/apt focal main" {
		t.Errorf("stanza 1: got %q", entries[0].AptfileLine)
	}
	if entries[1].AptfileLine != "deb [arch=amd64] https://repo-b.example.com/apt jammy stable" {
		t.Errorf("stanza 2: got %q", entries[1].AptfileLine)
	}
}
