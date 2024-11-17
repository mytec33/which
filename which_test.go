package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf) // Redirect output to the buffer

	originalArgs := os.Args
	os.Args = []string{"./which"}
	defer func() { os.Args = originalArgs }()

	printUsage()

	output := buf.String()
	expected := "Usage: ./which [options] program1 [program2 ...]\n"
	if !bytes.Contains([]byte(output), []byte(expected)) {
		t.Errorf("Expected usage to include %q, got %q", expected, output)
	}
}

func TestIsThere(t *testing.T) {
	tempDir := t.TempDir()

	// Test cases
	tests := []struct {
		name         string
		file         string
		isRegular    bool
		isExecutable bool
		expected     string
	}{
		{
			name:     "File does not exist",
			file:     "nonexistent",
			expected: "",
		},
		{
			name:      "File exists but is a directory",
			file:      "dir",
			isRegular: false,
			expected:  "",
		},
		{
			name:         "File exists but is not executable",
			file:         "regular_non_executable",
			isRegular:    true,
			isExecutable: false,
			expected:     "",
		},
		{
			name:         "File exists and is executable",
			file:         "regular_executable",
			isRegular:    true,
			isExecutable: true,
			expected:     filepath.Join(tempDir, "regular_executable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(tempDir, tt.file)
			if tt.isRegular {
				err := os.WriteFile(fullPath, []byte("dummy content"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				if tt.isExecutable {
					err := os.Chmod(fullPath, 0755)
					if err != nil {
						t.Fatalf("Failed to make test file executable: %v", err)
					}
				}
			} else if tt.file == "dir" {
				err := os.Mkdir(fullPath, 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			}

			result := isThere(tt.file, tempDir)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
