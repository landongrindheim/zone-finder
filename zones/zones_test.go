package zones

import (
	"testing"
	"time"

	"zone-finder/parser"
)

func TestGetLastNMinutes(t *testing.T) {
	// Create 40 minutes of data
	dataPoints := []parser.HRDataPoint{}
	baseTime := time.Date(2025, 10, 28, 18, 0, 0, 0, time.UTC)

	for i := 0; i < 2400; i++ { // 2400 seconds = 40 minutes
		dataPoints = append(dataPoints, parser.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: 150 + i/100, // HR gradually increases
		})
	}

	// Get last 20 minutes
	result := lastTwentyMinutes(dataPoints)

	// Should have 1201 data points (20 minutes * 60 seconds + last data point)
	if len(result) != 1201 {
		t.Errorf("Expected 1201 data points, got %d", len(result))
	}

	// First point in result should be at 20-minute mark
	expectedFirstTime := baseTime.Add(20 * time.Minute).Add(-1 * time.Second)
	if !result[0].Timestamp.Equal(expectedFirstTime) {
		t.Errorf("First timestamp = %v, want %v", result[0].Timestamp, expectedFirstTime)
	}

	// Last point should be at 40-minute mark
	expectedLastTime := baseTime.Add(40 * time.Minute).Add(-1 * time.Second)
	if !result[len(result)-1].Timestamp.Equal(expectedLastTime) {
		t.Errorf("Last timestamp = %v, want %v", result[len(result)-1].Timestamp, expectedLastTime)
	}
}

func TestCalculateLTHR(t *testing.T) {
	// Create synthetic HR data for a 30-minute workout
	// Simulating a time trial where HR builds and stabilizes
	dataPoints := []parser.HRDataPoint{}
	baseTime := time.Date(2025, 10, 28, 18, 0, 0, 0, time.UTC)

	// First 10 minutes: HR builds from 140 to 165
	for i := 0; i < 600; i++ { // 600 seconds = 10 minutes
		hr := 140 + (i / 24) // gradually increases
		dataPoints = append(dataPoints, parser.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: hr,
		})
	}

	// Last 20 minutes: HR stabilizes around 170-172 (this is our LTHR)
	for i := 600; i < 1800; i++ { // 1200 seconds = 20 minutes
		hr := 170 + (i % 3) // oscillates 170, 171, 172
		dataPoints = append(dataPoints, parser.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: hr,
		})
	}

	lthr, err := CalculateLTHR(dataPoints)
	if err != nil {
		t.Fatalf("CalculateLTHR() error = %v", err)
	}

	// Average of last 20 minutes should be ~171
	if lthr < 170 || lthr > 172 {
		t.Errorf("CalculateLTHR() = %d, want between 170-172", lthr)
	}
}

func TestCalculateLTHR_InsufficientData(t *testing.T) {
	// Less than 20 minutes of data
	dataPoints := []parser.HRDataPoint{
		{Timestamp: time.Now(), HeartRate: 150},
		{Timestamp: time.Now().Add(1 * time.Minute), HeartRate: 155},
	}

	_, err := CalculateLTHR(dataPoints)
	if err == nil {
		t.Error("CalculateLTHR() expected error for insufficient data, got nil")
	}
}

