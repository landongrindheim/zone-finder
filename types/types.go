package types

import "time"

type HRDataPoint struct {
	Timestamp time.Time
	HeartRate int
}
