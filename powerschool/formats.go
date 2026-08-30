package powerschool

const TimeFormat = "2006-01-02T15:04:05.000Z"

type Section struct {
	ClassName string `json:"schoolCourseTitle"`
	ClassID   int    `json:"id"`

	Assignments []*Assignment

	Low, Mid, High *GradeHolder
}

type Assignment struct {
	Name       string `json:"name"`
	ID         int    `json:"id"`         // get the score with matching id
	CategoryID int    `json:"categoryId"` // match this to Low/Mid/High
	DueDate    string `json:"dueDate"`
	SectionID  int    `json:"sectionid"`

	// set these yourself
	Grade         uint64 // current grade, used for calculation
	OriginalGrade uint64 // only meaningful when Edited is true
	Edited        bool
}

// Edit sets a's grade to newGrade, remembering the pre-edit grade the first
// time (a second Edit before a Restore does NOT stomp the original — matches
// the old bit-packed code's "if !gradeIsEdited" guard).
func (a *Assignment) Edit(newGrade uint64) {
	if !a.Edited {
		a.OriginalGrade = a.Grade
		a.Edited = true
	}
	a.Grade = newGrade
}

// Restore reverts an edited assignment's grade back to what it was before
// Edit was called. No-op if the assignment was never edited.
func (a *Assignment) Restore() {
	if !a.Edited {
		return
	}
	a.Grade = a.OriginalGrade
	a.Edited = false
}

type Term struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type Category struct {
	WeightID   int    `json:"id"`
	WeightName string `json:"name"` // Low/Mid/High
}

type Score struct {
	Collected    bool `json:"collected"`
	Exempt       bool `json:"exempt"`
	AssignmentID int  `json:"assignmentId"`
	Percent      any  `json:"percent"`
}

type School struct {
	Disabled    bool `json:"schoolDisabled"`
	DisabledMsg any  `json:"schoolDisabledMessage"`
}

type DataResponse struct {
	Sections    []*Section    `json:"sections"`
	Assignments []*Assignment `json:"assignments"`
	Terms       []*Term       `json:"terms"`
	Scores      []Score       `json:"assignmentScores"`
	Categories  []Category    `json:"assignmentCategories"`

	Schools School `json:"schools"` // yes, this is a singular item with a plural name, nice one powerschool
}

type FullResponse struct {
	Response struct {
		Return struct {
			Data *DataResponse `json:"studentDataVOs"`
		} `json:"return"`
	} `json:"getStudentDataResponse"`
}
