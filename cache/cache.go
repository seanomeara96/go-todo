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

	for i, _ := range cachedUsers {
		if user.ID == cachedUsers[i].ID {
			// if the new user has updated properties I want to make sure that its updated
			cachedUsers[i] = user
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

type TodoCache struct {
	cache  *cache.Cache
	logger *logger.Logger
}

func (c *TodoCache) getTodosFromCache() []models.Todo {
	todos := []models.Todo{}
	todoCache, found := c.cache.Get("todos")
	if !found {
		c.cache.Set("todos", todos, cache.DefaultExpiration)
		return todos
	}
	cachedTodos, ok := todoCache.([]models.Todo)
	if !ok {
		return todos
	}
	return cachedTodos
}

func (c *TodoCache) GetTodosByUserID(userID string) []models.Todo {
	todos := c.getTodosFromCache()
	userTodos := []models.Todo{}
	for i := 0; i < len(todos); i++ {
		if todos[i].UserID == userID {
			userTodos = append(userTodos, todos[i])
		}
	}
	return userTodos
}

func (c *TodoCache) GetTodoByID(todoID int) *models.Todo {
	todos := c.getTodosFromCache()
	for _, todo := range todos {
		if todo.ID == todoID {
			return &todo
		}
	}
	return nil
}
