package powerschool

// with 3 categories: 20% low, 30% mid, 50% high
// with 1 category: no weighting
//
// with 2 categories:
// low and mid, 40% low and 60% mid (verified)
// low and high, 30% low and 70% high (verified)
// mid and high, 40% mid and 60% high (not verified, but probably)

// GradeHolder accumulates the assignment grades for one weight category
// (Low/Mid/High) within a Section and computes that category's contribution
// to the section's final grade.
type GradeHolder struct {
	grade       uint64
	num, weight float64
}

// NewGradeHolder creates a GradeHolder for a category with the given weight
// (e.g. 0.2 for a 20%-weighted "Low" category).
func NewGradeHolder(weight float64) *GradeHolder {
	return &GradeHolder{weight: weight}
}

// Add records one assignment's grade into this category.
func (h *GradeHolder) Add(grade uint64) {
	h.grade += grade
	h.num++
}

// Num returns how many grades have been added since the last Reset.
func (h *GradeHolder) Num() float64 {
	return h.num
}

// Weight returns this category's configured weight.
func (h *GradeHolder) Weight() float64 {
	return h.weight
}

// Final returns this category's weighted contribution to the section's final
// grade, given how many categories (1, 2, or 3) have grades this section. If
// categories == 2, other must contain the weight of the other category
// present — PowerSchool re-splits the remaining weight between exactly two
// active categories rather than just dropping the third's share.
func (h *GradeHolder) Final(categories float64, other ...float64) float64 {
	avg := float64(h.grade) / h.num

	// powerschool will be the end of me
	switch categories {
	case 3: // with 3 categories: 20% low, 30% mid, 50% high
		return avg * h.weight
	case 1: // with 1 category: no weighting
		return avg
	default:
		var weight float64
		cat := other[0]

		switch h.weight {
		case 0.2:
			if cat == 0.3 { // low and mid
				weight = 0.4
			} else { // low and high
				weight = 0.3
			}
		case 0.3:
			if cat == 0.2 { // mid and low
				weight = 0.6
			} else { // mid and high
				weight = 0.4
			}
		case 0.5:
			if cat == 0.2 { // high and low
				weight = 0.7
			} else { // high and mid
				weight = 0.6
			}
		}

		return avg * weight
	}
}

// Reset clears the accumulated grades, ready to reuse this GradeHolder for a
// fresh FinalGrade calculation.
func (h *GradeHolder) Reset() {
	h.grade = 0
	h.num = 0
}
