package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/unmistakenly/PSGradeUtility/powerschool"
)

const CalcMenuHelpText = `calculator menu commands:

h - view this help text at any time
q - fully quit at any time
b - exit grade calculator and return to main menu

or any number below associated with a class:`

const CalcHelpText = `calculator commands:

h - view this help text at any time
q - fully quit at any time
b - return to grade calculator main menu

add <low/mid/high> <0-100> [<name>]
del <index>
view`

// gradeCalculator will start its own input loops
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

	sclasses := make([]*powerschool.Section, 0, len(classes))
	for _, c := range classes {
		if len(c.Assignments) == 0 {
			continue
		}
		sclasses = append(sclasses, c)
	}

	printClasses := func() {
		for i, c := range sclasses {
			fmt.Printf("[%d] %s (%.0f%%)\n", i, c.ClassName, c.FinalGrade(weightIDs))
		}
	}

	var input string
	printClasses()

	for {
		fmt.Print("\n>> ")
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "h", "help":
			fmt.Println(CalcMenuHelpText)
			fmt.Println()
			printClasses()
		case "q", "quit":
			os.Exit(0)
		case "b":
			return nil
		default:
			i, err := strconv.Atoi(input)
			if err != nil || i < 0 || i > len(sclasses)-1 {
				fmt.Println("unrecognized input")
				break
			}
			classCalculator(sclasses[i].Assignments, i, weightIDs)
		}
	}
}

func classCalculator(origAssignments []*powerschool.Assignment, i int, weightIDs map[int]string) error {
	var input string

	// enforce access of assignments through section only, as it will otherwise cause a runtime error (TOTALLY didnt happen)
	section := func() *powerschool.Section {
		// deep copy of origAssignments, so as to not modify it
		assignments := make([]*powerschool.Assignment, len(origAssignments), len(origAssignments)+3)
		for i := range assignments {
			orig := origAssignments[i]
			assignments[i] = &powerschool.Assignment{
				Name:       orig.Name,
				CategoryID: orig.CategoryID,
				Percent:    orig.Percent,
			}
		}
		return &powerschool.Section{Assignments: assignments}
	}()

	// i couldnt think of any other way to do this
	weightToIDs := make(map[string]int, 3)
	for id, weight := range weightIDs {
		weight = strings.ToLower(weight)
		if _, ok := weightToIDs[weight]; !ok {
			weightToIDs[weight] = id
		}
	}

	printAssignments := func() {
		for i, a := range section.Assignments {
			fmt.Printf("[%d] %s - %.0f%% (%s)\n", i, a.Name, a.Percent, weightIDs[a.CategoryID])
		}
	}
	printAssignments()

	// i actually didnt know fmt.Scanln seperated inputs with spaces. funky!
	stdinReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n(%d) >> ", i)
		input, _ = stdinReader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "h", "help":
			fmt.Println(CalcHelpText)
		case "q", "quit":
			os.Exit(0)
		case "b":
			return nil
		default:
			args := strings.SplitN(input, " ", 4)
			switch args[0] { // lol rip
			case "v", "view":
				printAssignments()
				fmt.Printf("\nfinal grade: %.0f%%\n", section.FinalGrade(weightIDs))
			case "a", "add":
				if len(args) < 3 {
					fmt.Println("expected at least 3 arguments, got", len(args))
					break
				}

				weightID, ok := weightToIDs[args[1]]
				if !ok {
					fmt.Println("unexpected weight", args[1])
					break
				}

				grade, err := strconv.ParseFloat(args[2], 64)
				if err != nil {
					fmt.Println("couldnt parse grade:", err)
					break
				}

				name := "manually added assignment"
				if len(args) == 4 {
					name = args[3]
				}

				section.Assignments = slices.Insert(section.Assignments, 0, &powerschool.Assignment{
					Name:       name,
					CategoryID: weightID,
					Percent:    grade,
				})

				fmt.Printf("after adding this assignment, your final grade is %.0f%%\n", section.FinalGrade(weightIDs))
			case "del", "delete":
				if len(args) < 2 {
					fmt.Println("expected 2 arguments, got", len(args))
					break
				}

				i, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Println("couldnt parse index:", err)
					break
				}

				if i+1 > len(section.Assignments) {
					fmt.Println("invalid index")
					break
				}

				section.Assignments = slices.Delete(section.Assignments, i, i+1)
				fmt.Printf("after deleting this assignment, your final grade is %.0f%%\n", section.FinalGrade(weightIDs))
			default:
				fmt.Println("unrecognized input")
			}
		}
	}
}
