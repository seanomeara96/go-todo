package main

import (
	"go-todo/internal/cli"
	"go-todo/internal/db"
	"go-todo/internal/repositories"
	"go-todo/internal/server/cache"
	"go-todo/internal/services"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("./configs/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		// You can choose to handle the error here or exit the program.
	}

	db, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to datbase %v", err)
	}
	defer db.Close()

	defaultExpiration := 5 * time.Minute
	cleanupInterval := 10 * time.Minute

	userCache := cache.NewUserCache(defaultExpiration, cleanupInterval)
	todoCache := cache.NewTodoCache(defaultExpiration, cleanupInterval)

	caches := &cache.Caches{
		UserCache: userCache,
		TodoCache: todoCache,
	}

	command := cli.New(services.NewService(repositories.NewRepository(db), caches))
	err = command.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
