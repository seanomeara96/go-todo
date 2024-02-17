package main

import (
	"go-todo/internal/server"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("./configs/.env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
		// You can choose to handle the error here or exit the program.
	}
	if os.Getenv("ENV") == "" || os.Getenv("PORT") == "" {
		log.Fatal("Expected a PORT and ENV var")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Cant find port in env vars")
	}
	env := os.Getenv("ENV")

	err = server.Serve(env, port)
	if err != nil {
		log.Fatal(err)
	}
}
