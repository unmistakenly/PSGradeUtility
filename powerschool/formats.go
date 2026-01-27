package powerschool

const TimeFormat = "2006-01-02T15:04:05.000Z"

type Section struct {
	ClassName string `json:"schoolCourseTitle"`
	ClassID   int    `json:"id"`

	Assignments []*Assignment
}

type Assignment struct {
	Name       string  `json:"name"`
	ID         int     `json:"id"`         // get the score with matching id
	CategoryID int     `json:"categoryId"` // match this to Low/Mid/High
	DueDate    string  `json:"dueDate"`
	SectionID  int     `json:"sectionid"`
	Percent    float64 // set this yourself
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
