package fit

import (
	"testing"
)

func TestParseFIT(t *testing.T) {
	tests := []struct {
		name        string
		filepath    string
		wantErr     bool
		wantSport   string
		wantDevice  string
		wantProduct int
	}{
		{
			name:     "valid garmin watch file",
			filepath: "testdata/treadmill_run_watch.fit",
			wantErr:  false,
			// Device info will depend on what's in the FIT file
			// We'll validate these after examining the actual data
		},
		{
			name:     "valid armband file",
			filepath: "testdata/outside_run_armband.fit",
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filepath: "testdata/does_not_exist.fit",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fitData, err := ParseFIT(tt.filepath)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFIT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we're done
			if tt.wantErr {
				return
			}

			// Validate we got data back
			if fitData == nil {
				t.Error("ParseFIT() returned nil data")
			}

			// TODO: Add sport/device validation once we know what's in the files
			// This will be filled in after examining the FIT file contents
		})
	}
}

func TestGetHRDataPoints(t *testing.T) {
	fitData, err := ParseFIT("testdata/treadmill_run_watch.fit")
	if err != nil {
		t.Fatalf("failed to parse FIT file: %v", err)
	}

	dataPoints, err := fitData.GetHRDataPoints()
	if err != nil {
		t.Fatalf("GetHRDataPoints() error = %v", err)
	}

	// Basic validation
	if len(dataPoints) == 0 {
		t.Error("expected heart rate data points, got none")
	}

	// Validate first data point structure
	if len(dataPoints) > 0 {
		first := dataPoints[0]

		// Check timestamp is valid
		if first.Timestamp.IsZero() {
			t.Error("first data point has zero timestamp")
		}

		// Check heart rate is reasonable (between 30 and 220)
		if first.HeartRate < 20 || first.HeartRate > 240 {
			t.Errorf("first HR value %d is outside reasonable range", first.HeartRate)
		}
	}
}

func TestGetHRDataPoints_MultipleFiles(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		wantData bool
	}{
		{
			name:     "treadmill run watch",
			filepath: "testdata/treadmill_run_watch.fit",
			wantData: true,
		},
		{
			name:     "outside run armband",
			filepath: "testdata/outside_run_armband.fit",
			wantData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fitData, err := ParseFIT(tt.filepath)
			if err != nil {
				t.Fatalf("ParseFIT() error = %v", err)
			}

			dataPoints, err := fitData.GetHRDataPoints()
			if err != nil {
				t.Fatalf("GetHRDataPoints() error = %v", err)
			}

			if tt.wantData && len(dataPoints) == 0 {
				t.Error("expected HR data points, got none")
			}

			// Validate data quality
			if len(dataPoints) > 0 {
				// Check for reasonable HR values throughout
				for i, dp := range dataPoints {
					if dp.HeartRate < 20 || dp.HeartRate > 240 {
						t.Errorf("data point %d: HR %d outside reasonable range", i, dp.HeartRate)
					}
				}
			}
		})
	}
}

func TestGetDeviceInfo(t *testing.T) {
	fitData, err := ParseFIT("testdata/treadmill_run_watch.fit")
	if err != nil {
		t.Fatalf("failed to parse FIT file: %v", err)
	}

	// These methods should exist to match TCX parser interface
	deviceName := fitData.GetDeviceName()
	if deviceName == "" {
		t.Error("expected device name, got empty string")
	}

	productID := fitData.GetProductID()
	// Product ID should be a reasonable value
	// FIT files use manufacturer-specific product IDs
	if productID == 0 {
		t.Log("Product ID is 0, might be expected for some devices")
	}
}

func TestParseFIT_ValidatesFileFormat(t *testing.T) {
	// Try parsing a TCX file as FIT - should fail gracefully
	_, err := ParseFIT("../tcx/testdata/treadmill_run_watch.tcx")
	if err == nil {
		t.Error("expected error when parsing TCX file as FIT")
	}
}

// Benchmark parsing performance
func BenchmarkParseFIT(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseFIT("testdata/treadmill_run_watch.fit")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark HR data extraction
func BenchmarkGetHRDataPoints(b *testing.B) {
	fitData, err := ParseFIT("testdata/treadmill_run_watch.fit")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fitData.GetHRDataPoints()
		if err != nil {
			b.Fatal(err)
		}
	}
}
