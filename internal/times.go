package internal

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
