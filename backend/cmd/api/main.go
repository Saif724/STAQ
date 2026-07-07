package main

import (
	"fmt"
	"log"

	"github.com/Saif724/STAQ/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.App.Env)
}