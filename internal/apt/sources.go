package apt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SourcesListPath is the path to the main sources.list (overridable for testing)
var SourcesListPath = "/etc/apt/sources.list"

// SourceEntry represents one repository line to emit in an Aptfile
type SourceEntry struct {
	Type        string // "ppa" or "deb"
	AptfileLine string
}

// defaultURIs are distro default repository hosts; entries from these are skipped when dumping
var defaultURIs = []string{
	"archive.ubuntu.com",
	"security.ubuntu.com",
	"deb.debian.org",
	"ftp.debian.org",
}

func isDefaultURI(uri string) bool {
	for _, d := range defaultURIs {
		if strings.Contains(uri, d) {
			return true
		}
	}
	return false
}

// ppaLaunchpadRegex matches PPA deb lines: deb http://ppa.launchpad.net/OWNER/PPA/ubuntu ...
var ppaLaunchpadRegex = regexp.MustCompile(`ppa\.launchpad\.net/([^/]+)/([^/]+)/`)

func parseDebLineToSource(line string) (SourceEntry, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return SourceEntry{}, false
	}
	// Handle deb and deb-src lines
	var rest string
	if strings.HasPrefix(line, "deb-src ") {
		rest = strings.TrimPrefix(line, "deb-src ")
	} else if strings.HasPrefix(line, "deb ") {
		rest = strings.TrimPrefix(line, "deb ")
	} else {
		return SourceEntry{}, false
	}
	// Extract URI (first field, or after [options])
	if strings.HasPrefix(rest, "[") {
		idx := strings.Index(rest, "]")
		if idx == -1 {
			return SourceEntry{}, false
		}
		rest = strings.TrimSpace(rest[idx+1:])
	}
	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return SourceEntry{}, false
	}
	uri := fields[0]
	if matches := ppaLaunchpadRegex.FindStringSubmatch(uri); len(matches) >= 3 {
		owner := matches[1]
		ppa := matches[2]
		return SourceEntry{Type: "ppa", AptfileLine: "ppa ppa:" + owner + "/" + ppa}, true
	}
	if isDefaultURI(uri) {
		return SourceEntry{}, false
	}
	return SourceEntry{Type: "deb", AptfileLine: line}, true
}

// ListCustomSources reads sources.list and sources.list.d and returns Aptfile lines for custom repos
func ListCustomSources(sourcesListPath, sourcesDir string) ([]SourceEntry, error) {
	var entries []SourceEntry
	seen := make(map[string]bool)

	// Read main sources.list
	data, err := os.ReadFile(sourcesListPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading %s: %w", sourcesListPath, err)
	}
	if err == nil {
		lines, _ := splitLines(string(data))
		for _, line := range lines {
			if e, ok := parseDebLineToSource(line); ok && !seen[e.AptfileLine] {
				seen[e.AptfileLine] = true
				entries = append(entries, e)
			}
		}
	}

	// Read sources.list.d
	dirEntries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return entries, nil // directory missing is ok
	}
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		name := de.Name()
		if !strings.HasSuffix(name, ".list") && !strings.HasSuffix(name, ".sources") {
			continue
		}
		path := filepath.Join(sourcesDir, name)
		if strings.HasSuffix(name, ".list") {
			listEntries, err := readListFile(path)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", path, err)
			}
			for _, e := range listEntries {
				if !seen[e.AptfileLine] {
					seen[e.AptfileLine] = true
					entries = append(entries, e)
				}
			}
		} else {
			sourceEntries, err := readDEB822File(path)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", path, err)
			}
			for _, e := range sourceEntries {
				if !seen[e.AptfileLine] {
					seen[e.AptfileLine] = true
					entries = append(entries, e)
				}
			}
		}
	}

	// Stable order: ppa first, then deb; same type sorted by line
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return entries[i].Type < entries[j].Type
		}
		return entries[i].AptfileLine < entries[j].AptfileLine
	})
	return entries, nil
}

func readListFile(path string) ([]SourceEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var entries []SourceEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if e, ok := parseDebLineToSource(sc.Text()); ok {
			entries = append(entries, e)
		}
	}
	return entries, sc.Err()
}

// readDEB822File parses a DEB822 .sources file and returns Aptfile entries (deb only; PPA not in .sources typically).
// Supports multi-stanza files where stanzas are separated by blank lines.
func readDEB822File(path string) ([]SourceEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines, _ := splitLines(string(data))
	var entries []SourceEntry
	var stanzaLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(stanzaLines) > 0 {
				if entry, ok := parseDEB822Stanza(stanzaLines); ok {
					entries = append(entries, entry)
				}
				stanzaLines = nil
			}
			continue
		}
		stanzaLines = append(stanzaLines, line)
	}
	// Handle final stanza (no trailing blank line)
	if len(stanzaLines) > 0 {
		if entry, ok := parseDEB822Stanza(stanzaLines); ok {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// parseDEB822Stanza parses a single DEB822 stanza (key-value pairs) into a SourceEntry.
func parseDEB822Stanza(lines []string) (SourceEntry, bool) {
	var types, uris, suites, components, architectures string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			switch key {
			case "Types":
				types = val
			case "URIs":
				uris = val
			case "Suites":
				suites = val
			case "Components":
				components = val
			case "Architectures":
				architectures = val
			}
		}
	}
	if uris == "" || suites == "" {
		return SourceEntry{}, false
	}
	if isDefaultURI(uris) {
		return SourceEntry{}, false
	}
	// Build deb line: [arch=X] URI Suite Components
	debType := "deb"
	if types == "deb-src" {
		debType = "deb-src"
	}
	line := debType + " "
	if architectures != "" {
		line += "[arch=" + architectures + "] "
	}
	line += uris + " " + suites
	if components != "" {
		line += " " + components
	}
	return SourceEntry{Type: "deb", AptfileLine: line}, true
}
