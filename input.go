package main

import (
	"os"
	"strings"
)

var (
	inc chan string
	sc  chan os.Signal
)

func GetInput() string {
	go func() {
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			return
		}
		inc <- strings.ToLower(strings.TrimSpace(input))
	}()

	select {
	case s := <-inc:
		return s
	case <-sc:
		return ""
	}
}
