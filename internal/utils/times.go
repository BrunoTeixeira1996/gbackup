package utils

import (
	"time"
)

type ElapsedTime struct {
	Target string
	Value  float64
}

func CalculateTotalElapsedTime(times []ElapsedTime) float64 {
	var totalElapsedTime float64
	for _, t := range times {
		totalElapsedTime += t.Value
	}

	return totalElapsedTime
}

func CurrentTime() string {
	return time.Now().Format("2006-01-02")
}

// Gets epoch time for the current day at midnight
// So we do the PBS backup at 12 and then at 13
// we grab every backup that was done during the current day
// meaning we grab the backup at 12 and this works in prod
// and in debug mode
func Epoch() int64 {
	now := time.Now()

	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		now.Location()).Unix()
}

// Used to calculate where next friday is
func NextFridayAt13(now time.Time) time.Time {
	nextFriday := now.AddDate(0, 0, (int(time.Friday)-int(now.Weekday())+7)%7)
	return time.Date(nextFriday.Year(), nextFriday.Month(), nextFriday.Day(), 13, 0, 0, 0, now.Location())
}
