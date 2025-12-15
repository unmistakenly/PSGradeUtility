package powerschool

import "math"

func calculateFinalGrade(score, num, weight float64) float64 {
	if num == 0 {
		return 100 * weight
	}
	return score / num * weight
}

func (s *Section) FinalGrade(weightIDs map[int]string) float64 {
	var low, mid, high float64
	var nLow, nMid, nHigh float64

	for _, a := range s.Assignments {
		switch weightIDs[a.CategoryID] {
		case "Low":
			low += a.Percent
			nLow++
		case "Mid":
			mid += a.Percent
			nMid++
		case "High":
			high += a.Percent
			nHigh++
		}
	}

	return math.Round(
		calculateFinalGrade(low, nLow, 0.2) +
			calculateFinalGrade(mid, nMid, 0.3) +
			calculateFinalGrade(high, nHigh, 0.5),
	)
}
