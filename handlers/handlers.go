package handlers

import (
	"encoding/json"
	"fmt"
	"go-todo/models"
	"go-todo/renderer"
	"go-todo/services"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/stripe/stripe-go/v75"
	billingportalsession "github.com/stripe/stripe-go/v75/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/webhook"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

type Handler struct {
	service *services.Service
	store   *sqlitestore.SqliteStore
	render  *renderer.Renderer
}

func NewHandler(service *services.Service, store *sqlitestore.SqliteStore, renderer *renderer.Renderer) *Handler {
	return &Handler{
		service: service,
		store:   store,
		render:  renderer,
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
		http.Error(w, "could not get session from request", http.StatusInternalServerError)
		return
	}

	user := getUserFromSession(session)

	if user != nil {
		// user already logged in
		noCacheRedirect(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, userErrors, err := h.service.Login(email, password)
	if user == nil {
		loginFormProps := renderer.NewLoginFormProps(userErrors.EmailErrors, userErrors.PasswordErrors)
		basePageProps := renderer.NewBasePageProps(nil)
		homePageProps := renderer.NewHomePageProps(basePageProps, nil, &loginFormProps)
		bytes, err := h.render.HomePage(homePageProps)
		w.Write(bytes)
		return
	}
	// TODO resend user login page with user errors
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
	homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps)
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

	s, _ := checkoutsession.Get(checkoutSessionID, nil)
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

	s, _ := checkoutsession.New(params)
	// handle error?

	// Then redirect to the URL on the Checkout Session
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func (h *Handler) CreateCustomerPortalSession(w http.ResponseWriter, r *http.Request) {
	userSession, _ := h.store.Get(r, "user-session")

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
		Customer:  stripe.String(user.StripeCustomerID),
		ReturnURL: stripe.String("http://localhost:3000/"),
	}

	s, _ := billingportalsession.New(params)

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

	// TODO add writeheaders before the returns in this switch statement
	switch event.Type {
	case "checkout.session.completed":
		log.Println("Handling checkout.session.completed event for completed checkouts.")

		var checkoutSession stripe.CheckoutSession

		err = json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			log.Println("Error unmarshalling checkout.session.completed webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) completed checkout", checkoutSession.Customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.paid":
		log.Println("Handling invoice.paid event for payments.")
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			log.Printf("Failed to parse invoice paid webhook")
			return
		}

		customerID := invoice.Customer.ID
		customerEmail := invoice.Customer.Email

		user, err := h.service.GetUserByEmail(customerEmail)
		if err != nil {
			log.Println("error finding customer by customer email")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			log.Println("no customer by that email address")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userDoesNotAlreadyHaveAStripeID := user.StripeCustomerID == ""
		if userDoesNotAlreadyHaveAStripeID {
			err = h.service.AddStripeIDToUser(user.ID, customerID)
			if err != nil {
				// URGENT TODO will need a notification for this
				log.Printf("customer with ID of %s paid invoice but customer ID could not be addedto user", customerID)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		// TODO find a way to handle this better

		err = h.service.UpdateUserPaymentStatus(user.ID, true)
		if err != nil {
			log.Printf("customer %s paid their invoice but could not update the user record", customerID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return

	case "invoice.payment_failed":
		log.Println("Handling invoice.payment_failed event for failed payments.")

		var invoice stripe.Invoice

		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			log.Printf("Failed to unmarshal invoice.payment_failed webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		customerStripeID := invoice.Customer.ID
		if customerStripeID == "" {
			log.Println("cant get customer id from invoice.payment_failed webhook data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("invoice payment failed for customer ID: %s \n", customerStripeID)

		user, err := h.service.GetUserByStripeID(customerStripeID)
		if err != nil {
			log.Println("error searching for customer by stripe ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			log.Println("could not find matching user for that stripe ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// deactivate user premium status
		err = h.service.UpdateUserPaymentStatus(user.ID, false)
		if err != nil {
			log.Println("could not update customer payment status")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("downgraded plan for user: %s", user.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.updated":
		// Check subscription.items.data[0].price attribute and grant/revoke access accordingly.
		log.Println("Handling customer.subscription.updated event for price changes.")
		// cancel plan in customer billing portal triggers this event first
		log.Println(event.Data.Object)

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			log.Println("Failed to parse customer.subscription.updated webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			log.Println("Error trying to get customer by stripe ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			log.Println("No user by that stripe ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer: %s updated subscription", customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.deleted":
		// after the remaining days remining on the cancelled plan have elapsed, i assume this webhook gets called
		// Revoke customer's access to the product.
		log.Println("Handling customer.subscription.deleted event for subscription cancellations.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			log.Println("Could not unmarshal event data to customer struct")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		customerID := customer.ID
		if customerID == "" {
			log.Printf("cant get customer if from webhook event data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = h.service.UpdateUserPaymentStatus(customerID, false)
		if err != nil {
			log.Printf("Could not set is_paid_user to false, subscription deleted")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.paused":
		// Revoke customer's access until subscription resumes.
		log.Println("Handling customer.subscription.paused event for paused subscriptions.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			log.Println("Error unmarshalling customer.subscription.paused webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			log.Println("Error getting customer by stripe ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			log.Printf("no user exists with the ID of %s", customer.ID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("User (%s) paused their subscription", user.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.resumed":
		// Grant customer access when subscription resumes.
		log.Println("Handling customer.subscription.resumed event for resumed subscriptions.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			log.Println("Error unmarshalling customer.subscription.resumed webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			log.Println("Error getting customer by stripe ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			log.Printf("no user exists with the ID of %s", customer.ID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("User (%s) resumed their subscription", user.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "payment_method.attached":
		// Handle payment method attachment.
		log.Println("Handling payment_method.attached event for payment method attachment.")

		var paymentMethod stripe.PaymentMethod

		err = json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			log.Println("Error unmarshalling payment_method.attached webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) updated their payment method", paymentMethod.Customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "payment_method.detached":
		// Handle payment method detachment.
		log.Println("Handling payment_method.detached event for payment method detachment.")

		var paymentMethod stripe.PaymentMethod

		err = json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			log.Println("Error unmarshalling payment_method.detached webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) detached their payment method", paymentMethod.Customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.updated":
		// Check and update default payment method information.
		log.Println("Handling customer.updated event for default payment method updates.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			log.Println("Error unmarshalling customer.updated webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) updated", customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.tax_id.created", "customer.tax_id.deleted", "customer.tax_id.updated":
		// Handle tax ID related events.
		log.Println("Handling tax ID related event:", event)

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			log.Println("Error unmarshalling customer.tax_id.created webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) added tax ID", customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "billing_portal.configuration.created", "billing_portal.configuration.updated":
		// Handle billing portal configuration events.
		log.Println("Handling billing portal configuration event:", event)
		w.WriteHeader(http.StatusOK)
		return

	case "billing_portal.session.created":
		// Handle billing portal session creation.
		log.Println("Handling billing portal session created event.")
		w.WriteHeader(http.StatusOK)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
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
