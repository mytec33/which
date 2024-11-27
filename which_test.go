package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const DARWIN_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG = "/usr/bin/which: illegal option -- z\nusage: which [-as] program ...\n"
const DARWIN_EXPECTED_NATIVE_OUTPUT_INVALID_ARG = DARWIN_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG
const LINUX_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG = "Illegal option -z\nUsage: ./which [-as] args ...\n"
const LINUX_EXPECTED_NATIVE_OUTPUT_INVALID_ARG = "Illegal option -z\nUsage: /usr/bin/which [-as] args\n"
const OPENBSD_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG = "which: unknown option -- z\nusage: which [-a] name ...\n"
const OPENBSD_EXPECTED_NATIVE_OUTPUT_INVALID_ARG = "which: unknown option -- z\nusage: which [-a] name ...\n"

// Override shell built-in, like on macOS
const NATIVE_WHICH = "/usr/bin/which"
const CUSTOM_WHICH = "./which"

func TestPrintUsage(t *testing.T) {
	var expected = ""

	output, exitCode, err := runCommand(CUSTOM_WHICH)

	if err != nil {
		t.Fatalf("Failed to run command %q: %v", CUSTOM_WHICH, err)
	}

	if runtime.GOOS == "darwin" {
		expected = "usage: which [-as] program ...\n"
	} else if runtime.GOOS == "openbsd" {
		expected = "usage: which [-a] name ...\n"
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !bytes.Contains([]byte(output), []byte(expected)) {
		t.Errorf("Expected %q in output, but got %q", expected, output)
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

func runCommand(cmd string, args ...string) (string, int, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stdout = &out
	c.Stderr = &stderr
	err := c.Run()

	// Get exit code
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		return "", 0, err
	}

	return out.String() + stderr.String(), exitCode, nil
}

func TestWhich(t *testing.T) {
	testCases := []struct {
		description         string
		args                []string
		nativeExitCode      int
		customExitCode      int
		expectedOutputMatch bool
	}{
		{
			description:         "Single found program",
			args:                []string{"ls"},
			nativeExitCode:      0,
			customExitCode:      0,
			expectedOutputMatch: true,
		},
		{
			description:         "Multiple found programs",
			args:                []string{"ls", "grep", "ssh"},
			nativeExitCode:      0,
			customExitCode:      0,
			expectedOutputMatch: true,
		},
		{
			description:         "Mix of found and not found",
			args:                []string{"ls", "nonexistent", "bash"},
			nativeExitCode:      1,
			customExitCode:      1,
			expectedOutputMatch: true,
		},
		{
			description:         "All not found",
			args:                []string{"nonexistent", "unknown"},
			nativeExitCode:      1,
			customExitCode:      1,
			expectedOutputMatch: true,
		},
		{
			description:         "Empty input",
			args:                []string{},
			nativeExitCode:      1,
			customExitCode:      1,
			expectedOutputMatch: true,
		},
	}

	for _, tc := range testCases {
		// Test broken out on its own
		if runtime.GOOS == "openbsd" && tc.description == "All not found" {
			continue
		}

		t.Run(tc.description, func(t *testing.T) {

			nativeOutput, nativeCode, err := runCommand(NATIVE_WHICH, tc.args...)
			if err != nil {
				t.Fatalf("Error running native which: %v", err)
			}

			customOutput, customCode, err := runCommand(CUSTOM_WHICH, tc.args...)
			if err != nil {
				t.Fatalf("Error running custom which: %v", err)
			}

			if nativeCode != tc.nativeExitCode {
				t.Errorf("Native exit code mismatch for '%s': got %d, want %d",
					tc.description, nativeCode, tc.nativeExitCode)
			}

			if customCode != tc.customExitCode {
				t.Errorf("Custom exit code mismatch for '%s': got %d, want %d",
					tc.description, customCode, tc.customExitCode)
			}

			if tc.expectedOutputMatch && nativeOutput != customOutput {
				t.Errorf("Output mismatch for '%s':\nNative: %q\nCustom: %q",
					tc.description, nativeOutput, customOutput)
			}
		})
	}
}

func TestWhich_OpenBSD(t *testing.T) {
	if runtime.GOOS != "openbsd" {
		return
	}

	testCases := []struct {
		description         string
		args                []string
		nativeExitCode      int
		customExitCode      int
		expectedOutputMatch bool
	}{
		{
			description:         "All not found",
			args:                []string{"nonexistent", "unknown"},
			nativeExitCode:      2,
			customExitCode:      2,
			expectedOutputMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			nativeOutput, nativeCode, err := runCommand(NATIVE_WHICH, tc.args...)
			if err != nil {
				t.Fatalf("Error running native which: %v", err)
			}

			customOutput, customCode, err := runCommand(CUSTOM_WHICH, tc.args...)
			if err != nil {
				t.Fatalf("Error running custom which: %v", err)
			}

			if nativeCode != tc.nativeExitCode {
				t.Errorf("Native exit code mismatch for '%s': got %d, want %d",
					tc.description, nativeCode, tc.nativeExitCode)
			}

			if customCode != tc.customExitCode {
				t.Errorf("Custom exit code mismatch for '%s': got %d, want %d",
					tc.description, customCode, tc.customExitCode)
			}

			if tc.expectedOutputMatch && nativeOutput != customOutput {
				t.Errorf("Output mismatch for '%s':\nNative: %q\nCustom: %q",
					tc.description, nativeOutput, customOutput)
			}
		})
	}
}

func runCommandWithEnv(cmd string, args []string, env []string) (string, int, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	c := exec.Command(cmd, args...)
	c.Env = append(os.Environ(), env...)
	c.Stdout = &out
	c.Stderr = &stderr
	err := c.Run()

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		return "", 0, err
	}

	return strings.TrimSpace(out.String() + stderr.String()), exitCode, nil
}

