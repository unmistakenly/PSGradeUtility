package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
)

// just change this if you're in another school district
const PowerSchoolInstance = "https://myps.horrycountyschools.net"

// i actually didnt know fmt.Scanln seperated inputs with spaces. funky!
var stdinReader = bufio.NewReader(os.Stdin)

// hm, a main menu actually sounds like a nice idea here!
func MainInteractive() error {
	var username, ticket, studentID string

	inc = make(chan string, 1)
	sc = make(chan os.Signal, 1)

	preferClassNames := true
	signal.Notify(sc, os.Interrupt)

	fmt.Println("hello! to see a list of commands, you can use \033[1mh\033[0m.")

	for {
		fmt.Print("\n> ")

		switch GetInput() {
		case "":
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
		case "o":
			if username == "" {
				fmt.Println("you arent signed in")
				break
			}
			username, ticket, studentID = "", "", ""
			fmt.Println("signed out")
		case "a":
			if err := showAllGrades(ticket, studentID); err != nil {
				fmt.Println(err)
			}
		case "c":
			if err := gradeCalculator(ticket, studentID, preferClassNames); err != nil {
				fmt.Println(err)
			}
		case "p":
			if preferClassNames {
				fmt.Println("now preferring class indexes over names")
			} else {
				fmt.Println("now preferring class names over indexes")
			}
			preferClassNames = !preferClassNames
		case "i":
			if username == "" {
				fmt.Println("you arent signed in")
			} else {
				fmt.Println("currently signed in as", username)
			}
			fmt.Println("\nprefer class names over indexes:", preferClassNames)
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
