package zones

import (
	"errors"
	"math"
	"sort"
	"time"
	"zone-finder/workoutfile"
)

type HeartRateZones struct {
	LTHR  int
	Zones [5]Zone
}

type Zone struct {
	Number int
	Min    int
	Max    int
}

const (
	maxHeartRate         = 220
	zone2Lower   float64 = 0.80
	zone2Upper   float64 = 0.88
	zone3Upper   float64 = 0.94
)

func CalculateZonesFromHRData(dataPoints []workoutfile.HRDataPoint) (HeartRateZones, error) {
	sorted := sortByTimestamp(dataPoints)

	lactateThreshold, err := CalculateLTHR(sorted)
	if err != nil {
		return HeartRateZones{}, err
	}

	zones := CalculateZones(lactateThreshold)

	return zones, nil
}

func sortByTimestamp(dataPoints []workoutfile.HRDataPoint) []workoutfile.HRDataPoint {
	sort.Slice(dataPoints, func(i, j int) bool { return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp) })

	return dataPoints
}

func CalculateLTHR(dataPoints []workoutfile.HRDataPoint) (int, error) {
	totalDuration := dataPoints[len(dataPoints)-1].Timestamp.Sub(dataPoints[0].Timestamp)
	if totalDuration < 20*time.Minute {
		return 0, errors.New("insufficient data: need at least 20 minutes")
	}

	lastTwentyMinutes := lastTwentyMinutes(dataPoints)

	sum := 0
	for _, dp := range lastTwentyMinutes {
		sum += dp.HeartRate
	}

	lactateThreshold := sum / len(lastTwentyMinutes)
	return lactateThreshold, nil
}

func lastTwentyMinutes(dataPoints []workoutfile.HRDataPoint) []workoutfile.HRDataPoint {
	var lastdataPoints []workoutfile.HRDataPoint

	lastTimestamp := dataPoints[len(dataPoints)-1].Timestamp
	twentyMinutesPrior := lastTimestamp.Add(-20 * time.Minute)

	for _, dp := range dataPoints {
		if !dp.Timestamp.Before(twentyMinutesPrior) {
			lastdataPoints = append(lastdataPoints, dp)
		}
	}

	return lastdataPoints
}

func CalculateZones(lthr int) HeartRateZones {
	z2Lower := calculateZoneBoundary(lthr, zone2Lower)
	z2Upper := calculateZoneBoundary(lthr, zone2Upper)
	z3Upper := calculateZoneBoundary(lthr, zone3Upper)
	z4Upper := lthr

	return HeartRateZones{
		LTHR: lthr,
		Zones: [5]Zone{
			{Number: 1, Min: 0, Max: z2Lower - 1},
			{Number: 2, Min: z2Lower, Max: z2Upper},
			{Number: 3, Min: z2Upper + 1, Max: z3Upper},
			{Number: 4, Min: z3Upper + 1, Max: z4Upper},
			{Number: 5, Min: z4Upper + 1, Max: maxHeartRate},
		},
	}
}

func calculateZoneBoundary(lthr int, percentage float64) int {
	return int(math.Round(float64(lthr) * percentage))
}
