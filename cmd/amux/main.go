package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: amux <command>")
		os.Exit(2)
	}
	command := os.Args[1]
	switch command {
	case "test":
		if err := runTest(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "agent":
		if err := runAgent(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "git":
		if err := runGit(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(2)
	}
}
