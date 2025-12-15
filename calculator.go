package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
)

const CalcHelpText = `calculator commands:

h - view this help text at any time
q - fully quit at any time
b - exit grade calculator and return to main menu

and 0-3 or whatever indexes to enter the grade calculator for a class.`

// gradeCalculator will start its own input loop
func gradeCalculator(ticket, studentID string) error {
	// this will have a lot of shared code with [showAllGrades]
	if ticket == "" {
		return ErrNotSignedIn
	}

	data, err := getFullDecodedResponse(ticket, studentID)
	if err != nil {
		return err
	}

	quarterStart, quarterEnd := data.Response.Return.Data.GetCurrentQuarter()
	classes, weightIDs := extractInfoFromResponse(data, quarterStart, quarterEnd)
	fmt.Println()

	sclasses := make([]*powerschool.Section, 0, len(classes))
	for _, c := range classes {
		if len(c.Assignments) == 0 {
			continue
		}
		sclasses = append(sclasses, c)
	}
	for i, c := range sclasses {
		fmt.Printf("[%d] %s (%.0f%%)\n", i, c.ClassName, c.FinalGrade(weightIDs))
	}

	var input string
	fmt.Print("\n>> ")

	for {
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "h", "help":
			fmt.Println(HelpText)
		case "q", "quit":
			return nil
		case "b":
			return nil
		default:
			i, err := strconv.Atoi(input)
			if err != nil || i < 0 || i > len(sclasses)-1 {
				break
			}
			classCalculator(sclasses[i])
		}

		fmt.Print("\n>> ")
	}
}

func classCalculator(section *powerschool.Section) error {
	return nil
}
