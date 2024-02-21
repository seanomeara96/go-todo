package handlers

import (
	"fmt"
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
			h.logger.Error("Could not get user list of todos")
			h.logger.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		canCreateNewTodo, err = h.service.UserCanCreateNewTodo(user, list)
		if err != nil {
			h.logger.Error("Cannot determine whether user can create new todo")
			h.logger.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	}

	basePageProps := renderer.NewBasePageProps(user)
	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	noErrors := []string{}
	loginFormProps := renderer.NewLoginFormProps(noErrors, noErrors)
	homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps, loginFormProps)

	bytes, err := h.render.HomePage(homePageProps)
	if err != nil {
		h.logger.Error("could not render home-logged-out")
		h.logger.Debug(err.Error())
		http.Error(w, "could not render home-logged-out", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		h.logger.Error("could not write homepage")
		h.logger.Debug(err.Error())
		http.Error(w, "could not write home-logged-out", http.StatusInternalServerError)
		return
	}

	if user != nil {
		infoMsg := fmt.Sprintf("User (%s) loaded their todo list", user.ID)
		h.logger.Info(infoMsg)
	}
}
