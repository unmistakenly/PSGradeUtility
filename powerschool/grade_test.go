package powerschool

import "testing"

// Table-driven coverage of FinalGrade's 1/2/3-category weighting, matching
// the verified splits documented in grade.go's own comments (40/60 low+mid,
// 30/70 low+high, 20/30/50 all three).

func section(weightIDs map[int]string, grades map[string][]uint64) *Section {
	s := &Section{
		Low:  NewGradeHolder(0.2),
		Mid:  NewGradeHolder(0.3),
		High: NewGradeHolder(0.5),
	}
	id := 1
	for cat, gs := range grades {
		catID := id
		id++
		weightIDs[catID] = cat
		for _, g := range gs {
			s.Assignments = append(s.Assignments, &Assignment{CategoryID: catID, Grade: g})
		}
	}
	return s
}

func TestFinalGrade_SingleCategory(t *testing.T) {
	weightIDs := map[int]string{}
	s := section(weightIDs, map[string][]uint64{"Mid": {80, 90, 100}})
	got := s.FinalGrade(weightIDs)
	if got != 90 {
		t.Fatalf("want 90, got %v", got)
	}
}

func TestFinalGrade_TwoCategories_LowMid(t *testing.T) {
	weightIDs := map[int]string{}
	s := section(weightIDs, map[string][]uint64{"Low": {100}, "Mid": {0}})
	// 40% low, 60% mid per the verified comment
	got := s.FinalGrade(weightIDs)
	if got != 40 {
		t.Fatalf("want 40 (100*0.4 + 0*0.6), got %v", got)
	}
}

func TestFinalGrade_TwoCategories_LowHigh(t *testing.T) {
	weightIDs := map[int]string{}
	s := section(weightIDs, map[string][]uint64{"Low": {100}, "High": {0}})
	// 30% low, 70% high per the verified comment
	got := s.FinalGrade(weightIDs)
	if got != 30 {
		t.Fatalf("want 30, got %v", got)
	}
}

func TestFinalGrade_ThreeCategories(t *testing.T) {
	weightIDs := map[int]string{}
	s := section(weightIDs, map[string][]uint64{"Low": {100}, "Mid": {100}, "High": {0}})
	// 20% low + 30% mid + 0% high = 50
	got := s.FinalGrade(weightIDs)
	if got != 50 {
		t.Fatalf("want 50, got %v", got)
	}
}

func TestFinalGrade_EmptyCategoryExcluded(t *testing.T) {
	weightIDs := map[int]string{}
	// a category with zero assignments must not be counted (h.Num()==0 filter)
	s := section(weightIDs, map[string][]uint64{"High": {80}})
	got := s.FinalGrade(weightIDs)
	if got != 80 {
		t.Fatalf("want 80 (single-category path, no weighting), got %v", got)
	}
}

// Assignment.Edit/Restore round-trip — this replaces the old bit-packed
// uint64 hack; behavior (not repeated edits stomping the original, restore
// reverting exactly) must be identical.

func TestAssignment_EditThenRestore(t *testing.T) {
	a := &Assignment{Grade: 75}

	a.Edit(90)
	if a.Grade != 90 || a.OriginalGrade != 75 || !a.Edited {
		t.Fatalf("after first edit: got Grade=%d OriginalGrade=%d Edited=%v", a.Grade, a.OriginalGrade, a.Edited)
	}

	// a second edit before restore must NOT overwrite OriginalGrade
	a.Edit(60)
	if a.Grade != 60 || a.OriginalGrade != 75 {
		t.Fatalf("after second edit: got Grade=%d OriginalGrade=%d, want Grade=60 OriginalGrade=75", a.Grade, a.OriginalGrade)
	}

	a.Restore()
	if a.Grade != 75 || a.Edited {
		t.Fatalf("after restore: got Grade=%d Edited=%v, want Grade=75 Edited=false", a.Grade, a.Edited)
	}
}

func TestAssignment_RestoreWithoutEditIsNoop(t *testing.T) {
	a := &Assignment{Grade: 88}
	a.Restore()
	if a.Grade != 88 || a.Edited {
		t.Fatalf("restore on unedited assignment changed state: Grade=%d Edited=%v", a.Grade, a.Edited)
	}
}

func TestAssignment_EditZeroGrade(t *testing.T) {
	// the old bit-pack used the MSB specifically "to account for a grade of 0"
	// — confirm a 0 grade still round-trips correctly with the new struct.
	a := &Assignment{Grade: 50}
	a.Edit(0)
	if a.Grade != 0 || a.OriginalGrade != 50 || !a.Edited {
		t.Fatalf("edit to 0: got Grade=%d OriginalGrade=%d Edited=%v", a.Grade, a.OriginalGrade, a.Edited)
	}
	a.Restore()
	if a.Grade != 50 || a.Edited {
		t.Fatalf("restore after edit-to-0: got Grade=%d Edited=%v", a.Grade, a.Edited)
	}
}
