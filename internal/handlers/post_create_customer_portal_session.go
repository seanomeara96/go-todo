package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	billingportalsession "github.com/stripe/stripe-go/v75/billingportal/session"
)

func (h *Handler) CreateCustomerPortalSession(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		h.logger.Error("Could not get user from session on create customer portal session hanlder")
		h.logger.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		// session no longer belongs to user. Destroy session
		h.Logout(w, r)
		return
	}

	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		h.logger.Error("Could not get user by id in createcustomerportalsession handler")
		h.logger.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	stripeKey := os.Getenv(STRIPE_API_KEY)
	if stripeKey == "" {
		h.logger.Error("Cant access stripe api key from ENV")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// do something
		return
	}

	stripe.Key = stripeKey

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		h.logger.Error("Expected a domain in env vars")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// do something
		return
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(user.StripeCustomerID),
		ReturnURL: stripe.String(domain),
	}

	s, err := billingportalsession.New(params)
	if err != nil {
		h.logger.Error("Failed to create a billing portal session")
		http.Error(w, "There was a problem connecting with our payment provider. Please try again later", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)

	infoMsg := fmt.Sprintf("User (%s) visited their billing portal", user.ID)
	h.logger.Info(infoMsg)
}
