package cache

import (
	"go-todo/logger"
	"go-todo/models"

	"github.com/patrickmn/go-cache"
)

type UserCache struct {
	cache  *cache.Cache
	logger *logger.Logger
}

func NewUserCache(cache *cache.Cache, logger *logger.Logger) *UserCache {
	return &UserCache{
		cache:  cache,
		logger: logger,
	}
}

func (c *UserCache) getUsersFromCache() []models.User {
	users := []models.User{}
	usersCache, found := c.cache.Get("users")
	if !found {
		c.cache.Set("users", users, cache.DefaultExpiration)
		return users
	}
	cachedUsers, ok := usersCache.([]models.User)
	if !ok {
		return users
	}
	return cachedUsers
}

func (c *UserCache) CacheUser(user models.User) {
	cachedUsers := c.getUsersFromCache()

	for _, cachedUser := range cachedUsers {
		if user.ID == cachedUser.ID {
			return
		}
	}

	cachedUsers = append(cachedUsers, user)

	c.cache.Set("users", cachedUsers, cache.DefaultExpiration)

}

func (c *UserCache) GetUserByID(userID string) *models.User {
	cachedUsers := c.getUsersFromCache()
	for _, user := range cachedUsers {
		if user.ID == userID {
			return &user
		}
	}
	return nil
}

func (c *UserCache) GetUserByEmail(email string) *models.User {
	cachedUsers := c.getUsersFromCache()
	for _, user := range cachedUsers {
		if user.Email == email {
			return &user
		}
	}
	return nil
}

func (c *UserCache) GetUserByStripeID(userStripeID string) *models.User {
	cachedUsers := c.getUsersFromCache()
	for _, user := range cachedUsers {
		if user.StripeCustomerID == userStripeID {
			return &user
		}
	}
	return nil
}
