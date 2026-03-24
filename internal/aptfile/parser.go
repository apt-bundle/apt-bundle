package aptfile

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// EntryType represents the kind of directive in an Aptfile line (apt, ppa, deb, or key).
type EntryType string

const (
	EntryTypeApt EntryType = "apt"
	EntryTypePPA EntryType = "ppa"
	EntryTypeDeb EntryType = "deb"
	EntryTypeKey EntryType = "key"
)

// validKeyNameRe matches valid key alias names used in "key <url> as <name>".
// Names must start with an alphanumeric character and may contain alphanumeric
// characters, dots, hyphens, or underscores.
var validKeyNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// Entry represents a single parsed directive from an Aptfile, including
// the directive type, its argument value, the source line number, and
// the original unparsed line text.
type Entry struct {
	Type     EntryType
	Value    string
	Name     string // optional alias for key directives: "key <url> as <name>"
	LineNum  int
	Original string
}

// Parse reads an Aptfile at the given path and returns the list of entries.
// Blank lines and comment lines (starting with #) are skipped.
func Parse(filePath string) ([]Entry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Aptfile: %w", err)
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		original := line

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		line = stripInlineComment(line)
		if line == "" {
			continue
		}

		entry, err := parseLine(line, lineNum, original)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

func parseLine(line string, lineNum int, original string) (Entry, error) {
	parts := splitRespectingQuotes(line)
	if len(parts) < 2 {
		return Entry{}, fmt.Errorf("invalid line format: expected 'directive argument'")
	}

	directive := parts[0]
	value := strings.Join(parts[1:], " ")
	value = unquote(value)

	var entryType EntryType
	switch directive {
	case "apt":
		entryType = EntryTypeApt
	case "ppa":
		entryType = EntryTypePPA
	case "deb":
		entryType = EntryTypeDeb
	case "key":
		entryType = EntryTypeKey
	default:
		return Entry{}, fmt.Errorf("unknown directive: %s", directive)
	}

	// For key directives, parse optional "as <name>" suffix.
	// Example: key https://example.com/key.gpg as mykey
	var name string
	if entryType == EntryTypeKey {
		if idx := strings.LastIndex(value, " as "); idx >= 0 {
			potentialName := strings.TrimSpace(value[idx+4:])
			// Only treat as a name if it's a single word (no spaces) and non-empty.
			if potentialName != "" && !strings.Contains(potentialName, " ") {
				if !validKeyNameRe.MatchString(potentialName) {
					return Entry{}, fmt.Errorf("invalid key name %q: must start with alphanumeric and contain only alphanumeric, dot, hyphen, or underscore", potentialName)
				}
				name = potentialName
				value = strings.TrimSpace(value[:idx])
			}
		}
	}

	return Entry{
		Type:     entryType,
		Value:    value,
		Name:     name,
		LineNum:  lineNum,
		Original: original,
	}, nil
}

func splitRespectingQuotes(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range s {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = 0
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// ExtractPkgName returns the package name from a spec like "curl" or "nano=2.9.3-2".
func ExtractPkgName(spec string) string {
	if idx := strings.Index(spec, "="); idx > 0 {
		return spec[:idx]
	}
	return spec
}

// stripInlineComment removes an inline comment from a line. A ` #` (whitespace
// followed by `#`) outside of a quoted string is treated as the start of an
// inline comment; everything from that point to end-of-line is discarded.
func stripInlineComment(line string) string {
	inQuotes := false
	quoteChar := rune(0)
	runes := []rune(line)
	for i, r := range runes {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = 0
		case r == '#' && !inQuotes && i > 0 && (runes[i-1] == ' ' || runes[i-1] == '\t'):
			return strings.TrimRight(string(runes[:i]), " \t")
		}
	}
	return line
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
