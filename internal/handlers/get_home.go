package handlers

import (
	"fmt"
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return err
	}

	canCreateNewTodo := false
	var list []*models.Todo

	if user != nil {
		/*
			if user is logged in we need to get their todos
			and whether they have permission to create a new todo
		*/
		list, err = h.service.GetUserTodoList(user.ID)
		if err != nil {
			return fmt.Errorf("could not get user list of todos, %w", err)
		}

		canCreateNewTodo, err = h.service.UserCanCreateNewTodo(user, list)
		if err != nil {
			return fmt.Errorf("cannot determine whether user can create new todo, %w", err)
		}

	}

	basePageProps := renderer.NewBasePageProps(user)
	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo, nil)
	noErrors := []string{}
	loginFormProps := renderer.NewLoginFormProps(noErrors, noErrors)
	homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps, loginFormProps)

	bytes, err := h.render.HomePage(homePageProps)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	if user != nil {
		infoMsg := fmt.Sprintf("User (%s) loaded their todo list", user.ID)
		h.logger.Info(infoMsg)
	}
	return nil
}
