package main

import (
	"go-todo/internal/logger"
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

	var logLevel logger.LogLevel = 0
	if os.Getenv("env") == "prod" {
		logLevel = 1
		logFile, err := logger.SetOutputToFile()
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
	}

	if err = server.DefaultServer(logger.NewLogger(logLevel)).Serve(os.Getenv("port")); err != nil {
		log.Fatal(err)
	}
}
