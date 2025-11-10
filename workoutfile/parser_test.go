package workoutfile

import (
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid TCX file",
			path:    "../tcx/testdata/treadmill_run_watch.tcx",
			wantErr: false,
		},
		{
			name:    "valid FIT file",
			path:    "../fit/testdata/treadmill_run_watch.fit",
			wantErr: false,
		},
		{
			name:    "uppercase TCX extension",
			path:    "../tcx/testdata/treadmill_run_watch.TCX",
			wantErr: false, // Should handle case-insensitive
		},
		{
			name:    "unsupported format - GPX",
			path:    "workout.gpx",
			wantErr: true,
			errMsg:  "unsupported",
		},
		{
			name:    "unsupported format - no extension",
			path:    "workout",
			wantErr: true,
			errMsg:  "unsupported",
		},
		{
			name:    "non-existent file",
			path:    "does-not-exist.tcx",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutFile, err := ParseFile(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				// Check error message if specified
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want substring %q", err.Error(), tt.errMsg)
				}
				return
			}

			// For successful parses, verify interface is satisfied
			if workoutFile == nil {
				t.Error("ParseFile() returned nil WorkoutFile")
			}
		})
	}
}

func TestParseFile_ExtractsHRData(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantData bool
	}{
		{
			name:     "TCX file has HR data",
			path:     "../tcx/testdata/treadmill_run_watch.tcx",
			wantData: true,
		},
		{
			name:     "FIT file has HR data",
			path:     "../fit/testdata/treadmill_run_watch.fit",
			wantData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutFile, err := ParseFile(tt.path)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			dataPoints, err := workoutFile.GetHRDataPoints()
			if err != nil {
				t.Fatalf("GetHRDataPoints() error = %v", err)
			}

			if tt.wantData && len(dataPoints) == 0 {
				t.Error("Expected HR data points, got none")
			}

			// Verify first data point is reasonable
			if len(dataPoints) > 0 {
				first := dataPoints[0]
				if first.HeartRate < 20 || first.HeartRate > 220 {
					t.Errorf("First HR %d outside reasonable range", first.HeartRate)
				}
			}
		})
	}
}

func TestParseFile_ExtractsDeviceInfo(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "TCX file",
			path: "../tcx/testdata/treadmill_run_watch.tcx",
		},
		{
			name: "FIT file",
			path: "../fit/testdata/treadmill_run_watch.fit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutFile, err := ParseFile(tt.path)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			deviceName := workoutFile.GetDeviceName()
			if deviceName == "" {
				t.Error("Expected device name, got empty string")
			}

			productID := workoutFile.GetProductID()
			if productID == 0 {
				t.Log("Product ID is 0, might be expected for some devices")
			}
		})
	}
}
