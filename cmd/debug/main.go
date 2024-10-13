package main

import (
	"log"

	"github.com/andymarkow/gophkeeper/internal/server/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(cfg)
}
