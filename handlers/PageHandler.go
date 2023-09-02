package handlers

import (
	"fmt"
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/services"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/michaeljs1990/sqlitestore"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/webhook"
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

func noCacheRedirect(w http.ResponseWriter, r *http.Request) {
	// Set cache-control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect the user to a new URL
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := GetUserFromSession(session)

	if user == nil {
		noCacheRedirect(w, r)
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

func (h *PageHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {

	stripe.Key = os.Getenv("STRIPE_API_KEY")
	priceId := "price_1NlpMHJ6hGciURAFUvHsGcdM"

	successUrl := "https://localhost:3000/success?session_id={CHECKOUT_SESSION_ID}"
	canceledUrl := "https://localhost:3000/canceled"
	params := &stripe.CheckoutSessionParams{
		SuccessURL: &successUrl,
		CancelURL:  &canceledUrl,
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price: stripe.String(priceId),
				// For metered billing, do not pass quantity
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, _ := session.New(params)

	// Then redirect to the URL on the Checkout Session
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func (h *PageHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		http.Error(w, "could not get stripe key from env", http.StatusInternalServerError)
		return
	}
	stripe.Key = stripeKey
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("bad request from stripe")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		fmt.Println("cehckout session completed")
		// do somehting here
		// activate  user premium status
	case "invoice.paid":
		fmt.Println("invoice paid")
		//do somthing here
		// activate user premium status
	case "invoice.payment_failed":
		// deactivate user premium status
	default:
		// something else happened
	}

	/*
			The minimum event types to monitor:
		Event name	Description
		checkout.session.completed	Sent when a customer clicks the Pay or Subscribe button in Checkout, informing you of a new purchase.
		invoice.paid	Sent each billing interval when a payment succeeds.
		invoice.payment_failed	Sent each billing interval if there is an issue with your customer’s payment method.*/

}
