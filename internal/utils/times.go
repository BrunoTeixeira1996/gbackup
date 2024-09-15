package utils

import "time"

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
	return time.Now().Format("2006-02-01")
}
