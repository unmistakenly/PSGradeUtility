package powerschool

import (
	"math"
	"slices"
)

// FinalGrade computes this section's overall grade (0-100, rounded) from its
// current Assignments, weighted by category per weightIDs (category ID ->
// "Low"/"Mid"/"High"). Categories with no graded assignments are excluded
// from the calculation rather than counted as 0%. Returns
// ErrUnsupportedWeightCombination if exactly 2 categories have grades and
// the weights aren't PowerSchool's verified default split (see that error).
func (s *Section) FinalGrade(weightIDs map[int]string) (float64, error) {
	// this way, we dont need to allocate so much
	s.Low.Reset()
	s.Mid.Reset()
	s.High.Reset()

	for _, a := range s.Assignments {
		switch weightIDs[a.CategoryID] {
		case "Low":
			s.Low.Add(a.Grade)
		case "Mid":
			s.Mid.Add(a.Grade)
		case "High":
			s.High.Add(a.Grade)
		}
	}

	categories := []*GradeHolder{s.Low, s.Mid, s.High}
	categories = slices.DeleteFunc(categories, func(h *GradeHolder) bool {
		return h.Num() == 0 // if this category doesnt have any grades in it
	})

	// if there is only 1 category of grades, dont calculate weight at all
	if l := len(categories); l == 1 {
		v, err := categories[0].Final(1)
		return math.Round(v), err
	} else if l == 2 {
		c0 := categories[0]
		c1 := categories[1]
		v0, err := c0.Final(2, c1.Weight())
		if err != nil {
			return 0, err
		}
		v1, err := c1.Final(2, c0.Weight())
		if err != nil {
			return 0, err
		}
		return math.Round(v0 + v1), nil
	}

	low, err := s.Low.Final(3)
	if err != nil {
		return 0, err
	}
	mid, err := s.Mid.Final(3)
	if err != nil {
		return 0, err
	}
	high, err := s.High.Final(3)
	if err != nil {
		return 0, err
	}
	return math.Round(low + mid + high), nil
}
