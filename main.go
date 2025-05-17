package main

import (
	"github.com/joeycumines/simple-command-output-filter/internal/cli"
	"os"
)

func main() {
	os.Exit((&cli.CLI{
		Input:  os.Stdin,
		Output: os.Stdout,
		ErrOut: os.Stderr,
	}).Main(os.Args[1:]))
}
