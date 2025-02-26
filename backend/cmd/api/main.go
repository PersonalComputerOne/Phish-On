package main

import (
	"github.com/PersonalComputerOne/Phish-On/internal/routes"
)

func main() {
	r := routes.SetupRouter()

	r.Run(":8080")
}
