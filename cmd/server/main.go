package main

import (
	"fmt"
	"os"

	"github.com/andymarkow/gophkeeper/internal/server"
)

func main() {
	srv, err := server.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "server.NewServer: %v", err)
	}

	if err := srv.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "server.Run: %v", err)
	}
}
