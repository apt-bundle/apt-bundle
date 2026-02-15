package aptfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantEntries int
		wantErr     bool
		validate    func(t *testing.T, entries []Entry)
	}{
		{
			name: "valid aptfile with all directives",
			content: `# Comment line
apt curl
apt git wget

ppa ppa:deadsnakes/ppa
deb http://packages.example.com/ubuntu jammy main
key https://example.com/key.gpg`,
			wantEntries: 5,
			wantErr:     false,
			validate: func(t *testing.T, entries []Entry) {
				if entries[0].Type != EntryTypeApt || entries[0].Value != "curl" {
					t.Errorf("Expected first entry to be apt curl, got %s %s", entries[0].Type, entries[0].Value)
				}
				if entries[1].Type != EntryTypeApt || entries[1].Value != "git wget" {
					t.Errorf("Expected second entry to be apt git wget, got %s %s", entries[1].Type, entries[1].Value)
				}
				if entries[2].Type != EntryTypePPA || entries[2].Value != "ppa:deadsnakes/ppa" {
					t.Errorf("Expected third entry to be ppa, got %s %s", entries[2].Type, entries[2].Value)
				}
				if entries[3].Type != EntryTypeDeb {
					t.Errorf("Expected fourth entry to be deb, got %s", entries[3].Type)
				}
				if entries[4].Type != EntryTypeKey {
					t.Errorf("Expected fifth entry to be key, got %s", entries[4].Type)
				}
			},
		},
		{
			name:        "empty file",
			content:     "",
			wantEntries: 0,
			wantErr:     false,
		},
		{
			name: "only comments",
			content: `# Comment 1
# Comment 2
# Comment 3`,
			wantEntries: 0,
			wantErr:     false,
		},
		{
			name: "whitespace handling",
			content: `
apt   curl   

  apt    vim  

`,
			wantEntries: 2,
			wantErr:     false,
		},
		{
			name:        "invalid directive",
			content:     "invalid curl",
			wantEntries: 0,
			wantErr:     true,
		},
		{
			name:        "missing argument",
			content:     "apt",
			wantEntries: 0,
			wantErr:     true,
		},
		{
			name: "quoted values",
			content: `apt "package-with-spaces"
deb "http://example.com/repo main"`,
			wantEntries: 2,
			wantErr:     false,
			validate: func(t *testing.T, entries []Entry) {
				if entries[0].Value != "package-with-spaces" {
					t.Errorf("Expected quoted value to be unquoted, got %s", entries[0].Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "Aptfile")

			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			entries, err := Parse(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(entries) != tt.wantEntries {
				t.Errorf("Parse() got %d entries, want %d", len(entries), tt.wantEntries)
			}

			if tt.validate != nil && !tt.wantErr {
				tt.validate(t, entries)
			}
		})
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/file")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantType EntryType
		wantVal  string
		wantErr  bool
	}{
		{
			name:     "apt directive",
			line:     "apt curl",
			wantType: EntryTypeApt,
			wantVal:  "curl",
			wantErr:  false,
		},
		{
			name:     "ppa directive",
			line:     "ppa ppa:user/repo",
			wantType: EntryTypePPA,
			wantVal:  "ppa:user/repo",
			wantErr:  false,
		},
		{
			name:     "deb directive",
			line:     "deb http://example.com/ubuntu jammy main",
			wantType: EntryTypeDeb,
			wantVal:  "http://example.com/ubuntu jammy main",
			wantErr:  false,
		},
		{
			name:     "key directive",
			line:     "key https://example.com/key.gpg",
			wantType: EntryTypeKey,
			wantVal:  "https://example.com/key.gpg",
			wantErr:  false,
		},
		{
			name:    "unknown directive",
			line:    "unknown value",
			wantErr: true,
		},
		{
			name:    "no value",
			line:    "apt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parseLine(tt.line, 1, tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if entry.Type != tt.wantType {
					t.Errorf("parseLine() type = %v, want %v", entry.Type, tt.wantType)
				}
				if entry.Value != tt.wantVal {
					t.Errorf("parseLine() value = %v, want %v", entry.Value, tt.wantVal)
				}
				if entry.LineNum != 1 {
					t.Errorf("parseLine() lineNum = %v, want 1", entry.LineNum)
				}
			}
		})
	}
}

func TestSplitRespectingQuotes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "no quotes",
			input: "apt curl wget",
			want:  []string{"apt", "curl", "wget"},
		},
		{
			name:  "double quotes",
			input: `apt "package name"`,
			want:  []string{"apt", "package name"},
		},
		{
			name:  "single quotes",
			input: "apt 'package name'",
			want:  []string{"apt", "package name"},
		},
		{
			name:  "mixed spaces",
			input: "apt    curl   wget",
			want:  []string{"apt", "curl", "wget"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  []string{},
		},
		{
			name:  "complex deb line",
			input: `deb http://example.com/ubuntu jammy main contrib`,
			want:  []string{"deb", "http://example.com/ubuntu", "jammy", "main", "contrib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitRespectingQuotes(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitRespectingQuotes() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitRespectingQuotes()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "double quotes",
			input: `"value"`,
			want:  "value",
		},
		{
			name:  "single quotes",
			input: "'value'",
			want:  "value",
		},
		{
			name:  "no quotes",
			input: "value",
			want:  "value",
		},
		{
			name:  "mismatched quotes",
			input: `"value'`,
			want:  `"value'`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single char",
			input: "a",
			want:  "a",
		},
		{
			name:  "empty quotes",
			input: `""`,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unquote(tt.input)
			if got != tt.want {
				t.Errorf("unquote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractPkgName(t *testing.T) {
	tests := []struct {
		spec string
		want string
	}{
		{"curl", "curl"},
		{"nano=2.9.3-2", "nano"},
		{"docker-ce=5:19.03.13~3-0~ubuntu-focal", "docker-ce"},
		{"pkg", "pkg"},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			got := ExtractPkgName(tt.spec)
			if got != tt.want {
				t.Errorf("ExtractPkgName(%q) = %q, want %q", tt.spec, got, tt.want)
			}
		})
	}
}

func TestEntryTypes(t *testing.T) {
	tests := []struct {
		entryType EntryType
		want      string
	}{
		{EntryTypeApt, "apt"},
		{EntryTypePPA, "ppa"},
		{EntryTypeDeb, "deb"},
		{EntryTypeKey, "key"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entryType), func(t *testing.T) {
			if string(tt.entryType) != tt.want {
				t.Errorf("EntryType = %v, want %v", tt.entryType, tt.want)
			}
		})
	}
}

func TestEntry(t *testing.T) {
	entry := Entry{
		Type:     EntryTypeApt,
		Value:    "curl",
		LineNum:  5,
		Original: "apt curl",
	}

	if entry.Type != EntryTypeApt {
		t.Errorf("Entry.Type = %v, want %v", entry.Type, EntryTypeApt)
	}
	if entry.Value != "curl" {
		t.Errorf("Entry.Value = %v, want curl", entry.Value)
	}
	if entry.LineNum != 5 {
		t.Errorf("Entry.LineNum = %v, want 5", entry.LineNum)
	}
	if entry.Original != "apt curl" {
		t.Errorf("Entry.Original = %v, want 'apt curl'", entry.Original)
	}
}

func TestParseScannerError(t *testing.T) {
	t.Run("line too long triggers scanner error", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")

		longLine := "apt " + string(make([]byte, 70000))

		if err := os.WriteFile(tmpFile, []byte(longLine), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		_, err := Parse(tmpFile)
		if err == nil {
			t.Error("Expected error for line too long")
		}
	})
}
