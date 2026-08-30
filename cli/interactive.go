package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
	"golang.org/x/term"
)

const QuarterFormat = "Monday, January 2, 2006"

func (s *Session) signIn() (username, ticket, studentID string, err error) {
	fmt.Print("username: ")
	fmt.Scanln(&username)

	// just in case they accidentally use the command and back out
	// (definitely didnt happen to me)
	if username == "" {
		err = ErrSignInCancelled
		return
	}

	fmt.Print("password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return
	}

	ticket, studentID, err = s.Client.GetServiceTicket(username, string(password))
	if err != nil {
		return "", "", "", fmt.Errorf("\ncouldnt sign in: %w", err)
	}
	return
}

func (s *Session) getFullDecodedResponse() (*powerschool.FullResponse, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// todo: maybe when signing in as well? that takes a while sometimes
	go func(ctx context.Context) {
		var chars = [4]string{"/", "-", `\`, "|"}
		var i int

		for {
			if ctx.Err() != nil {
				fmt.Print("\033[K") // clear line
				return
			}

			fmt.Printf("fetching data... %s\r", chars[i%4])
			i++
			time.Sleep(80 * time.Millisecond)
		}
	}(ctx)

	r, err := s.Client.GetFullData(s.ticket, s.studentID)
	if err != nil {
		return nil, fmt.Errorf("couldnt get data from powerschool: %w", err)
	}
	defer r.Close()

	var data *powerschool.FullResponse
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, fmt.Errorf("couldnt parse response from powerschool: %w", err)
	}

	// data.Response.Return.Data can genuinely be nil (session expired) even
	// on a successful decode — this check must come AFTER the decode-error
	// check above, not before it (the old code checked it first, which
	// nil-derefs data.Response if Decode itself had failed).
	if data.Response.Return.Data == nil {
		return nil, ErrNoStudentData
	} else if data.Response.Return.Data.Schools.Disabled {
		return nil, fmt.Errorf("warning: data access is disabled. message:\n\n%v", data.Response.Return.Data.Schools.DisabledMsg)
	}

	return data, nil
}

func extractInfoFromResponse(data *powerschool.FullResponse, qStart, qEnd time.Time, weights [3]float64) (
	classes map[int]*powerschool.Section,
	weightIDs map[int]string,
) {
	// this could all probably be optimized, but it's good enough
	weightIDs = map[int]string{}
	for _, w := range data.Response.Return.Data.Categories {
		weightIDs[w.WeightID] = w.WeightName
	}

	classes = map[int]*powerschool.Section{}
	for _, sec := range data.Response.Return.Data.Sections {
		sec.Low = powerschool.NewGradeHolder(weights[0])
		sec.Mid = powerschool.NewGradeHolder(weights[1])
		sec.High = powerschool.NewGradeHolder(weights[2])
		classes[sec.ClassID] = sec
	}

	// this part is split into 2 steps to avoid adding assignments that arent yet graded
	assignments := map[int]*powerschool.Assignment{}
	for _, a := range data.Response.Return.Data.Assignments {
		assigned, err := time.Parse(time.RFC3339, a.DueDate)
		if err != nil {
			continue
		}
		after := assigned.Compare(qStart)
		before := assigned.Compare(qEnd)
		if after >= 0 && before <= 0 {
			a.Grade = 100 // default?
			assignments[a.ID] = a
		}
	}
	for _, sc := range data.Response.Return.Data.Scores {
		if assignment, ok := assignments[sc.AssignmentID]; ok {
			if sc.Exempt || sc.Collected {
				continue
			}
			if p, ok := sc.Percent.(float64); ok {
				assignment.Grade = uint64(p)
			}
			classes[assignment.SectionID].Assignments = append(classes[assignment.SectionID].Assignments, assignment)
		}
	}

	for _, c := range classes {
		slices.SortFunc(c.Assignments, func(a1, a2 *powerschool.Assignment) int {
			t1, _ := time.Parse(powerschool.TimeFormat, a1.DueDate)
			t2, _ := time.Parse(powerschool.TimeFormat, a2.DueDate)
			return -t1.Compare(t2) // negative to sort in descending order, like the app
		})
	}

	return
}

func (s *Session) showAllGrades() error {
	if s.ticket == "" {
		return ErrNotSignedIn
	}

	data, err := s.getFullDecodedResponse()
	if err != nil {
		return err
	}

	quarterStart, quarterEnd := data.Response.Return.Data.GetCurrentQuarter()
	fmt.Println("Quarter Start:", quarterStart.Format(QuarterFormat))
	fmt.Println("Quarter End:", quarterEnd.Format(QuarterFormat))
	fmt.Println()

	classes, weightIDs := extractInfoFromResponse(data, quarterStart, quarterEnd, s.weights())

	for _, c := range classes {
		if len(c.Assignments) == 0 {
			continue
		}
		grade, err := c.FinalGrade(weightIDs)
		fmt.Printf("\033[1m+ %s (%s)\033[0m\n", c.ClassName, formatGrade(grade, err))
		for _, v := range c.Assignments {
			fmt.Printf("%s - %v%% (%s)\n", v.Name, v.Grade, weightIDs[v.CategoryID])
		}
		fmt.Println()
	}

	return nil
}
