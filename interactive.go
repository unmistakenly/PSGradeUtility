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

var ErrNotSignedIn = errors.New("you need to sign in first")

const QuarterFormat = "Monday, January 2, 2006"

const HelpText = `commands:

h - view this help text at any time
q - fully quit at any time
s - sign in to powerschool
u - check who you are currently signed in as
a - get ALL grades`

func signIn() (username, ticket, studentID string, err error) {
	fmt.Print("username: ")
	fmt.Scanln(&username)

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

func showAllGrades(ticket, studentID string) error {
	if ticket == "" {
		return ErrNotSignedIn
	}

	r, err := GetFullData(ticket, studentID)
	if err != nil {
		return fmt.Errorf("couldnt get data from powerschool: %w", err)
	}
	defer r.Close()

	var data *powerschool.FullResponse
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return err
	}

	quarterStart, quarterEnd := data.Response.Return.Data.GetCurrentQuarter()
	fmt.Println("Quarter Start:", quarterStart.Format(QuarterFormat))
	fmt.Println("Quarter End:", quarterEnd.Format(QuarterFormat))
	fmt.Println()

	weightIDs := map[int]string{}
	for _, w := range data.Response.Return.Data.Categories {
		weightIDs[w.WeightID] = w.WeightName
	}

	classes := map[int]*powerschool.Section{}
	for _, s := range data.Response.Return.Data.Sections {
		classes[s.ClassID] = s
	}

	assignments := map[int]*powerschool.Assignment{}
	for _, a := range data.Response.Return.Data.Assignments {
		assigned, err := time.Parse(time.RFC3339, a.DueDate)
		if err != nil {
			continue
		}
		after := assigned.Compare(quarterStart)
		before := assigned.Compare(quarterEnd)
		if after >= 0 && before <= 0 {
			assignments[a.ID] = a
		}
	}
	for _, s := range data.Response.Return.Data.Scores {
		if assignment, ok := assignments[s.AssignmentID]; ok {
			assignment.Percent = s.Percent
			classes[assignment.SectionID].Assignments = append(classes[assignment.SectionID].Assignments, assignment)
		}
	}

	// fmt.Println("Quarter Start:", QuarterStart)
	// fmt.Println("Quarter End:", QuarterEnd)
	// fmt.Println()

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
