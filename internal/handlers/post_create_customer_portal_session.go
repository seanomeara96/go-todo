package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	billingportalsession "github.com/stripe/stripe-go/v75/billingportal/session"
)

func (h *Handler) CreateCustomerPortalSession(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return fmt.Errorf("Could not get user from session on create customer portal session hanlder")
	}

	if user == nil {
		// session no longer belongs to user. Destroy session
		return h.Logout(w, r)
	}

	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		return fmt.Errorf("Could not get user by id in createcustomerportalsession handler")
	}

	stripeKey := os.Getenv(STRIPE_API_KEY)
	if stripeKey == "" {
		// do something
		return fmt.Errorf("no stripe key")
	}

	stripe.Key = stripeKey

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return fmt.Errorf("no domain key")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(user.StripeCustomerID),
		ReturnURL: stripe.String(domain),
	}

	s, err := billingportalsession.New(params)
	if err != nil {
		return fmt.Errorf("Failed to create a billing portal session")
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)

	infoMsg := fmt.Sprintf("User (%s) visited their billing portal", user.ID)
	h.logger.Info(infoMsg)
	return nil
}
