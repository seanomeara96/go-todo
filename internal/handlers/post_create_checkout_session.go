package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	checkoutsession "github.com/stripe/stripe-go/v75/checkout/session"
)

func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))

	if err != nil {

		return err
	}

	if user == nil {
		return h.Logout(w, r)

	}

	stripeKey := os.Getenv(STRIPE_API_KEY)
	if stripeKey == "" {
		return fmt.Errorf("no stripe key")
	}

	stripe.Key = stripeKey

	priceId := "price_1NlpMHJ6hGciURAFUvHsGcdM"

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return fmt.Errorf("no domain key")
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
		return fmt.Errorf("Could not generate new checkout session for user (%s)", user.ID)
	}
	// handle error?

	infoMsg := fmt.Sprintf("User (%s) initiated checkout", user.ID)
	h.logger.Info(infoMsg)

	return noCacheRedirect(s.URL, w, r)
}
