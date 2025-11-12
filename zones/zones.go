package zones

import (
	"errors"
	"math"
	"sort"
	"time"
	"zone-finder/types"
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
	windowDuration                = 20 * time.Minute
	minAcceptableDuration         = (windowDuration - 2*time.Second)
	maxHeartRate                  = 220
	zone2Lower            float64 = 0.80
	zone2Upper            float64 = 0.88
	zone3Upper            float64 = 0.94
)

func CalculateZonesFromHRData(dataPoints []types.HRDataPoint) (HeartRateZones, error) {
	sorted := sortByTimestamp(dataPoints)

	lactateThreshold, err := CalculateLTHR(sorted)
	if err != nil {
		return HeartRateZones{}, err
	}

	zones := CalculateZones(lactateThreshold)

	return zones, nil
}

func sortByTimestamp(dataPoints []types.HRDataPoint) []types.HRDataPoint {
	sort.Slice(dataPoints, func(i, j int) bool { return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp) })

	return dataPoints
}

func CalculateLTHR(dataPoints []types.HRDataPoint) (int, error) {
	totalDuration := dataPoints[len(dataPoints)-1].Timestamp.Sub(dataPoints[0].Timestamp)
	if totalDuration < minAcceptableDuration {
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

func lastTwentyMinutes(dataPoints []types.HRDataPoint) []types.HRDataPoint {
	var lastDataPoints []types.HRDataPoint

	lastTimestamp := dataPoints[len(dataPoints)-1].Timestamp
	twentyMinutesPrior := lastTimestamp.Add(-20 * time.Minute)

	for _, dp := range dataPoints {
		if !dp.Timestamp.Before(twentyMinutesPrior) {
			lastDataPoints = append(lastDataPoints, dp)
		}
	}

	return lastDataPoints
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

// Finds the 20-minute window with the highest average heart rate
func FindBestWindow(dataPoints []types.HRDataPoint) ([]types.HRDataPoint, error) {
	dataPoints = sortByTimestamp(dataPoints)

	var bestWindow []types.HRDataPoint
	var bestAvg float64

	if len(dataPoints) == 0 {
		return nil, errors.New("No HR data provided")
	}

	workoutDuration := dataPoints[len(dataPoints)-1].Timestamp.Sub(dataPoints[0].Timestamp)
	if workoutDuration < windowDuration {
		return nil, errors.New("workout too short")
	}

	for i := 0; i < len(dataPoints); i++ {
		startTime := dataPoints[i].Timestamp
		endTime := startTime.Add(windowDuration)

		var window []types.HRDataPoint
		for j := i; j < len(dataPoints); j++ {
			if !dataPoints[j].Timestamp.After(endTime) {
				window = append(window, dataPoints[j])
			} else {
				break
			}
		}

		if len(window) == 0 {
			continue
		}

		actualDuration := window[len(window)-1].Timestamp.Sub(window[0].Timestamp)
		if actualDuration < minAcceptableDuration {
			// stop iterating when there's no longer 20 minutes of data left
			break
		}

		sum := 0
		for _, dp := range window {
			sum += dp.HeartRate
		}

		avg := float64(sum) / float64(len(window))
		if avg > bestAvg {
			bestAvg = avg
			bestWindow = window
		}
	}

	if bestWindow == nil {
		return nil, errors.New("Could not find valid 20-minute window")
	}

	return bestWindow, nil
}
