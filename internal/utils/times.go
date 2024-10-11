package utils

import (
	"time"
)

type ElapsedTime struct {
	Target  string
	Elapsed float64
}

func CalculateTotalElaspedTime(times []ElapsedTime) float64 {
	var totalElapsedTime float64
	for _, t := range times {
		totalElapsedTime += t.Elapsed
	}

	return totalElapsedTime
}

func CurrentTime() string {
	return time.Now().Format("2006-01-02")
}

// Gets epoch time for the current day at 12 PM
func Epoch() int64 {
	now := time.Now()

	timeAt2PM := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		12, 0, 0, 0,
		now.Location())

	return timeAt2PM.Unix()
}
