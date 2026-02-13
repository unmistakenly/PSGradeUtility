package powerschool

import (
	"math"
	"slices"
)

func (s *Section) FinalGrade(weightIDs map[int]string) float64 {
	// this way, we dont need to allocate so much
	s.Low.Reset()
	s.Mid.Reset()
	s.High.Reset()

	for _, a := range s.Assignments {
		switch weightIDs[a.CategoryID] {
		case "Low":
			s.Low.Add(a.Percent)
		case "Mid":
			s.Mid.Add(a.Percent)
		case "High":
			s.High.Add(a.Percent)
		}
	}

	categories := []*GradeHolder{s.Low, s.Mid, s.High}
	categories = slices.DeleteFunc(categories, func(h *GradeHolder) bool {
		return h.Num() == 0 // if this category doesnt have any grades in it
	})

	// if there is only 1 category of grades, dont calculate weight at all
	if l := len(categories); l == 1 {
		return math.Round(categories[0].Final(1))
	} else if l == 2 {
		c0 := categories[0]
		c1 := categories[1]
		return math.Round(c0.Final(2, c1.Weight()) + c1.Final(2, c0.Weight()))
	}

	return math.Round(
		s.Low.Final(3) +
			s.Mid.Final(3) +
			s.High.Final(3),
	)
}
