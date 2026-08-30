package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/unmistakenly/PSGradeUtility/cli"
	"github.com/unmistakenly/PSGradeUtility/powerschool"
)

func main() {
	instance := flag.String("instance", "", "your district's PowerSchool public-portal base URL, e.g. https://myps.<district>.org (required)")
	low := flag.Float64("low-weight", cli.DefaultLowWeight, "weight of the Low grading category (PowerSchool's common scheme is 20/30/50 — override if your district differs)")
	mid := flag.Float64("mid-weight", cli.DefaultMidWeight, "weight of the Mid grading category")
	high := flag.Float64("high-weight", cli.DefaultHighWeight, "weight of the High grading category")
	flag.Parse()

	if *instance == "" {
		if env := os.Getenv("PSGRADE_INSTANCE"); env != "" {
			*instance = env
		} else {
			fmt.Println("error: no PowerSchool instance configured — pass -instance <url> or set PSGRADE_INSTANCE")
			os.Exit(1)
		}
	}

	client := powerschool.NewClient(*instance)
	session := cli.NewSession(client, *low, *mid, *high)

	if err := session.Run(); err != nil && !errors.Is(err, cli.ErrQuit) {
		fmt.Println(err)
		os.Exit(1)
	}
}
