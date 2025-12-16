package main

import (
	"fmt"
	"os"
	"strings"
)

// hm, a main menu actually sounds like a nice idea here!
func MainInteractive() error {
	var input string
	var err error
	var username, ticket, studentID string

	fmt.Println("hello! to see a list of commands, you can use \033[1mh\033[0m.")

	for {
		fmt.Print("\n> ")
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "h", "help":
			fmt.Println(HelpText)
		case "q", "quit":
			return nil
		case "s":
			u, t, s, err := signIn()
			if err != nil {
				fmt.Println(err)
				break
			}
			username, ticket, studentID = u, t, s
			fmt.Println("\nsigned in as", username)
		case "u":
			if username == "" {
				fmt.Println("you arent signed in")
				break
			}
			fmt.Println("currently signed in as", username)
		case "a":
			if err = showAllGrades(ticket, studentID); err != nil {
				fmt.Println(err)
			}
		case "c":
			if err = gradeCalculator(ticket, studentID); err != nil {
				fmt.Println(err)
			}
		default:
			fmt.Println("unrecognized input")
		}
	}
}

func main() {
	if err := MainInteractive(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
