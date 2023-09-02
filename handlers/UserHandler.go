package handlers

import (
	"go-todo/services"
	"io"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/webhook"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse signup form", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// TODO sanitize and clean input

	_, err = h.userService.NewUser(name, email, password)
	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (h *UserHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {

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

func (h *UserHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		// do somehting here
	case "invoice.paid":
		//do somthing here
	case "invoice.payment_failed":
		// do something here
	default:
		// something went horribly wrong
	}

	/*
			The minimum event types to monitor:
		Event name	Description
		checkout.session.completed	Sent when a customer clicks the Pay or Subscribe button in Checkout, informing you of a new purchase.
		invoice.paid	Sent each billing interval when a payment succeeds.
		invoice.payment_failed	Sent each billing interval if there is an issue with your customer’s payment method.*/

}

// func (h *UserHandler) Upgrade(w http.ResponseWriter, r *http.Request) {

// 	// stripe code

// }
