package main

import (
	"fmt"
	"os"

	"github.com/andymarkow/gophkeeper/internal/commands"
)

func main() {
	cmd, err := commands.NewCommand()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