func TestCalculateZones(t *testing.T) {
	tests := []struct {
		name         string
		lthr         int
		wantZone1Max int
		wantZone2Min int
		wantZone2Max int
		wantZone3Min int
		wantZone3Max int
		wantZone4Min int
		wantZone4Max int
		wantZone5Min int
	}{
		{
			name:         "LTHR 172 (David Roche's example)",
			lthr:         172,
			wantZone1Max: 137, // round(172 * 0.80) - 1
			wantZone2Min: 138, // round(172 * 0.80)
			wantZone2Max: 151, // round(172 * 0.88)
			wantZone3Min: 152, // zone2Max + 1
			wantZone3Max: 162, // round(172 * 0.94)
			wantZone4Min: 163, // zone3Max + 1
			wantZone4Max: 172, // LTHR
			wantZone5Min: 173, // LTHR + 1
		},
		{
			name:         "LTHR 160",
			lthr:         160,
			wantZone1Max: 127, // round(160 * 0.80) - 1 = 128 - 1
			wantZone2Min: 128, // round(160 * 0.80)
			wantZone2Max: 141, // round(160 * 0.88) = 140.8 rounds to 141
			wantZone3Min: 142, // zone2Max + 1
			wantZone3Max: 150, // round(160 * 0.94) = 150.4 rounds to 150
			wantZone4Min: 151, // zone3Max + 1
			wantZone4Max: 160, // LTHR
			wantZone5Min: 161, // LTHR + 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateZones(tt.lthr)

			// Check LTHR is stored
			if result.LTHR != tt.lthr {
				t.Errorf("LTHR = %d, want %d", result.LTHR, tt.lthr)
			}

			// Check we have 5 zones
			if len(result.Zones) != 5 {
				t.Fatalf("Expected 5 zones, got %d", len(result.Zones))
			}

			// Validate Zone 1
			z1 := result.Zones[0]
			if z1.Number != 1 {
				t.Errorf("Zone 1 Number = %d, want 1", z1.Number)
			}
			if z1.Max != tt.wantZone1Max {
				t.Errorf("Zone 1 Max = %d, want %d", z1.Max, tt.wantZone1Max)
			}

			// Validate Zone 2
			z2 := result.Zones[1]
			if z2.Number != 2 {
				t.Errorf("Zone 2 Number = %d, want 2", z2.Number)
			}
			if z2.Min != tt.wantZone2Min {
				t.Errorf("Zone 2 Min = %d, want %d", z2.Min, tt.wantZone2Min)
			}
			if z2.Max != tt.wantZone2Max {
				t.Errorf("Zone 2 Max = %d, want %d", z2.Max, tt.wantZone2Max)
			}

			// Validate Zone 3
			z3 := result.Zones[2]
			if z3.Number != 3 {
				t.Errorf("Zone 3 Number = %d, want 3", z3.Number)
			}
			if z3.Min != tt.wantZone3Min {
				t.Errorf("Zone 3 Min = %d, want %d", z3.Min, tt.wantZone3Min)
			}
			if z3.Max != tt.wantZone3Max {
				t.Errorf("Zone 3 Max = %d, want %d", z3.Max, tt.wantZone3Max)
			}

			// Validate Zone 4
			z4 := result.Zones[3]
			if z4.Number != 4 {
				t.Errorf("Zone 4 Number = %d, want 4", z4.Number)
			}
			if z4.Min != tt.wantZone4Min {
				t.Errorf("Zone 4 Min = %d, want %d", z4.Min, tt.wantZone4Min)
			}
			if z4.Max != tt.wantZone4Max {
				t.Errorf("Zone 4 Max = %d, want %d", z4.Max, tt.wantZone4Max)
			}

			// Validate Zone 5
			z5 := result.Zones[4]
			if z5.Number != 5 {
				t.Errorf("Zone 5 Number = %d, want 5", z5.Number)
			}
			if z5.Min != tt.wantZone5Min {
				t.Errorf("Zone 5 Min = %d, want %d", z5.Min, tt.wantZone5Min)
			}
		})
	}
}

func TestCalculateZonesFromData(t *testing.T) {
	// Integration test: from data points to zones
	dataPoints := []parser.HRDataPoint{}
	baseTime := time.Date(2025, 10, 28, 18, 0, 0, 0, time.UTC)

	// Create 30 minutes of data with LTHR around 170
	for i := 0; i < 600; i++ {
		hr := 140 + (i / 24)
		dataPoints = append(dataPoints, parser.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: hr,
		})
	}

	for i := 600; i < 1800; i++ {
		hr := 170 + (i % 3)
		dataPoints = append(dataPoints, parser.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: hr,
		})
	}

	result, err := CalculateZonesFromHRData(dataPoints)
	if err != nil {
		t.Fatalf("CalculateZonesFromHRData() error = %v", err)
	}

	// Should have calculated LTHR and zones
	if result.LTHR < 170 || result.LTHR > 172 {
		t.Errorf("LTHR = %d, want between 170-172", result.LTHR)
	}

	if len(result.Zones) != 5 {
		t.Errorf("Expected 5 zones, got %d", len(result.Zones))
	}

	// Zone 2 max should be around 88% of 171 = ~150
	z2 := result.Zones[1]
	if z2.Max < 149 || z2.Max > 151 {
		t.Errorf("Zone 2 Max = %d, want around 150", z2.Max)
	}
}
