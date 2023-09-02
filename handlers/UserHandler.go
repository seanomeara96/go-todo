package handlers

import (
	"go-todo/services"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
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

	successUrl := "https://example.com/success.html?session_id={CHECKOUT_SESSION_ID}"
	canceledUrl := "https://example.com/canceled.html"
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

// func (h *UserHandler) Upgrade(w http.ResponseWriter, r *http.Request) {

// 	// stripe code

// }
