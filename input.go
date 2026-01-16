package main

import (
	"os"
	"strings"
)

var sc chan os.Signal

func GetInput() string {
	inc := make(chan string, 1)

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
