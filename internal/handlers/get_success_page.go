package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
	"os"

	checkoutsession "github.com/stripe/stripe-go/v75/checkout/session"

	"github.com/stripe/stripe-go/v75"
)

func (h *Handler) SuccessPage(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return err
	}

	if user == nil {
		return h.Logout(w, r)
	}

	checkoutSessionID := r.URL.Query().Get("session_id")

	stripeKey := os.Getenv(STRIPE_API_KEY)

	if stripeKey == "" {
		return fmt.Errorf("could not access key(%s) from ENV", STRIPE_API_KEY)
	}

	stripe.Key = stripeKey

	s, err := checkoutsession.Get(checkoutSessionID, nil)
	if err != nil {
		errMsg := "could get checkout session from stripe"
		return fmt.Errorf(errMsg)
	}
	// handle error ?

	// get user from db
	user, err = h.service.GetUserByEmail(s.CustomerEmail)
	if err != nil {
		return err
	}

	err = h.service.AddStripeIDToUser(user.ID, s.Customer.ID)
	if err != nil {
		return err
	}

	if s.PaymentStatus == "paid" {
		err = h.service.UpdateUserPaymentStatus(user.ID, true)
		if err != nil {
			return err
		}

		infoMsg := fmt.Sprintf("User (%s) became a premium user", user.ID)
		h.logger.Info(infoMsg)
	}

	basePageProps := renderer.NewBasePageProps(user)
	successPageProps := renderer.NewSuccessPageProps(basePageProps)
	bytes, err := h.render.Success(successPageProps)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return fmt.Errorf("Could not write success page")
	}
	return nil
}
