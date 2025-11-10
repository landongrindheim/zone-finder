package tcx

import (
	"testing"
	"time"
)

func TestParseTCX(t *testing.T) {
	tests := []struct {
		name        string
		filepath    string
		wantErr     bool
		wantSport   string
		wantDevice  string
		wantProduct int
	}{
		{
			name:        "valid garmin watch file",
			filepath:    "testdata/treadmill_run_watch.tcx",
			wantErr:     false,
			wantSport:   "Running",
			wantDevice:  "Forerunner 265",
			wantProduct: 4257,
		},
		{
			name:     "non-existent file",
			filepath: "testdata/does_not_exist.tcx",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcx, err := ParseTCX(tt.filepath)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTCX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected an error, we're done
			if tt.wantErr {
				return
			}

			// Validate sport
			if got := tcx.Activities.Activity.Sport; got != tt.wantSport {
				t.Errorf("Sport = %v, want %v", got, tt.wantSport)
			}

			// Validate device name
			if got := tcx.GetDeviceName(); got != tt.wantDevice {
				t.Errorf("DeviceName = %v, want %v", got, tt.wantDevice)
			}

			// Validate product ID
			if got := tcx.GetProductID(); got != tt.wantProduct {
				t.Errorf("ProductID = %v, want %v", got, tt.wantProduct)
			}
		})
	}
}

func TestGetHRDataPoints(t *testing.T) {
	tcx, err := ParseTCX("testdata/treadmill_run_watch.tcx")
	if err != nil {
		t.Fatalf("Failed to parse TCX file: %v", err)
	}

	dataPoints, err := tcx.GetHRDataPoints()
	if err != nil {
		t.Fatalf("GetHRDataPoints() error = %v", err)
	}

	// Basic validation
	if len(dataPoints) == 0 {
		t.Error("Expected heart rate data points, got none")
	}

	// Validate first data point structure
	if len(dataPoints) > 0 {
		first := dataPoints[0]

		// Check timestamp is valid
		if first.Timestamp.IsZero() {
			t.Error("First data point has zero timestamp")
		}

		// Check heart rate is reasonable (between 30 and 220)
		if first.HeartRate < 30 || first.HeartRate > 220 {
			t.Errorf("First HR value %d is outside reasonable range", first.HeartRate)
		}
	}

	// Check timestamps are in order
	for i := 1; i < len(dataPoints); i++ {
		if dataPoints[i].Timestamp.Before(dataPoints[i-1].Timestamp) {
			t.Errorf("Timestamps out of order at index %d", i)
		}
	}
}

func TestGetHRDataPoints_Specific(t *testing.T) {
	tcx, err := ParseTCX("testdata/treadmill_run_watch.tcx")
	if err != nil {
		t.Fatalf("Failed to parse TCX file: %v", err)
	}

	dataPoints, err := tcx.GetHRDataPoints()
	if err != nil {
		t.Fatalf("GetHRDataPoints() error = %v", err)
	}

	// Test specific known values from the file
	// First HR reading appears at 2025-10-28T18:42:51.000Z with value 110
	expectedTime, _ := time.Parse(time.RFC3339, "2025-10-28T18:42:51.000Z")
	expectedHR := 110

	found := false
	for _, dp := range dataPoints {
		if dp.Timestamp.Equal(expectedTime) {
			found = true
			if dp.HeartRate != expectedHR {
				t.Errorf("At time %v, expected HR %d, got %d",
					expectedTime, expectedHR, dp.HeartRate)
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected to find data point at time %v", expectedTime)
	}
}

func TestGetDeviceInfo(t *testing.T) {
	tcx, err := ParseTCX("testdata/treadmill_run_watch.tcx")
	if err != nil {
		t.Fatalf("Failed to parse TCX file: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{
			name: "device name",
			got:  tcx.GetDeviceName(),
			want: "Forerunner 265",
		},
		{
			name: "product ID",
			got:  tcx.GetProductID(),
			want: 4257,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// Benchmark parsing performance
func BenchmarkParseTCX(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseTCX("testdata/treadmill_run_watch.tcx")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark HR data extraction
func BenchmarkGetHRDataPoints(b *testing.B) {
	tcx, err := ParseTCX("testdata/treadmill_run_watch.tcx")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tcx.GetHRDataPoints()
		if err != nil {
			b.Fatal(err)
		}
	}
}