func TestWhichEmptyPath(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	testCases := []struct {
		description         string
		args                []string
		temporaryPath       string
		nativeExitCode      int
		customExitCode      int
		expectedOutputMatch bool
	}{
		{
			description:         "Single found program",
			args:                []string{"ls"},
			temporaryPath:       originalPath,
			nativeExitCode:      0,
			customExitCode:      0,
			expectedOutputMatch: true,
		},
		{
			description:         "Empty PATH",
			args:                []string{"ls"},
			temporaryPath:       "",
			nativeExitCode:      1,
			customExitCode:      1,
			expectedOutputMatch: true,
		},
	}

	for _, tc := range testCases {
		if runtime.GOOS == "openbsd" && tc.description == "Empty PATH" {
			fmt.Printf("Skipping 'Empty PATH' test on OpenBSD (not yet configured for KSH)\n")
			continue
		}

		t.Run(tc.description, func(t *testing.T) {
			var env []string

			if tc.temporaryPath == "" {
				os.Unsetenv("PATH")
			} else {
				os.Setenv("PATH", tc.temporaryPath)
			}

			env = append(env, "PATH="+os.Getenv("PATH"))
			nativeOutput, nativeCode, err := runCommandWithEnv(NATIVE_WHICH, tc.args, env)
			if err != nil {
				t.Fatalf("Error running native which: %v", err)
			}

			customOutput, customCode, err := runCommandWithEnv(CUSTOM_WHICH, tc.args, env)
			if err != nil {
				t.Fatalf("Error running custom which: %v", err)
			}

			if nativeCode != tc.nativeExitCode {
				t.Errorf("Native exit code mismatch for '%s': got %d, want %d",
					tc.description, nativeCode, tc.nativeExitCode)
			}

			if customCode != tc.customExitCode {
				t.Errorf("Custom exit code mismatch for '%s': got %d, want %d",
					tc.description, customCode, tc.customExitCode)
			}

			if tc.expectedOutputMatch && nativeOutput != customOutput {
				t.Errorf("Output mismatch for '%s':\nNative: %q\nCustom: %q",
					tc.description, nativeOutput, customOutput)
			}
		})
	}
}

