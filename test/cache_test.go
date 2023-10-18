package test

import (
	"go-todo/internal/cache"
	"go-todo/internal/logger"
	"go-todo/internal/models"
	"testing"
	"time"
)

func initCache() *cache.UserCache {
	logger := logger.NewLogger(0)
	return cache.NewUserCache(5*time.Minute, 10*time.Minute, logger)
}

func TestCacheUser(t *testing.T) {

	cache := initCache()

	users := []models.User{
		models.NewUser("id1", "name1", "email1@email.com", "password", false, ""),
		models.NewUser("id2", "name2", "email2@email.com", "password", false, ""),
		models.NewUser("id3", "name3", "email3@email.com", "password", false, ""),
		models.NewUser("id4", "name4", "email4@email.com", "password", false, ""),
		models.NewUser("id5", "name5", "email5@email.com", "password", false, ""),
		models.NewUser("id6", "name6", "email6@email.com", "password", false, ""),
		models.NewUser("id7", "name7", "email7@email.com", "password", false, ""),
	}

	for _, user := range users {
		cache.CacheUser(user)
	}

	cachedUser := cache.GetUserByID("id1")
	if cachedUser == nil || cachedUser.ID != "id1" {
		t.Error("user 1 was not cached")
	}

	cachedUser = cache.GetUserByEmail("email2@email.com")
	if cachedUser == nil || cachedUser.ID != "id2" {
		t.Error("user 2 was not cached")
	}

	updatedUser := users[0]
	updatedUser.Name = "Sean"
	updatedUser.Email = "sean@email.com"
	updatedUser.Password = "seanPassword"

	cache.CacheUser(updatedUser)

	cachedUser = cache.GetUserByID("id1")

	if cachedUser == nil || cachedUser.Name != "Sean" {
		t.Error("updates were not saved into existing cached user")
	}

}
