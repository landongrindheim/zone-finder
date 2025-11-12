package zones

import (
	"testing"
	"time"
	"zone-finder/types"
)

// Test helpers moved to top
func createHRData(startTime time.Time, hrValues []int) []types.HRDataPoint {
	dataPoints := make([]types.HRDataPoint, len(hrValues))
	for i, hr := range hrValues {
		dataPoints[i] = types.HRDataPoint{
			Timestamp: startTime.Add(time.Duration(i) * time.Second),
			HeartRate: hr,
		}
	}
	return dataPoints
}

func createConstantHR(startTime time.Time, hr int, seconds int) []types.HRDataPoint {
	hrValues := make([]int, seconds)
	for i := range hrValues {
		hrValues[i] = hr
	}
	return createHRData(startTime, hrValues)
}

func TestLastTwentyMinutes(t *testing.T) {
	baseTime := time.Date(2025, 10, 28, 18, 0, 0, 0, time.UTC)

	// Create 40 minutes of gradually increasing HR
	dataPoints := make([]types.HRDataPoint, 2400) // Pre-allocate
	for i := 0; i < 2400; i++ {
		dataPoints[i] = types.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: 150 + i/100,
		}
	}

	result := lastTwentyMinutes(dataPoints)

	if len(result) != 1201 {
		t.Errorf("got %d data points, want 1201", len(result))
	}

	expectedFirstTime := baseTime.Add(19*time.Minute + 59*time.Second)
	if !result[0].Timestamp.Equal(expectedFirstTime) {
		t.Errorf("first timestamp = %v, want %v", result[0].Timestamp, expectedFirstTime)
	}

	expectedLastTime := baseTime.Add(39*time.Minute + 59*time.Second)
	if !result[len(result)-1].Timestamp.Equal(expectedLastTime) {
		t.Errorf("last timestamp = %v, want %v", result[len(result)-1].Timestamp, expectedLastTime)
	}
}

