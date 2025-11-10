package workoutfile

import "zone-finder/types"

type WorkoutFile interface {
	GetHRDataPoints() ([]types.HRDataPoint, error)
	GetDeviceName() string
	GetProductID() int
}
