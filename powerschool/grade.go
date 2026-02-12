package powerschool

type GradeHolder struct {
	grade, num, weight float64
}

func NewGradeHolder(weight float64) *GradeHolder {
	return &GradeHolder{weight: weight}
}

func (h *GradeHolder) Add(grade float64) {
	h.grade += grade
	h.num++
}

func (h *GradeHolder) Num() float64 {
	return h.num
}

// the caller is responsible for managing weighted grades
func (h *GradeHolder) Final(categories float64) float64 {
	avg := h.grade / h.num

	// powerschool will be the end of me
	switch categories {
	case 1:
		return avg
	case 2:
		return avg * h.weight * categories
	default: // 3
		return avg * h.weight
	}
}

func (h *GradeHolder) Reset() {
	h.grade = 0
	h.num = 0
}
