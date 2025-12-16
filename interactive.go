package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
	"golang.org/x/term"
)

var (
	ErrNotSignedIn     = errors.New("you need to sign in first")
	ErrSignInCancelled = errors.New("sign in cancelled")
	ErrNoStudentData   = errors.New("studentDataVOs is nil, maybe your session expired? try signing in again")
)

const QuarterFormat = "Monday, January 2, 2006"

const HelpText = `commands:

h - view this help text at any time
q - fully quit at any time
s - sign in to powerschool
u - check who you are currently signed in as
a - get ALL grades
c - enter grade calculator`

func signIn() (username, ticket, studentID string, err error) {
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

	ticket, studentID, err = GetServiceTicket(username, string(password))
	if err != nil {
		return "", "", "", fmt.Errorf("\ncouldnt sign in: %w", err)
	}
	return
}

func getFullDecodedResponse(ticket, studentID string) (*powerschool.FullResponse, error) {
	r, err := GetFullData(ticket, studentID)
	if err != nil {
		return nil, fmt.Errorf("couldnt get data from powerschool: %w", err)
	}
	defer r.Close()

	var data *powerschool.FullResponse
	err = json.NewDecoder(r).Decode(&data)

	// yeah nvm forget about it, reading the whole response body is fine!
	// and we can just check if studentDataVOs is nil
	if data.Response.Return.Data == nil {
		return nil, ErrNoStudentData
	}

	return data, err
}

func extractInfoFromResponse(data *powerschool.FullResponse, qStart, qEnd time.Time) (
	classes map[int]*powerschool.Section,
	weightIDs map[int]string,
) {
	// this could all probably be optimized, but it's good enough
	weightIDs = map[int]string{}
	for _, w := range data.Response.Return.Data.Categories {
		weightIDs[w.WeightID] = w.WeightName
	}

	classes = map[int]*powerschool.Section{}
	for _, s := range data.Response.Return.Data.Sections {
		classes[s.ClassID] = s
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
			a.Percent = 100 // default?
			assignments[a.ID] = a
		}
	}
	for _, s := range data.Response.Return.Data.Scores {
		if assignment, ok := assignments[s.AssignmentID]; ok {
			if s.Exempt || s.Collected {
				continue
			}
			if p, ok := s.Percent.(float64); ok {
				assignment.Percent = p
			}
			classes[assignment.SectionID].Assignments = append(classes[assignment.SectionID].Assignments, assignment)
		}
	}

	return
}

func showAllGrades(ticket, studentID string) error {
	if ticket == "" {
		return ErrNotSignedIn
	}

	data, err := getFullDecodedResponse(ticket, studentID)
	if err != nil {
		return err
	}

	quarterStart, quarterEnd := data.Response.Return.Data.GetCurrentQuarter()
	fmt.Println("Quarter Start:", quarterStart.Format(QuarterFormat))
	fmt.Println("Quarter End:", quarterEnd.Format(QuarterFormat))
	fmt.Println()

	classes, weightIDs := extractInfoFromResponse(data, quarterStart, quarterEnd)

	for _, c := range classes {
		if len(c.Assignments) == 0 {
			continue
		}
		slices.SortFunc(c.Assignments, func(a1, a2 *powerschool.Assignment) int {
			t1, _ := time.Parse(powerschool.TimeFormat, a1.DueDate)
			t2, _ := time.Parse(powerschool.TimeFormat, a2.DueDate)
			return -t1.Compare(t2) // negative to sort in descending order, like the app
		})
		fmt.Printf("\033[1m+ %s (%.0f%%)\033[0m\n", c.ClassName, c.FinalGrade(weightIDs))
		for _, v := range c.Assignments {
			fmt.Printf("%s - %v%% (%s)\n", v.Name, v.Percent, weightIDs[v.CategoryID])
		}
		fmt.Println()
	}

	return nil
}
