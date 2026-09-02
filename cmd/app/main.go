package main

import (
	"log"

	"flight-service-api/internal/api"
)

func main() {
	log.Println("Application start!")

	api.StartServer()

	log.Println("Application terminated!")
}
