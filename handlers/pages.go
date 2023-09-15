package handlers

import (
	"fmt"
	"go-todo/models"
	"go-todo/renderer"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/webhook"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	basePageProps := renderer.NewBasePageProps(user)

	var list []*models.Todo
	if user != nil {
		list, err = h.service.GetUserTodoList(user.ID)
		if err != nil {
			http.Error(w, "could not get users list of todos", http.StatusInternalServerError)
			return
		}

		userIsPayedUser, err := h.service.UserIsPaidUser(user.ID)
		if err != nil {
			http.Error(w, "error determining user payment status", http.StatusInternalServerError)
			return
		}

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

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user != nil {
		noCacheRedirect(w, r)
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

func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get user from store", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user == nil {
		noCacheRedirect(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	upgradePageProps := renderer.NewUpgradePageProps(basePageProps)
	upgradePageBytes, err := h.renderer.Upgrade(upgradePageProps)
	if err != nil {
		http.Error(w, "could not render upgrade page", http.StatusInternalServerError)
		return
	}

	w.Write(upgradePageBytes)
}

func (h *Handler) Success(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(userSession)
	if user == nil {
		noCacheRedirect(w, r)
		return
	}

	checkoutSessionID := r.URL.Query().Get("session_id")

	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		http.Error(w, "no stripe api key", http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey

	s, _ := session.Get(checkoutSessionID, nil)
	// handle error ?

	// get user from db
	user, err = h.service.GetUserByEmail(s.CustomerEmail)
	if err != nil {
		http.Error(w, "could not find user", http.StatusInternalServerError)
		return
	}

	err = h.service.AddStripeIDToUser(user.ID, s.Customer.ID)
	if err != nil {
		http.Error(w, "could not add customer details to user", http.StatusInternalServerError)
		return
	}

	if s.PaymentStatus == "paid" {
		err = h.service.UpdateUserPaymentStatus(user.ID, true)
		if err != nil {
			http.Error(w, "could not update user payment status", http.StatusInternalServerError)
			return
		}
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

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(session)
	if user == nil {
		noCacheRedirect(w, r)
		return
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

