package handlers

import (
	"encoding/json"
	"fmt"
	"go-todo/internal/models"
	"go-todo/internal/server/logger"
	"go-todo/internal/server/renderer"
	"go-todo/internal/services"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/stripe/stripe-go/v75"
	billingportalsession "github.com/stripe/stripe-go/v75/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/webhook"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

const STRIPE_API_KEY = "STRIPE_API_KEY"
const USER_SESSION = "user-session"
const STRIPE_WEBHOOK_SECRET = "STRIPE_WEBHOOK_SECRET"

type Handler struct {
	service *services.Service
	store   *sqlitestore.SqliteStore
	render  *renderer.Renderer
	logger  *logger.Logger
}

func NewHandler(service *services.Service, store *sqlitestore.SqliteStore, renderer *renderer.Renderer, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		store:   store,
		render:  renderer,
		logger:  logger,
	}
}

func (h *Handler) getUserFromSession(s *sessions.Session, err error) (*models.User, error) {
	if err != nil {
		return nil, err
	}
	val := s.Values["user"]
	userID, ok := val.(string)
	if !ok {
		return nil, nil
	}
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func noCacheRedirect(path string, w http.ResponseWriter, r *http.Request) {
	// Set cache-control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect the user to a new URL
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// POST /login
/*
	Redirects to homepage when login is success. If there are client
	errors then the homepage will be rerendered with error messages
*/
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, USER_SESSION)
	user, err := h.getUserFromSession(session, err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		// user already logged in
		noCacheRedirect("/", w, r)
		return
	}

	/*
		no need to call logout if user == nil
		as we'll just render the login form instead.
	*/

	err = r.ParseForm()
	if err != nil {
		h.logger.Error("Could not parse login form")
		h.logger.Debug(err.Error())
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	email, password := r.FormValue("email"), r.FormValue("password")

	user, userErrors, err := h.service.Login(email, password)
	if err != nil {
		h.logger.Error("Could not log in user")
		h.logger.Debug(err.Error())
		http.Error(w, "error logging user in", http.StatusInternalServerError)
		return
	}

	if userErrors != nil {
		loginFormProps := renderer.NewLoginFormProps(userErrors.EmailErrors, userErrors.PasswordErrors)
		basePageProps := renderer.NewBasePageProps(nil)
		todoListProps := renderer.NewTodoListProps([]*models.Todo{}, false)
		homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps, loginFormProps)
		bytes, err := h.render.HomePage(homePageProps)
		if err != nil {
			http.Error(w, "could not render homepage", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bytes)

		if err != nil {
			h.logger.Error("Could not write homepage")
			h.logger.Debug(err.Error())
			http.Error(w, "could not write homepage", http.StatusInternalServerError)
		}

		return
	}

	session.Values["user"] = user.ID
	session.Save(r, w)
	noCacheRedirect("/", w, r)

	infoMsg := fmt.Sprintf("Session created for user (%s) logged in", user.ID)
	h.logger.Info(infoMsg)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, USER_SESSION)
	user, err := h.getUserFromSession(session, err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.store.Delete(r, w, session)
	if err != nil {
		h.logger.Error("Could not delete user session")
		h.logger.Debug(err.Error())
		http.Error(w, "could not delete session", http.StatusInternalServerError)
		return
	}

	noCacheRedirect("/", w, r)

	infoMsg := fmt.Sprintf("Session deleted for user (%s)", user.ID)
	h.logger.Info(infoMsg)
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	canCreateNewTodo := false
	var list []*models.Todo

	if user != nil {
		/*
			if user is logged in we need to get their todos
			and whether they have permission to create a new todo
		*/
		list, err = h.service.GetUserTodoList(user.ID)
		if err != nil {
			h.logger.Error("Could not get user list of todos")
			h.logger.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		canCreateNewTodo, err = h.service.UserCanCreateNewTodo(user, list)
		if err != nil {
			h.logger.Error("Cannot determine whether user can create new todo")
			h.logger.Debug(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	basePageProps := renderer.NewBasePageProps(user)
	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	noErrors := []string{}
	loginFormProps := renderer.NewLoginFormProps(noErrors, noErrors)
	homePageProps := renderer.NewHomePageProps(basePageProps, todoListProps, loginFormProps)

	bytes, err := h.render.HomePage(homePageProps)
	if err != nil {
		h.logger.Error("could not render home-logged-out")
		h.logger.Debug(err.Error())
		http.Error(w, "could not render home-logged-out", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		h.logger.Error("could not write homepage")
		h.logger.Debug(err.Error())
		http.Error(w, "could not write home-logged-out", http.StatusInternalServerError)
		return
	}

	if user != nil {
		infoMsg := fmt.Sprintf("User (%s) loaded their todo list", user.ID)
		h.logger.Info(infoMsg)
	}
}

func (h *Handler) SignupPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		noCacheRedirect("/", w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	noErrors := []string{}
	signupFormProps := renderer.NewSignupFormProps(noErrors, noErrors, noErrors)
	signupPageProps := renderer.NewSignupPageProps(basePageProps, signupFormProps)
	signupPageBytes, err := h.render.Signup(signupPageProps)
	if err != nil {
		h.logger.Error("could not render signup page")
		h.logger.Debug(err.Error())
		http.Error(w, "could not render sigup page", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(signupPageBytes)
	if err != nil {
		h.logger.Error("Could not write signup page")
		h.logger.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpgradePage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		h.logger.Error("Could not get user from session in upgrade handler")
		h.logger.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	upgradePageProps := renderer.NewUpgradePageProps(basePageProps)
	upgradePageBytes, err := h.render.Upgrade(upgradePageProps)
	if err != nil {
		h.logger.Error("Could not render upgrade page")
		h.logger.Debug(err.Error())
		http.Error(w, "could not render upgrade page", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(upgradePageBytes)
	if err != nil {
		h.logger.Error("could not write upgrade page")
		h.logger.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	infoMsg := fmt.Sprintf("User (%s) started upgrade flow", user.ID)
	h.logger.Info(infoMsg)
}

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

func (h *Handler) CancelPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// redirect to homepage if not logged in
	if user == nil {
		h.Logout(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	cancelPageProps := renderer.NewCancelPageProps(basePageProps)
	bytes, err := h.render.Cancel(cancelPageProps)
	if err != nil {
		h.logger.Error("Cannot render cancel page")
		h.logger.Debug(err.Error())
		http.Error(w, "Cant render the cancel page at the moment. Please reach out to support so we can resolve this issue.", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		h.logger.Error("Could not write cancel page")
		h.logger.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	infoMsg := fmt.Sprintf("User (%s) has started cancellation flow", user.ID)
	h.logger.Info(infoMsg)
}

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

func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Warning("Received a http method to webhook that wasn't POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		h.logger.Warning("Can not get stripe key from env")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey
	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Bad request from stripe")
		h.logger.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	stripeWebhookSecret := os.Getenv(STRIPE_WEBHOOK_SECRET)
	if stripeWebhookSecret == "" {
		h.logger.Error("Can not find stripe webhook secret in env")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		errMsg := fmt.Sprintf("webhook.ConstructEvent: %v", err)
		h.logger.Error(errMsg)
		return
	}

	// TODO add writeheaders before the returns in this switch statement
	switch event.Type {
	case "checkout.session.completed":
		h.logger.Info("Handling checkout.session.completed event for completed checkouts.")

		var checkoutSession stripe.CheckoutSession

		err = json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			h.logger.Error("Error unmarshalling checkout.session.completed webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByEmail(checkoutSession.Customer.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			h.logger.Warning("Checkout session wasc ompleted but customer email does not match any users")
		}

		infoMsg := fmt.Sprintf("Customer (%s) completed checkout", checkoutSession.Customer.Email)
		h.logger.Info(infoMsg)

		fmt.Println("testing if customer email is included in the checkout session completed webhook", checkoutSession.Customer.Email)
		fmt.Println("testing checkout session completed webhook to se  if payment status == paid", checkoutSession.PaymentStatus)

		w.WriteHeader(http.StatusOK)
		return

	case "invoice.paid":
		log.Println("Handling invoice.paid event for payments.")
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			h.logger.Error("Failed to parse invoice paid webhook")
			h.logger.Debug(err.Error())
			return
		}

		customerID := invoice.Customer.ID
		customerEmail := invoice.Customer.Email

		user, err := h.service.GetUserByEmail(customerEmail)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			h.logger.Warning("Invoice paid webhook was called but the email address on the invoice did not return a user")
			debugMsg := fmt.Sprintf("This customer email did not return a user: %s", invoice.Customer.Email)
			h.logger.Debug(debugMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.StripeCustomerID == "" {
			err = h.service.AddStripeIDToUser(user.ID, customerID)
			if err != nil {
				// URGENT TODO will need a notification for this
				warningMsg := fmt.Sprintf("customer with ID of %s paid invoice but customer ID could not be addedto user (%s)", customerID, user.ID)
				h.logger.Warning(warningMsg)
			}
		}
		// TODO find a way to handle this better

		err = h.service.UpdateUserPaymentStatus(user.ID, true)
		if err != nil {
			warningMsg := fmt.Sprintf("customer %s paid their invoice but could not update the user (%s) record", customerID, user.ID)
			h.logger.Warning(warningMsg)
		}

		w.WriteHeader(http.StatusOK)
		return

	case "invoice.payment_failed":
		h.logger.Info("Handling invoice.payment_failed event for failed payments.")

		var invoice stripe.Invoice

		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			h.logger.Error("Failed to unmarshal invoice.payment_failed webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		customerStripeID := invoice.Customer.ID
		if customerStripeID == "" {
			h.logger.Error("cant get customer id from invoice.payment_failed webhook data")
			h.logger.Debug(err.Error())
		}

		if customerStripeID != "" {
			infoMsg := fmt.Sprintf("invoice payment failed for customer ID: %s \n", customerStripeID)
			h.logger.Info(infoMsg)
		}

		// TODO maybe add get user by email fallback
		user, err := h.service.GetUserByStripeID(customerStripeID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			warningMsg := fmt.Sprintf("invoice payment failed could not find matching user for that stripe ID (%s)", customerStripeID)
			h.logger.Warning(warningMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// deactivate user premium status
		err = h.service.UpdateUserPaymentStatus(user.ID, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infoMsg := fmt.Sprintf("downgraded plan for user: %s", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.updated":
		// Check subscription.items.data[0].price attribute and grant/revoke access accordingly.
		h.logger.Info("Handling customer.subscription.updated event for price changes.")
		// cancel plan in customer billing portal triggers this event first
		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			h.logger.Error("Failed to parse customer.subscription.updated webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// customer.ID in this instances looks like a subscription id
		if user == nil {
			warningMsg := fmt.Sprintf("Customer subscription updated but No user by that stripe ID (%s)", customer.ID)
			h.logger.Warning(warningMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infoMsg := fmt.Sprintf("User (%s) updated subscription", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.deleted":
		// after the remaining days remining on the cancelled plan have elapsed, i assume this webhook gets called
		// Revoke customer's access to the product.
		h.logger.Info("Handling customer.subscription.deleted event for subscription cancellations.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			h.logger.Error("Could not unmarshal customer.subscription.deleted data")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		customerID := customer.ID
		if customerID == "" {
			h.logger.Error("cant get customer ID from webhook event data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = h.service.UpdateUserPaymentStatus(customerID, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.paused":
		// Revoke customer's access until subscription resumes.
		h.logger.Info("Handling customer.subscription.paused event for paused subscriptions.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			errMsg := fmt.Sprintf("Error unmarshalling customer.subscription.paused webhook: %s", err.Error())
			h.logger.Error(errMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			errMsg := fmt.Sprintf("Customer subscription paused by customer (%s) with no matching user", customer.ID)
			h.logger.Error(errMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			warningmsg := fmt.Sprintf("no user exists with the stripe ID of %s", customer.ID)
			h.logger.Warning(warningmsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infoMsg := fmt.Sprintf("User (%s) paused their subscription", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.resumed":
		// Grant customer access when subscription resumes.
		h.logger.Info("Handling customer.subscription.resumed event for resumed subscriptions.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			h.logger.Error("Error unmarshalling customer.subscription.resumed webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			errMsg := fmt.Sprintf("Error getting customer by stripe ID (%s)", customer.ID)
			h.logger.Error(errMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user == nil {
			errMsg := fmt.Sprintf("no user exists with the ID of %s", customer.ID)
			h.logger.Error(errMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infoMsg := fmt.Sprintf("User (%s) resumed their subscription", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return

	case "payment_method.attached":
		// Handle payment method attachment.
		h.logger.Info("Handling payment_method.attached event for payment method attachment.")

		var paymentMethod stripe.PaymentMethod

		err = json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			h.logger.Error("Error unmarshalling payment_method.attached webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) updated their payment method", paymentMethod.Customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "payment_method.detached":
		// Handle payment method detachment.
		h.logger.Info("Handling payment_method.detached event for payment method detachment.")

		var paymentMethod stripe.PaymentMethod

		err = json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			h.logger.Error("Error unmarshalling payment_method.detached webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infoMsg := fmt.Sprintf("Customer (%s) detached their payment method", paymentMethod.Customer.Email)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.updated":
		// Check and update default payment method information.
		h.logger.Info("Handling customer.updated event for default payment method updates.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			h.logger.Error("Error unmarshalling customer.updated webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infoMsg := fmt.Sprintf("Customer (%s) updated", customer.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.tax_id.created", "customer.tax_id.deleted", "customer.tax_id.updated":
		// Handle tax ID related events.
		h.logger.Info("Handling tax ID related event")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			h.logger.Error("Error unmarshalling customer.tax_id.created webhook")
			h.logger.Debug(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Customer (%s) added tax ID", customer.Email)
		w.WriteHeader(http.StatusOK)
		return

	case "billing_portal.configuration.created", "billing_portal.configuration.updated":
		// Handle billing portal configuration events.
		h.logger.Info("Handling billing portal configuration event")
		w.WriteHeader(http.StatusOK)
		return

	case "billing_portal.session.created":
		// Handle billing portal session creation.
		h.logger.Info("Handling billing portal session created event.")
		w.WriteHeader(http.StatusOK)
		return

	default:
		warningMsg := fmt.Sprintf("Unknown event type sent to webhook endpoint: %s", event.Type)
		h.logger.Warning(warningMsg)
		w.WriteHeader(http.StatusBadRequest)
		return
		// something else happened
	}

	/*
		The minimum event types to monitor:
		Event name	Description
		checkout.session.completed	Sent when a customer clicks the Pay or Subscribe button in Checkout, informing you of a new purchase.
		invoice.paid	Sent each billing interval when a payment succeeds.
		invoice.payment_failed	Sent each billing interval if there is an issue with your customer’s payment method.*/
}

func (h *Handler) AddTodo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusInternalServerError)
		return
	}

	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	// TODO render client errors
	lastInsertedTodo, _, err := h.service.CreateTodo(user.ID, r.FormValue("description"))
	// TODO for now i am returning an error if todo is nil
	if err != nil || lastInsertedTodo == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := h.service.GetUserTodoList(user.ID)
	if err != nil {
		http.Error(w, "something went wrong while fetching todos", http.StatusInternalServerError)
		return
	}

	canCreateNewTodo, err := h.service.UserCanCreateNewTodo(user, list)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	todoListProps := renderer.NewTodoListProps(list, canCreateNewTodo)
	todoList, err := h.render.TodoList(todoListProps)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

	w.Write(todoList)

	infoMsg := fmt.Sprintf("User (%s) added a todo (%d)", user.ID, lastInsertedTodo.ID)
	h.logger.Info(infoMsg)
}
func (h *Handler) UpdateTodoDescription(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) RemoveTodo(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	vars := mux.Vars(r)
	todoIDString := vars["id"]

	todoID, err := strconv.Atoi(todoIDString)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, clientError, internalError := h.service.GetTodoByID(todoID, user.ID)
	if internalError != nil {
		http.Error(w, "could not get todo", http.StatusInternalServerError)
		return
	}

	if clientError != nil {
		http.Error(w, clientError.Message, clientError.Code)
		return
	}

	clientError, internalError = h.service.DeleteTodo(todo.ID, user.ID)
	if internalError != nil {
		http.Error(w, "could not remove todo", http.StatusInternalServerError)
		return
	}

	if clientError != nil {
		http.Error(w, clientError.Message, clientError.Code)
		return
	}

	w.Write([]byte(""))

	infoMsg := fmt.Sprintf("User (%s) removed todo (%s)", user.ID, todoIDString)
	h.logger.Info(infoMsg)
}

func (h *Handler) UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	user, err = h.service.GetUserByID(user.ID)
	if err != nil {
		http.Error(w, "trouble finding that user", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	idParam := vars["id"]
	todoID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "path does not contain valid id", http.StatusBadRequest)
		return
	}

	todo, clientError, err := h.service.UpdateTodoStatus(user.ID, todoID)
	if err != nil {
		http.Error(w, "could not update todo", http.StatusInternalServerError)
		return
	}

	if clientError != nil {
		http.Error(w, clientError.Message, clientError.Code)
		return
	}

	if todo == nil {
		http.Error(w, "service did not return a todo item", http.StatusInternalServerError)
		return
	}

	todoBytes, err := h.render.Todo(todo)
	if err != nil {
		http.Error(w, "could not render todo", http.StatusInternalServerError)
		return
	}

	w.Write(todoBytes)

	infoMsg := fmt.Sprintf("User (%s) updated todo (%d) to status: %v", user.ID, todo.ID, todo.IsComplete)
	h.logger.Info(infoMsg)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse signup form", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// TODO sanitize and clean input

	newUser, userErrors, err := h.service.NewUser(name, email, password)
	if userErrors != nil {
		basePageProps := renderer.NewBasePageProps(nil)
		signUpFormProps := renderer.NewSignupFormProps(userErrors.UsernameErrors, userErrors.EmailErrors, userErrors.PasswordErrors)
		signupPageProps := renderer.NewSignupPageProps(basePageProps, signUpFormProps)
		signupPageBytes, err := h.render.Signup(signupPageProps)
		if err != nil {
			http.Error(w, "could not render sigup page", http.StatusInternalServerError)
			return
		}
		w.Write(signupPageBytes)
		return
	}

	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

	infoMsg := fmt.Sprintf("New user (%s) created", newUser.ID)
	h.logger.Info(infoMsg)
}

// func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request) {

// 	// stripe code

// }
