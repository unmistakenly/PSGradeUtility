package powerschool

import "errors"

// ErrUnsupportedWeightCombination is returned by GradeHolder.Final when
// exactly 2 of the 3 categories have grades and the weights aren't
// PowerSchool's default 20/30/50 split. The default split's 2-category
// redistribution below is an EMPIRICAL table (see the comment block), only
// ever verified against a real account using the default weights — it is
// NOT proportional to the two active weights (low+high verified 30/70 vs.
// proportional 28.6/71.4; mid+high verified 40/60 vs. proportional 37.5/
// 62.5), so there's no formula to generalize it to custom weights, and no
// live account with custom weights to verify one against. Rather than fake a
// redistribution, custom weights are only supported where the math is
// general: 1-category (no weighting) and 3-category (a plain weighted sum,
// no redistribution needed).
var ErrUnsupportedWeightCombination = errors.New("2-category grade redistribution is only verified for PowerSchool's default 20/30/50 weighting; this section has custom weights with exactly 2 graded categories, which can't be computed without a live account to verify against")

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
// active categories rather than just dropping the third's share. Returns
// ErrUnsupportedWeightCombination for the 2-category case under non-default
// weights (see that error's doc for why).
func (h *GradeHolder) Final(categories float64, other ...float64) (float64, error) {
	avg := float64(h.grade) / h.num

	// powerschool will be the end of me
	switch categories {
	case 3: // with 3 categories: general weighted sum, valid for any weights
		return avg * h.weight, nil
	case 1: // with 1 category: no weighting
		return avg, nil
	default:
		var weight float64
		cat := other[0]

		// this table is only verified against PowerSchool's default 20/30/50
		// split — see ErrUnsupportedWeightCombination.
		switch {
		case h.weight == 0.2 && cat == 0.3: // low and mid
			weight = 0.4
		case h.weight == 0.2 && cat == 0.5: // low and high
			weight = 0.3
		case h.weight == 0.3 && cat == 0.2: // mid and low
			weight = 0.6
		case h.weight == 0.3 && cat == 0.5: // mid and high
			weight = 0.4
		case h.weight == 0.5 && cat == 0.2: // high and low
			weight = 0.7
		case h.weight == 0.5 && cat == 0.3: // high and mid
			weight = 0.6
		default:
			return 0, ErrUnsupportedWeightCombination
		}

		return avg * weight, nil
	}
}

// Reset clears the accumulated grades, ready to reuse this GradeHolder for a
// fresh FinalGrade calculation.
func (h *GradeHolder) Reset() {
	h.grade = 0
	h.num = 0
}
