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

	stripe.Key = os.Getenv("STRIPE_API_KEY")

	s, _ := session.Get(
		checkoutSessionID,
		nil,
	)

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

func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {

	sess, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not retrieve user session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(sess)
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

	// Then redirect to the URL on the Checkout Session
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