// Because both STDOUT and STDERR are used with an invalid arg error,
// the which command prints STDERR first and then the STDOUT (usage...)
// unlike the other version of this function.
func runCommandInvalidArg(cmd string, args ...string) (string, int, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stdout = &out
	c.Stderr = &stderr
	err := c.Run()

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		return "", 0, err
	}

	// This order is very important
	return stderr.String() + out.String(), exitCode, nil
}

type testCaseArgs struct {
	description          string
	args                 []string
	expectedCustomOutput string
	expectedNativeOutput string
	expectedExitCode     int
}

func runWhichArgTest(t *testing.T, tc testCaseArgs) {
	nativeOutput, nativeExitCode, err := runCommandInvalidArg(NATIVE_WHICH, tc.args...)
	if err != nil {
		t.Fatalf("Error running native which: %v", err)
	}

	if nativeOutput != tc.expectedNativeOutput {
		t.Errorf("Output mismatch for native '%s':\ngot: %q, want %q",
			tc.description, nativeOutput, tc.expectedNativeOutput)
	}

	if nativeExitCode != tc.expectedExitCode {
		t.Errorf("Exit code mismatch for native '%s': got %d, want %d",
			tc.description, nativeExitCode, tc.expectedExitCode)
	}

	customOutput, customExitCode, err := runCommandInvalidArg(CUSTOM_WHICH, tc.args...)
	if err != nil {
		t.Fatalf("Error running custom which: %v", err)
	}

	if customOutput != tc.expectedNativeOutput {
		t.Errorf("Output mismatch for custom '%s':\ngot: %q, want %q",
			tc.description, customOutput, tc.expectedCustomOutput)
	}

	if customExitCode != tc.expectedExitCode {
		t.Errorf("Exit code mismatch for custom '%s': got %d, want %d",
			tc.description, customExitCode, tc.expectedExitCode)
	}
}

func TestWhichInvalidArgDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		return
	}

	testCases := []testCaseArgs{
		{"Invalid Arg", []string{"-z"}, DARWIN_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			DARWIN_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
		{"Invalid Args", []string{"-z", "-y"}, DARWIN_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			DARWIN_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
		{"Invalid Arg and Good Arg", []string{"-a", "-z"}, DARWIN_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			DARWIN_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
		{"Invalid Args and Good Arg", []string{"-z", "-a", "-y"}, DARWIN_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			DARWIN_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
	}

	for _, tc := range testCases {
		runWhichArgTest(t, tc)
	}
}

func TestWhichInvalidArgLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	testCases := []testCaseArgs{
		{"Invalid Arg", []string{"-z"}, LINUX_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			LINUX_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 2},
		{"Invalid Args", []string{"-z", "-y"}, LINUX_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			LINUX_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 2},
		{"Invalid Arg and Good Arg", []string{"-a", "-z"}, LINUX_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			LINUX_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 2},
		{"Invalid Args and Good Arg", []string{"-z", "-a", "-y"}, LINUX_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			LINUX_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 2},
	}

	for _, tc := range testCases {
		runWhichArgTest(t, tc)
	}
}

func TestWhichInvalidArgOpenBSD(t *testing.T) {
	if runtime.GOOS != "openbsd" {
		return
	}

	testCases := []testCaseArgs{
		{"Invalid Arg", []string{"-z"}, OPENBSD_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			OPENBSD_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
		{"Invalid Args", []string{"-z", "-y"}, OPENBSD_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			OPENBSD_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
		{"Invalid Arg and Good Arg", []string{"-a", "-z"}, OPENBSD_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			OPENBSD_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
		{"Invalid Args and Good Arg", []string{"-z", "-a", "-y"}, OPENBSD_EXPECTED_CUSTOM_OUTPUT_INVALID_ARG,
			OPENBSD_EXPECTED_NATIVE_OUTPUT_INVALID_ARG, 1},
	}

	for _, tc := range testCases {
		runWhichArgTest(t, tc)
	}
}
