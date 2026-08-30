package cli

import (
	"testing"
	"time"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
)

// extractInfoFromResponse is pure (no network) — cover the quarter-window
// filtering and weight-holder wiring directly.
func TestExtractInfoFromResponse_FiltersToQuarterWindow(t *testing.T) {
	qStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	qEnd := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	data := &powerschool.FullResponse{}
	data.Response.Return.Data = &powerschool.DataResponse{
		Categories: []powerschool.Category{{WeightID: 1, WeightName: "Low"}},
		Sections:   []*powerschool.Section{{ClassID: 1, ClassName: "Math"}},
		Assignments: []*powerschool.Assignment{
			{ID: 10, SectionID: 1, CategoryID: 1, DueDate: "2026-02-01T00:00:00Z"}, // in window
			{ID: 11, SectionID: 1, CategoryID: 1, DueDate: "2026-06-01T00:00:00Z"}, // out of window
		},
		Scores: []powerschool.Score{
			{AssignmentID: 10, Percent: 91.0},
			{AssignmentID: 11, Percent: 50.0},
		},
	}

	classes, weightIDs := extractInfoFromResponse(data, qStart, qEnd, [3]float64{0.2, 0.3, 0.5})

	if weightIDs[1] != "Low" {
		t.Fatalf("weightIDs not built: %+v", weightIDs)
	}
	class := classes[1]
	if len(class.Assignments) != 1 {
		t.Fatalf("want 1 assignment surviving the quarter-window filter, got %d", len(class.Assignments))
	}
	if class.Assignments[0].ID != 10 || class.Assignments[0].Grade != 91 {
		t.Fatalf("got assignment %+v", class.Assignments[0])
	}
}

func TestExtractInfoFromResponse_ExemptAndCollectedScoresExcluded(t *testing.T) {
	qStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	qEnd := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	data := &powerschool.FullResponse{}
	data.Response.Return.Data = &powerschool.DataResponse{
		Sections: []*powerschool.Section{{ClassID: 1}},
		Assignments: []*powerschool.Assignment{
			{ID: 1, SectionID: 1, DueDate: "2026-02-01T00:00:00Z"},
			{ID: 2, SectionID: 1, DueDate: "2026-02-01T00:00:00Z"},
		},
		Scores: []powerschool.Score{
			{AssignmentID: 1, Exempt: true, Percent: 100.0},
			{AssignmentID: 2, Collected: true, Percent: 100.0},
		},
	}

	classes, _ := extractInfoFromResponse(data, qStart, qEnd, [3]float64{0.2, 0.3, 0.5})
	if len(classes[1].Assignments) != 0 {
		t.Fatalf("exempt/collected assignments should be excluded, got %+v", classes[1].Assignments)
	}
}
