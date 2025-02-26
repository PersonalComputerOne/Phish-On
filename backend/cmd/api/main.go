package main

import (
	"log"

	"github.com/PersonalComputerOne/Phish-On/internal/routes"
	"github.com/PersonalComputerOne/Phish-On/pkg/db"
)

func main() {
	pool, err := db.Init()

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer pool.Close()

	r := routes.SetupRouter()

	r.Run(":8080")
}