func TestCalculateLTHR(t *testing.T) {
	baseTime := time.Date(2025, 10, 28, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		data    []types.HRDataPoint
		wantMin int
		wantMax int
		wantErr bool
	}{
		{
			name: "30-minute workout with stable LTHR",
			data: func() []types.HRDataPoint {
				var data []types.HRDataPoint
				// 10 min warmup: 140-165
				for i := 0; i < 600; i++ {
					data = append(data, types.HRDataPoint{
						Timestamp: baseTime.Add(time.Duration(i) * time.Second),
						HeartRate: 140 + (i / 24),
					})
				}
				// 20 min at 170-172
				for i := 600; i < 1800; i++ {
					data = append(data, types.HRDataPoint{
						Timestamp: baseTime.Add(time.Duration(i) * time.Second),
						HeartRate: 170 + (i % 3),
					})
				}
				return data
			}(),
			wantMin: 170,
			wantMax: 172,
			wantErr: false,
		},
		{
			name: "insufficient data",
			data: []types.HRDataPoint{
				{Timestamp: time.Now(), HeartRate: 150},
				{Timestamp: time.Now().Add(1 * time.Minute), HeartRate: 155},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lthr, err := CalculateLTHR(tt.data)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CalculateLTHR() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if lthr < tt.wantMin || lthr > tt.wantMax {
				t.Errorf("CalculateLTHR() = %d, want between %d-%d", lthr, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateZones(t *testing.T) {
	tests := []struct {
		name  string
		lthr  int
		zones [5]Zone
	}{
		{
			name: "LTHR 172",
			lthr: 172,
			zones: [5]Zone{
				{Number: 1, Min: 0, Max: 137},
				{Number: 2, Min: 138, Max: 151},
				{Number: 3, Min: 152, Max: 162},
				{Number: 4, Min: 163, Max: 172},
				{Number: 5, Min: 173, Max: 220},
			},
		},
		{
			name: "LTHR 160",
			lthr: 160,
			zones: [5]Zone{
				{Number: 1, Min: 0, Max: 127},
				{Number: 2, Min: 128, Max: 141},
				{Number: 3, Min: 142, Max: 150},
				{Number: 4, Min: 151, Max: 160},
				{Number: 5, Min: 161, Max: 220},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateZones(tt.lthr)

			if result.LTHR != tt.lthr {
				t.Errorf("LTHR = %d, want %d", result.LTHR, tt.lthr)
			}

			for i, want := range tt.zones {
				got := result.Zones[i]
				if got != want {
					t.Errorf("Zone %d = %+v, want %+v", i+1, got, want)
				}
			}
		})
	}
}

func TestCalculateZonesFromHRData(t *testing.T) {
	baseTime := time.Date(2025, 10, 28, 18, 0, 0, 0, time.UTC)

	var dataPoints []types.HRDataPoint
	// 10 min warmup
	for i := 0; i < 600; i++ {
		dataPoints = append(dataPoints, types.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: 140 + (i / 24),
		})
	}
	// 20 min at stable HR
	for i := 600; i < 1800; i++ {
		dataPoints = append(dataPoints, types.HRDataPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			HeartRate: 170 + (i % 3),
		})
	}

	result, err := CalculateZonesFromHRData(dataPoints)
	if err != nil {
		t.Fatalf("CalculateZonesFromHRData() error = %v", err)
	}

	if result.LTHR < 170 || result.LTHR > 172 {
		t.Errorf("LTHR = %d, want between 170-172", result.LTHR)
	}

	if len(result.Zones) != 5 {
		t.Fatalf("got %d zones, want 5", len(result.Zones))
	}

	// Verify zone 2 calculation
	z2 := result.Zones[1]
	if z2.Max < 149 || z2.Max > 151 {
		t.Errorf("Zone 2 Max = %d, want ~150", z2.Max)
	}
}

func TestFindBestWindow(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		dataSetup func() []types.HRDataPoint
		wantLTHR  int
		wantErr   bool
	}{
		{
			name: "workout with cooldown",
			dataSetup: func() []types.HRDataPoint {
				var data []types.HRDataPoint
				data = append(data, createConstantHR(baseTime, 120, 5*60)...)                     // warmup
				data = append(data, createConstantHR(baseTime.Add(5*time.Minute), 170, 30*60)...) // effort
				data = append(data, createConstantHR(baseTime.Add(35*time.Minute), 110, 5*60)...) // cooldown
				return data
			},
			wantLTHR: 170,
			wantErr:  false,
		},
		{
			name: "workout with warmup",
			dataSetup: func() []types.HRDataPoint {
				var data []types.HRDataPoint
				data = append(data, createConstantHR(baseTime, 100, 10*60)...)
				data = append(data, createConstantHR(baseTime.Add(10*time.Minute), 165, 25*60)...)
				return data
			},
			wantLTHR: 165,
			wantErr:  false,
		},
		{
			name: "steady effort",
			dataSetup: func() []types.HRDataPoint {
				return createConstantHR(baseTime, 168, 35*60)
			},
			wantLTHR: 168,
			wantErr:  false,
		},
		{
			name: "workout too short",
			dataSetup: func() []types.HRDataPoint {
				return createConstantHR(baseTime, 160, 19*60)
			},
			wantErr: true,
		},
		{
			name: "progressive build",
			dataSetup: func() []types.HRDataPoint {
				var data []types.HRDataPoint
				for min := 0; min < 35; min++ {
					hr := 140 + min
					data = append(data, createConstantHR(
						baseTime.Add(time.Duration(min)*time.Minute),
						hr,
						60,
					)...)
				}
				return data
			},
			wantLTHR: 164,
			wantErr:  false,
		},
		{
			name: "interval workout",
			dataSetup: func() []types.HRDataPoint {
				var data []types.HRDataPoint
				data = append(data, createConstantHR(baseTime, 120, 10*60)...)

				// 5 intervals
				for i := 0; i < 5; i++ {
					offset := baseTime.Add(time.Duration(10+i*5) * time.Minute)
					data = append(data, createConstantHR(offset, 180, 3*60)...)
					data = append(data, createConstantHR(offset.Add(3*time.Minute), 140, 2*60)...)
				}

				data = append(data, createConstantHR(baseTime.Add(35*time.Minute), 110, 5*60)...)
				return data
			},
			wantLTHR: 166,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window, err := FindBestWindow(tt.dataSetup())

			if (err != nil) != tt.wantErr {
				t.Errorf("FindBestWindow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			lthr, err := CalculateLTHR(window)
			if err != nil {
				t.Fatalf("CalculateLTHR() error = %v", err)
			}

			if lthr < tt.wantLTHR-2 || lthr > tt.wantLTHR+2 {
				t.Errorf("LTHR = %d, want ~%d (±2)", lthr, tt.wantLTHR)
			}

			// Verify window duration
			duration := window[len(window)-1].Timestamp.Sub(window[0].Timestamp)
			minDuration := 20*time.Minute - 2*time.Second
			if duration < minDuration {
				t.Errorf("window duration = %v, want >= %v", duration, minDuration)
			}
		})
	}
}
