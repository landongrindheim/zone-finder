package main

import (
	"bytes"
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
		t.Error("Output should contain LTHR value")
	}

	// Check for zone information
	if !strings.Contains(output, "Zone 1") {
		t.Error("Output should contain Zone 1")
	}

	if !strings.Contains(output, "0-137") {
		t.Error("Output should contain Zone 1 range")
	}

	// Check for all 5 zones
	for i := 1; i <= 5; i++ {
		zoneName := "Zone " + string(rune('0'+i))
		if !strings.Contains(output, zoneName) {
			t.Errorf("Output should contain %s", zoneName)
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
		t.Error("Output should include 'bpm' units")
	}
}

func TestValidateArgs(t *testing.T) {
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

func TestRun_InvalidFile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"zone-finder", "nonexistent.tcx"}, &stdout, &stderr)

	// Should exit with error code
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid file")
	}

	// Should write error to stderr
	if stderr.Len() == 0 {
		t.Error("Expected error message in stderr")
	}

	// Should not write to stdout
	if stdout.Len() > 0 {
		t.Error("Expected no output to stdout on error")
	}
}

func TestRun_ValidFile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// This test would use a real test file
	// For now, we're just defining the expected behavior
	exitCode := run([]string{"zone-finder", "../parser/testdata/outside_run_armband.tcx"}, &stdout, &stderr)

	// Should exit successfully
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Should write output to stdout
	if stdout.Len() == 0 {
		t.Error("Expected output to stdout")
	}

	// Should contain zone information
	output := stdout.String()
	if !strings.Contains(output, "LTHR") {
		t.Error("Output should contain LTHR")
	}

	if !strings.Contains(output, "Zone") {
		t.Error("Output should contain zone information")
	}

	// Should not write errors to stderr
	if stderr.Len() > 0 {
		t.Errorf("Expected no errors, got: %s", stderr.String())
	}
}
