package handlers

import (
	"fmt"
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/services"
	"net/http"

	"github.com/michaeljs1990/sqlitestore"
)

type PageHandler struct {
	renderer    *renderer.Renderer
	store       *sqlitestore.SqliteStore
	userService *services.UserService
	todoService *services.TodoService
}

func NewPageHandler(userService *services.UserService, todoService *services.TodoService, renderer *renderer.Renderer, store *sqlitestore.SqliteStore) *PageHandler {
	return &PageHandler{
		renderer:    renderer,
		store:       store,
		userService: userService,
		todoService: todoService,
	}
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	basePageProps := renderer.NewBasePageProps(user)

	var list []*models.Todo
	if user != nil {
		list, err = h.todoService.GetUserTodoList(user.ID)
		if err != nil {
			http.Error(w, "could not get users list of todos", http.StatusInternalServerError)
			return
		}
		userIsPayedUser := h.userService.UserIsPayedUser(user.ID)
		canCreateNewTodo := (!userIsPayedUser && len(list) < 10) || userIsPayedUser

		todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
		homePageLoggedInProps := renderer.NewHomePageLoggedInProps(basePageProps, todoListProps)
		homePageBytes, err := h.renderer.HomePageLoggedIn(homePageLoggedInProps)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Could not render homepage", http.StatusInternalServerError)
			return
		}
		w.Write(homePageBytes)
		return
	}
	homePageLoggedOutProps := renderer.NewHomePageLoggedOutProps(basePageProps)
	homePageLoggedOutBytes, err := h.renderer.HomePageLoggedOut(homePageLoggedOutProps)
	if err != nil {
		http.Error(w, "could not render home-logged-out", http.StatusInternalServerError)
		return
	}
	w.Write(homePageLoggedOutBytes)
}

func (h *PageHandler) Signup(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	signupPageProps := renderer.NewSignupPageProps(basePageProps)
	signupPageBytes, err := h.renderer.Signup(signupPageProps)
	if err != nil {
		http.Error(w, "could not render sigup page", http.StatusInternalServerError)
		return
	}
	w.Write(signupPageBytes)
}

func (h *PageHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)

	basePageProps := renderer.NewBasePageProps(user)
	upgradePageProps := renderer.NewUpgradePageProps(basePageProps)
	upgradePageBytes, err := h.renderer.Upgrade(upgradePageProps)
	if err != nil {
		http.Error(w, "could not render upgrade page", http.StatusInternalServerError)
		return
	}
	w.Write(upgradePageBytes)

}

func (h *PageHandler) Success(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
	user := GetUserFromSession(session)

	if user == nil {
		// TODO handle this properly
		fmt.Println("handle this properly")
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
	basePageProps := renderer.NewBasePageProps(user)
	successPageProps := renderer.NewSuccessPageProps(basePageProps)
	bytes, err := h.renderer.Success(successPageProps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func (h *PageHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)

	if user == nil {
		//...
	}

	basePageProps := renderer.NewBasePageProps(user)
	cancelPageProps := renderer.NewCancelPageProps(basePageProps)
	bytes, err := h.renderer.Cancel(cancelPageProps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
