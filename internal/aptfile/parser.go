package aptfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type EntryType string

const (
	EntryTypeApt EntryType = "apt"
	EntryTypePPA EntryType = "ppa"
	EntryTypeDeb EntryType = "deb"
	EntryTypeKey EntryType = "key"
)

type Entry struct {
	Type     EntryType
	Value    string
	LineNum  int
	Original string
}

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

	return Entry{
		Type:     entryType,
		Value:    value,
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

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
