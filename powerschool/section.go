package powerschool

import (
	"math"
	"slices"
)

func (s *Section) FinalGrade(weightIDs map[int]string) float64 {
	low := NewGradeHolder(0.2)
	mid := NewGradeHolder(0.3)
	high := NewGradeHolder(0.5)

	for _, a := range s.Assignments {
		switch weightIDs[a.CategoryID] {
		case "Low":
			low.Add(a.Percent)
		case "Mid":
			mid.Add(a.Percent)
		case "High":
			high.Add(a.Percent)
		}
	}

	categories := []*GradeHolder{low, mid, high}
	categories = slices.DeleteFunc(categories, func(h *GradeHolder) bool {
		return h.Num() == 0 // if this category doesnt have any grades in it
	})

	// if there is only 1 category of grades, dont calculate weight at all
	if l := len(categories); l == 1 {
		return math.Round(categories[0].Final(1))
	} else if l == 2 { // otherwise, double the weights (i Pray that there is no class that assigns only mid and high grades)
		return math.Round(categories[0].Final(2) + categories[1].Final(2))
	}

	// so, when should we actually multiply by the weight? should be when all three categories are present
	// if theres only one, then you dont need to multiply by any weight at all
	//
	// but what if theres only two? apparently, having a low grade 100 and a mid grade 80 gives an 88% average
	// it works with: low * (0.2 * 2) + mid * (0.3 * 2) = 88
	// but what if there was only a mid and a high grade? errr, that shouldnt happen in the real world i think, so it should be fine
	return math.Round(
		low.Final(3) +
			mid.Final(3) +
			high.Final(3),
	)
}
