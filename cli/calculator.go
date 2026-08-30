package cli

import (
	"errors"
	"fmt"
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
view

edit <index> <0-100>
restore <index>`

// gradeCalculator will start its own input loops
func (s *Session) gradeCalculator() error {
	// this will have a lot of shared code with [showAllGrades]
	if s.ticket == "" {
		return ErrNotSignedIn
	}

	data, err := s.getFullDecodedResponse()
	if err != nil {
		return err
	}

	quarterStart, quarterEnd := data.Response.Return.Data.GetCurrentQuarter()
	classes, weightIDs := extractInfoFromResponse(data, quarterStart, quarterEnd, s.weights())

	sclasses := make([]*powerschool.Section, 0, len(classes))
	for _, c := range classes {
		if len(c.Assignments) == 0 {
			continue
		}
		sclasses = append(sclasses, c)
	}

	printClasses := func() {
		for i, c := range sclasses {
			grade, err := c.FinalGrade(weightIDs)
			fmt.Printf("[%d] %s (%s - %d grades)\n", i, c.ClassName, formatGrade(grade, err), len(c.Assignments))
		}
	}
	printClasses()

	for {
		fmt.Print("\n>> ")
		input := s.GetInput()

		switch input {
		case "":
		case "h", "help":
			fmt.Println(CalcMenuHelpText)
			fmt.Println()
			printClasses()
		case "q", "quit":
			return ErrQuit
		case "b":
			return nil
		default:
			i, err := strconv.Atoi(input)
			if err != nil || i < 0 || i > len(sclasses)-1 {
				fmt.Println("unrecognized input")
				break
			}
			var ref any
			if s.preferClassNames {
				ref = sclasses[i].ClassName
			} else {
				ref = i
			}
			if err := s.classCalculator(sclasses[i], weightIDs, ref); errors.Is(err, ErrQuit) {
				return ErrQuit
			}
		}
	}
}

func (s *Session) classCalculator(origSection *powerschool.Section, weightIDs map[int]string, ref any) error {
	// enforce access of assignments through section only, as it will otherwise cause a runtime error (TOTALLY didnt happen)
	section := func() *powerschool.Section {
		// deep copy of origAssignments, so as to not modify it
		assignments := make([]*powerschool.Assignment, len(origSection.Assignments), len(origSection.Assignments)+3)
		for i := range assignments {
			orig := origSection.Assignments[i]
			assignments[i] = &powerschool.Assignment{
				Name:       orig.Name,
				CategoryID: orig.CategoryID,
				Grade:      orig.Grade,
			}
		}
		return &powerschool.Section{
			Assignments: assignments,
			Low:         origSection.Low,
			Mid:         origSection.Mid,
			High:        origSection.High,
		}
	}()

	// i couldnt think of any other way to do this
	weightToIDs := make(map[string]int, 3)
	for id, weight := range weightIDs {
		weight = strings.ToLower(weight) // ToLower to match user input
		if _, ok := weightToIDs[weight]; !ok {
			weightToIDs[weight] = id
		}
	}

	note := func(a *powerschool.Assignment) string {
		if a.Edited {
			return " (Edited)"
		}
		return ""
	}

	printAssignments := func() {
		for i, a := range section.Assignments {
			fmt.Printf("[%d] %s - %d%% (%s)%s\n", i, a.Name, a.Grade, weightIDs[a.CategoryID], note(a))
		}
	}
	printAssignments()

	for {
		fmt.Printf("\n(%v) >> ", ref)
		input := s.GetInput()

		switch input {
		case "h", "help":
			fmt.Println(CalcHelpText)
		case "q", "quit":
			return ErrQuit
		case "b":
			return nil
		default:
			args := strings.SplitN(input, " ", 4)
			switch args[0] { // lol rip
			case "":
			case "v", "view":
				printAssignments()
				grade, err := section.FinalGrade(weightIDs)
				fmt.Printf("\nfinal grade: %s\n", formatGrade(grade, err))
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

				grade, err := strconv.ParseUint(args[2], 10, 64)
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
					Grade:      grade,
				})

				fg, ferr := section.FinalGrade(weightIDs)
				fmt.Printf("after adding this assignment, your final grade is %s\n", formatGrade(fg, ferr))
			case "d", "del", "delete":
				if len(args) < 2 {
					fmt.Println("expected 2 arguments, got", len(args))
					break
				}

				i, err := parseIndex(args, section.Assignments)
				if err != nil {
					fmt.Println(err)
					break
				}

				section.Assignments = slices.Delete(section.Assignments, i, i+1)
				fg, ferr := section.FinalGrade(weightIDs)
				fmt.Printf("after deleting this assignment, your final grade is %s\n", formatGrade(fg, ferr))
			case "e", "edit":
				if len(args) < 3 {
					fmt.Println("expected 3 arguments, got", len(args))
					break
				}

				i, err := parseIndex(args, section.Assignments)
				if err != nil {
					fmt.Println(err)
					break
				}

				grade, err := strconv.ParseUint(args[2], 10, 64)
				if err != nil {
					fmt.Println("couldnt parse grade:", err)
					break
				}

				section.Assignments[i].Edit(grade)
				fg, ferr := section.FinalGrade(weightIDs)
				fmt.Printf("after editing this assignment, your final grade is %s\n", formatGrade(fg, ferr))
			case "r", "restore":
				if len(args) < 2 {
					fmt.Println("expected 2 arguments, got", len(args))
					break
				}

				i, err := parseIndex(args, section.Assignments)
				if err != nil {
					fmt.Println(err)
					break
				}

				a := section.Assignments[i]
				if a.Edited {
					a.Restore()
					fg, ferr := section.FinalGrade(weightIDs)
					fmt.Printf("after changing this grade back to a %d%%, your final grade is %s\n", a.Grade, formatGrade(fg, ferr))
				} else {
					fmt.Println("this assignment hasnt had its grade edited")
				}
			default:
				fmt.Println("unrecognized input")
			}
		}
	}
}

func parseIndex(args []string, assignments []*powerschool.Assignment) (int, error) {
	i, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, fmt.Errorf("couldnt parse index: %w", err)
	}

	if i < 0 || i+1 > len(assignments) {
		return 0, ErrInvalidIndex
	}

	return i, nil
}
