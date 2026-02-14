package main

import (
	"fmt"
	"log"

	"github.com/jonasyke/gator/internal/config"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}

	err = cfg.SetUser("Jonathan")
	if err != nil {
		log.Fatalf("could not set user name: %v", err)
	}

	final, err := config.Read()
	if err != nil {
		log.Fatalf("failed to read final state: %v", err)
	}

	fmt.Println(final)
}
