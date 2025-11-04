package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"zone-finder/zones"
)

func TestFormatOutput(t *testing.T) {
	// Mock zone result
	result := zones.HeartRateZones{
		LTHR: 172,
		Zones: [5]zones.Zone{
			{Number: 1, Min: 0, Max: 137},
			{Number: 2, Min: 138, Max: 151},
			{Number: 3, Min: 152, Max: 162},
			{Number: 4, Min: 163, Max: 172},
			{Number: 5, Min: 173, Max: 220},
		},
	}

	output := formatOutput(result)

	// Check for LTHR in output
	if !strings.Contains(output, "LTHR: 172") {
		t.Error("Expected output to contain LTHR value")
	}

	// Check for zone information
	if !strings.Contains(output, "Zone 1") {
		t.Error("Expected output to contain Zone 1")
	}

	if !strings.Contains(output, "0-137") {
		t.Error("Expected output to contain Zone 1 range")
	}

	// Check for all 5 zones
	for i := 1; i <= 5; i++ {
		zoneName := fmt.Sprintf("Zone %d", i)
		if !strings.Contains(output, zoneName) {
			t.Errorf("Expected output to contain %s", zoneName)
		}
	}
}

func TestFormatOutput_Structure(t *testing.T) {
	result := zones.HeartRateZones{
		LTHR: 160,
		Zones: [5]zones.Zone{
			{Number: 1, Min: 0, Max: 127},
			{Number: 2, Min: 128, Max: 141},
			{Number: 3, Min: 142, Max: 150},
			{Number: 4, Min: 151, Max: 160},
			{Number: 5, Min: 161, Max: 220},
		},
	}

	output := formatOutput(result)

	// Output should be multi-line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 6 { // LTHR line + 5 zone lines minimum
		t.Errorf("Expected at least 6 lines of output, got %d", len(lines))
	}

	// Check LTHR formatting includes "bpm"
	if !strings.Contains(output, "bpm") {
		t.Error("Expected output to include 'bpm' units")
	}
}

func TestValidateArgs_ValidatesArgumentCount(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid single file argument",
			args:    []string{"program", "file.tcx"},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []string{"program"},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"program", "file1.tcx", "file2.tcx"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStdout   bool
		wantStderr   bool
	}{
		{
			name:         "valid file",
			args:         []string{"zone-finder", "../parser/testdata/outside_run_armband.tcx"},
			wantExitCode: 0,
			wantStdout:   true,
			wantStderr:   false,
		},
		{
			name:         "invalid file",
			args:         []string{"zone-finder", "nonexistent.tcx"},
			wantExitCode: 1,
			wantStdout:   false,
			wantStderr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exitCode := run(tt.args, &stdout, &stderr)

			// Should exit with correct code
			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %v, got: %v", tt.wantExitCode, exitCode)
			}

			if tt.wantStdout && stdout.Len() == 0 {
				t.Error("Expected message in stdout")
			}

			if !tt.wantStdout && stdout.Len() > 0 {
				t.Errorf("Expected no output to stdout, got %s", stdout.String())
			}

			if tt.wantStderr && stderr.Len() == 0 {
				t.Error("Expected error message in stderr")
			}

			if !tt.wantStderr && stderr.Len() > 0 {
				t.Errorf("Expected no stderr, got %s", stderr.String())
			}

			if tt.wantExitCode == 0 {
				output := stdout.String()
				if !strings.Contains(output, "LTHR") {
					t.Error("Expected output to contain LTHR")
				}

				if !strings.Contains(output, "Zone") {
					t.Error("Expected output to contain zone information")
				}
			}
		})
	}
}

func TestShowUsage(t *testing.T) {
	var stdout bytes.Buffer

	showUsage(&stdout)

	output := stdout.String()

	// Should contain program name
	if !strings.Contains(output, "zone-finder") {
		t.Error("Expected usage to contain program name")
	}

	// Should explain what arguments to provide
	if !strings.Contains(output, "file") || !strings.Contains(output, ".tcx") {
		t.Error("Expected usage to mention TCX file argument")
	}

	// Should contain "Usage:" header
	if !strings.Contains(output, "Usage:") {
		t.Error("Expected usage to contain 'Usage:' header")
	}

	// Should have examples section (optional but nice)
	if !strings.Contains(output, "Example") {
		t.Log("Consider adding examples to usage message")
	}
}

func TestRun_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "short help flag",
			args: []string{"zone-finder", "-h"},
		},
		{
			name: "long help flag",
			args: []string{"zone-finder", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := run(tt.args, &stdout, &stderr)

			// Should exit successfully (help is not an error)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0 for help, got %d", exitCode)
			}

			// Should write usage to stdout
			if stdout.Len() == 0 {
				t.Error("Expected usage message in stdout")
			}

			// Should contain usage information
			output := stdout.String()
			if !strings.Contains(output, "Usage:") {
				t.Error("Help output expected usage information")
			}

			// Should not write to stderr (help is not an error)
			if stderr.Len() > 0 {
				t.Errorf("Expected no error output for help, got: %s", stderr.String())
			}
		})
	}
}

func TestRun_NoArgs_ShowsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"zone-finder"}, &stdout, &stderr)

	// Should exit with error code (missing required argument)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when no file provided")
	}

	// Should show usage message to help user
	output := stderr.String()
	if !strings.Contains(output, "Usage:") && !strings.Contains(output, "zone-finder") {
		t.Error("Error output expected to include usage hint")
	}
}

func TestValidateArgs_HelpFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		isHelpFlag bool
	}{
		{
			name:       "help short flag",
			args:       []string{"program", "-h"},
			isHelpFlag: true,
		},
		{
			name:       "help long flag",
			args:       []string{"program", "--help"},
			isHelpFlag: true,
		},
		{
			name:       "regular file",
			args:       []string{"program", "file.tcx"},
			isHelpFlag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isHelp := checkHelpFlag(tt.args)

			if isHelp != tt.isHelpFlag {
				t.Errorf("checkHelpFlag() isHelp = %v, want %v", isHelp, tt.isHelpFlag)
			}
		})
	}
}
