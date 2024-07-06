package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	checkoutsession "github.com/stripe/stripe-go/v75/checkout/session"
)

func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	stripeKey := os.Getenv(STRIPE_API_KEY)
	if stripeKey == "" {
		http.Error(w, "no stripe api key", http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey

	priceId := "price_1NlpMHJ6hGciURAFUvHsGcdM"

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		h.logger.Error("Expected a domain set in ENV vars")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	successUrl := domain + "/success?session_id={CHECKOUT_SESSION_ID}"
	canceledUrl := domain + "/canceled"
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

	s, err := checkoutsession.New(params)
	if err != nil {
		errMsg := fmt.Sprintf("Could not generate new checkout session for user (%s)", user.ID)
		h.logger.Error(errMsg)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// handle error?

	// Then redirect to the URL on the Checkout Session
	http.Redirect(w, r, s.URL, http.StatusSeeOther)

	infoMsg := fmt.Sprintf("User (%s) initiated checkout", user.ID)
	h.logger.Info(infoMsg)
}
