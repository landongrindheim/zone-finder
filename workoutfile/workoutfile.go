package workoutfile

type WorkoutFile interface {
	GetHRDataPoints() ([]HRDataPoint, error)
	GetDeviceName() string
	GetProductID() int
}
