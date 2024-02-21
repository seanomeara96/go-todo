package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
	"os"

	checkoutsession "github.com/stripe/stripe-go/v75/checkout/session"

	"github.com/stripe/stripe-go/v75"
)

func (h *Handler) SuccessPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	checkoutSessionID := r.URL.Query().Get("session_id")

	stripeKey := os.Getenv(STRIPE_API_KEY)

	if stripeKey == "" {
		errorMsg := fmt.Sprintf("Could not access key(%s) from ENV", STRIPE_API_KEY)
		h.logger.Error(errorMsg)
		http.Error(w, "Trouble connecting with payment provider. Please try again later.", http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey

	s, err := checkoutsession.Get(checkoutSessionID, nil)
	if err != nil {
		errMsg := "Could get checkout session from stripe"
		h.logger.Error(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
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

		infoMsg := fmt.Sprintf("User (%s) became a premium user", user.ID)
		h.logger.Info(infoMsg)
	}

	basePageProps := renderer.NewBasePageProps(user)
	successPageProps := renderer.NewSuccessPageProps(basePageProps)
	bytes, err := h.render.Success(successPageProps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		h.logger.Error("Could not write success page")
		h.logger.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
