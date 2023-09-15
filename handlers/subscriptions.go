func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "could not retrieve user session", http.StatusInternalServerError)
		return
	}

	user := GetUserFromSession(userSession)
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
	// handle error?

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
