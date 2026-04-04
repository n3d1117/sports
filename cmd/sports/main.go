package main

import (
	"os"

	"sports/internal/commands"
)

func main() {
	os.Exit(commands.Run(os.Args[1:], os.Stdout, os.Stderr))
}
