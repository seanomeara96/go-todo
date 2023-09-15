package handlers

import (
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/services"
	"net/http"
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

	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

type Handler struct {
	service  *services.Service
	store    *sqlitestore.SqliteStore
	render *renderer.Renderer
}

func NewHandler(service *services.Service, store *sqlitestore.SqliteStore, renderer *renderer.Renderer) *Handler {
	return &Handler{
		service:  service,
		store:    store,
		renderer: renderer,
	}
}

func getUserFromSession(s *sessions.Session) *models.User {
	val := s.Values["user"]
	var user = models.User{}
	user, ok := val.(models.User)
	if !ok {
		return nil
	}
	return &user
}

func noCacheRedirect(w http.ResponseWriter, r *http.Request) {
	// Set cache-control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect the user to a new URL
	http.Redirect(w, r, "/", http.StatusSeeOther)
}


func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		panic(err)
	}

	user := getUserFromSession(session)

	if user != nil {
		// user already logged in
		http.Redirect(w, r, "", http.StatusFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err = h.service.Login(email, password)
	if err != nil {
		http.Error(w, "Something went wrong during login", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Incorrect credentials", http.StatusBadRequest)
		return
	}

	session.Values["user"] = *user
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	err = h.store.Delete(r, w, session)
	if err != nil {
		http.Error(w, "could not delete session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}


func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)
	canCreateNewTodo := false
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

		canCreateNewTodo = (!userIsPayedUser && len(list) < 10) || userIsPayedUser
	}

	basePageProps := renderer.NewBasePageProps(user)
	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	homePageProps := renderer.NewHomePageProps(basePageProps)
	bytes, err := h.render.HomePage(homePageProps)
	if err != nil {
		http.Error(w, "could not render home-logged-out", http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

func (h *Handler) SignupPage(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Could not get user from store", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)
	if user != nil {
		noCacheRedirect(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	signupPageProps := renderer.NewSignupPageProps(basePageProps)
	signupPageBytes, err := h.render.Signup(signupPageProps)
	if err != nil {
		http.Error(w, "could not render sigup page", http.StatusInternalServerError)
		return
	}
	w.Write(signupPageBytes)
}

func (h *Handler) UpgradePage(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get user from store", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)
	if user == nil {
		noCacheRedirect(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	upgradePageProps := renderer.NewUpgradePageProps(basePageProps)
	upgradePageBytes, err := h.render.Upgrade(upgradePageProps)
	if err != nil {
		http.Error(w, "could not render upgrade page", http.StatusInternalServerError)
		return
	}

	w.Write(upgradePageBytes)
}

func (h *Handler) SuccessPage(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(userSession)
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
	bytes, err := h.render.Success(successPageProps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func (h *Handler) CancelPage(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)
	if user == nil {
		noCacheRedirect(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	cancelPageProps := renderer.NewCancelPageProps(basePageProps)
	bytes, err := h.render.Cancel(cancelPageProps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not retrieve user session", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(userSession)
	if user == nil {
		noCacheRedirect(w, r)
		return
	}

	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		http.Error(w, "no stripe api key", http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey

	priceId := "price_1NlpMHJ6hGciURAFUvHsGcdM"

	successUrl := "http://localhost:3000/success?session_id={CHECKOUT_SESSION_ID}"
	canceledUrl := "http://localhost:3000/canceled"
	params := &stripe.CheckoutSessionParams{
		CustomerEmail: stripe.String(user.Email),
		SuccessURL:    &successUrl,
		CancelURL:     &canceledUrl,
		Mode:          stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price: stripe.String(priceId),
				// For metered billing, do not pass quantity
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, _ := session.New(params)
	// handle error?

	// Then redirect to the URL on the Checkout Session
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}



func (h *Handler) CreateCustomerPortalSession(w http.ResponseWriter, r *http.Request){
	userSession, _ := h.store.Get("user-session")

	user := getUserFromSession(userSession)
	if user == nil {
		// do something
		return
	}

	user, _ = h.service.GetUserByID(user.ID)

	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		// do something
		return
	}

	stripe.Key = stripeKey

	params := &stripe.BillingPortalSessionParams{
		Customer: stripe.String(user.StripeCustomerID),
		ReturnURL: stripe.String("http://localhost:3000/"),
	  };
	  
	result, _ := session.New(params);
	
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
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
	if stripeWebhookSecret == "" {
		http.Error(w, "could not find stripe webhook secret in env", http.StatusInternalServerError)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		fmt.Println("checkout session completed")
	case "invoice.paid":
		fmt.Println("invoice paid")
		checkoutID := event.Data.Object["id"]
		customerID := event.Data.Object["customer"]
		customerEmail := event.Data.Object["customer_email"]

		fmt.Println(checkoutID, customerID, customerEmail)
	case "invoice.payment_failed":
		// deactivate user premium status
	case "customer.subscription.updated":
        // Check subscription.items.data[0].price attribute and grant/revoke access accordingly.
        fmt.Println("Handling customer.subscription.updated event for price changes.")
    case "customer.subscription.deleted":
        // Revoke customer's access to the product.
        fmt.Println("Handling customer.subscription.deleted event for subscription cancellations.")
    case "customer.subscription.paused":
        // Revoke customer's access until subscription resumes.
        fmt.Println("Handling customer.subscription.paused event for paused subscriptions.")
    case "customer.subscription.resumed":
        // Grant customer access when subscription resumes.
        fmt.Println("Handling customer.subscription.resumed event for resumed subscriptions.")
    case "payment_method.attached":
        // Handle payment method attachment.
        fmt.Println("Handling payment_method.attached event for payment method attachment.")
    case "payment_method.detached":
        // Handle payment method detachment.
        fmt.Println("Handling payment_method.detached event for payment method detachment.")
    case "customer.updated":
        // Check and update default payment method information.
        fmt.Println("Handling customer.updated event for default payment method updates.")
    case "customer.tax_id.created", "customer.tax_id.deleted", "customer.tax_id.updated":
        // Handle tax ID related events.
        fmt.Println("Handling tax ID related event:", event)
    case "billing_portal.configuration.created", "billing_portal.configuration.updated":
        // Handle billing portal configuration events.
        fmt.Println("Handling billing portal configuration event:", event)
    case "billing_portal.session.created":
        // Handle billing portal session creation.
        fmt.Println("Handling billing portal session created event.")
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


func (h *Handler) userCanCreateNewTodo(user *models.User, list []*models.Todo) (bool, error) {
	userIsPaidUser, err := h.service.UserIsPaidUser(user.ID)
	if err != nil {
		return false, err
	}

	canCreateNewTodo := (!userIsPaidUser && len(list) < 10) || userIsPaidUser

	return canCreateNewTodo, nil
}

func (h *Handler) AddTodo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusInternalServerError)
		return
	}

	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get user session", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)

	if user == nil {
		http.Error(w, "Usermust be logged in", http.StatusForbidden)
		return
	}

	_, err = h.service.CreateTodo(user.ID, r.FormValue("description"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	userIsPayedUser, err := h.service.UserIsPaidUser(user.ID)
	if err != nil {
		http.Error(w, "error  determingin user paymnet status", http.StatusInternalServerError)
		return
	}
	canCreateNewTodo := (!userIsPayedUser && len(list) < 10) || userIsPayedUser

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	todoList, err := h.render.TodoList(todoListProps)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}
	w.Write(todoList)
}
func (h *Handler) UpdateTodoDescription(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) RemoveTodo(w http.ResponseWriter, r *http.Request) {

	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)
	if user == nil {
		http.Error(w, "user must be logged in", http.StatusForbidden)
		return
	}

	// get user from database
	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		http.Error(w, "could not find user", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	todoIDString := vars["id"]

	todoID, err := strconv.Atoi(todoIDString)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, err := h.service.GetTodoByID(todoID)
	if err != nil {
		http.Error(w, "could not get todo", http.StatusInternalServerError)
		return
	}

	userIsNotAuthor := user.ID != todo.UserID
	if userIsNotAuthor {
		http.Error(w, "not authorized", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteTodo(todo.ID)
	if err != nil {
		http.Error(w, "could not remove todo", http.StatusInternalServerError)
		return
	}

	// TODO this is duplicate code
	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	userCanCreateNewTodo, err := h.userCanCreateNewTodo(user, list)
	if err != nil {
		http.Error(w, "error determining user payment status", http.StatusInternalServerError)
		return
	}

	props := renderer.NewTodoListProps(list, userCanCreateNewTodo)
	todoListBytes, err := h.render.TodoList(props)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}
	w.Write(todoListBytes)
}

func (h *Handler) UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)

	if user == nil {
		http.Error(w, "user must be logged in", http.StatusForbidden)
		return
	}

	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		http.Error(w, "trouble finding that user", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	idParam := vars["id"]
	todoID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, err := h.service.UpdateTodoStatus(user.ID, todoID)
	if err != nil {
		http.Error(w, "could not update todo", http.StatusInternalServerError)
		return
	}

	if todo == nil {
		http.Error(w, "service did not return a todo item", http.StatusInternalServerError)
		return
	}

	todoBytes, err := h.render.Todo(todo)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

	w.Write(todoBytes)
}



func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse signup form", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// TODO sanitize and clean input

	_, err = h.service.NewUser(name, email, password)
	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request) {

// 	// stripe code

// }
