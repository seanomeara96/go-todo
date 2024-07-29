package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/webhook"
)

func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		h.logger.Warning("Received a http method to webhook that wasn't POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil
	}

	stripeKey := os.Getenv("STRIPE_API_KEY")
	if stripeKey == "" {
		return fmt.Errorf("Can not get stripe key from env")
	}

	stripe.Key = stripeKey
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Bad request from stripe")
	}

	stripeWebhookSecret := os.Getenv(STRIPE_WEBHOOK_SECRET)
	if stripeWebhookSecret == "" {
		return fmt.Errorf("Can not find stripe webhook secret in env")
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		return fmt.Errorf("webhook.ConstructEvent: %v", err)
	}

	// TODO add writeheaders before the returns in this switch statement
	switch event.Type {
	case "checkout.session.completed":
		h.logger.Info("Handling checkout.session.completed event for completed checkouts.")

		var checkoutSession stripe.CheckoutSession

		err = json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			return fmt.Errorf("Error unmarshalling checkout.session.completed webhook")
		}

		user, err := h.service.GetUserByEmail(checkoutSession.Customer.Email)
		if err != nil {
			return nil
		}

		if user == nil {
			h.logger.Warning("Checkout session wasc ompleted but customer email does not match any users")
		}

		infoMsg := fmt.Sprintf("Customer (%s) completed checkout", checkoutSession.Customer.Email)
		h.logger.Info(infoMsg)

		fmt.Println("testing if customer email is included in the checkout session completed webhook", checkoutSession.Customer.Email)
		fmt.Println("testing checkout session completed webhook to se  if payment status == paid", checkoutSession.PaymentStatus)

		w.WriteHeader(http.StatusOK)
		return nil

	case "invoice.paid":
		log.Println("Handling invoice.paid event for payments.")
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			return fmt.Errorf("Failed to parse invoice paid webhook, %w", err)

		}

		customerID := invoice.Customer.ID
		customerEmail := invoice.Customer.Email

		user, err := h.service.GetUserByEmail(customerEmail)
		if err != nil {
			return err
		}

		if user == nil {
			return fmt.Errorf("Invoice paid webhook was called but the email address on the invoice did not return a user, %w", err)
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
		return nil

	case "invoice.payment_failed":
		h.logger.Info("Handling invoice.payment_failed event for failed payments.")

		var invoice stripe.Invoice

		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			return fmt.Errorf("Failed to unmarshal invoice.payment_failed webhook, %w", err)
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
			return err
		}

		if user == nil {
			return fmt.Errorf("invoice payment failed could not find matching user for that stripe ID (%s)", customerStripeID)
		}

		// deactivate user premium status
		err = h.service.UpdateUserPaymentStatus(user.ID, false)
		if err != nil {
			return err
		}

		infoMsg := fmt.Sprintf("downgraded plan for user: %s", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return nil

	case "customer.subscription.updated":
		// Check subscription.items.data[0].price attribute and grant/revoke access accordingly.
		h.logger.Info("Handling customer.subscription.updated event for price changes.")
		// cancel plan in customer billing portal triggers this event first
		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			return fmt.Errorf("Failed to parse customer.subscription.updated webhook, %w", err)
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			return err
		}

		// customer.ID in this instances looks like a subscription id
		if user == nil {
			return fmt.Errorf("Customer subscription updated but No user by that stripe ID (%s)", customer.ID)
		}

		infoMsg := fmt.Sprintf("User (%s) updated subscription", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return nil

	case "customer.subscription.deleted":
		// after the remaining days remining on the cancelled plan have elapsed, i assume this webhook gets called
		// Revoke customer's access to the product.
		h.logger.Info("Handling customer.subscription.deleted event for subscription cancellations.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			return fmt.Errorf("Could not unmarshal customer.subscription.deleted data, %w")
		}

		customerID := customer.ID
		if customerID == "" {
			return fmt.Errorf("cant get customer ID from webhook event data")
		}

		err = h.service.UpdateUserPaymentStatus(customerID, false)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		return nil

	case "customer.subscription.paused":
		// Revoke customer's access until subscription resumes.
		h.logger.Info("Handling customer.subscription.paused event for paused subscriptions.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			return fmt.Errorf("Error unmarshalling customer.subscription.paused webhook: %w", err)
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			return fmt.Errorf("Customer subscription paused by customer (%s) with no matching user", customer.ID)
		}

		if user == nil {
			return fmt.Errorf("no user exists with the stripe ID of %s", customer.ID)
		}

		infoMsg := fmt.Sprintf("User (%s) paused their subscription", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return nil

	case "customer.subscription.resumed":
		// Grant customer access when subscription resumes.
		h.logger.Info("Handling customer.subscription.resumed event for resumed subscriptions.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			return fmt.Errorf("Error unmarshalling customer.subscription.resumed webhook")
		}

		user, err := h.service.GetUserByStripeID(customer.ID)
		if err != nil {
			return fmt.Errorf("Error getting customer by stripe ID (%s)", customer.ID)
		}

		if user == nil {
			return fmt.Errorf("no user exists with the ID of %s", customer.ID)

		}

		infoMsg := fmt.Sprintf("User (%s) resumed their subscription", user.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return nil

	case "payment_method.attached":
		// Handle payment method attachment.
		h.logger.Info("Handling payment_method.attached event for payment method attachment.")

		var paymentMethod stripe.PaymentMethod

		err = json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			return fmt.Errorf("Error unmarshalling payment_method.attached webhook")
		}

		log.Printf("Customer (%s) updated their payment method", paymentMethod.Customer.Email)
		w.WriteHeader(http.StatusOK)
		return nil

	case "payment_method.detached":
		// Handle payment method detachment.
		h.logger.Info("Handling payment_method.detached event for payment method detachment.")

		var paymentMethod stripe.PaymentMethod

		err = json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			return fmt.Errorf("Error unmarshalling payment_method.detached webhook")
		}

		infoMsg := fmt.Sprintf("Customer (%s) detached their payment method", paymentMethod.Customer.Email)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return nil

	case "customer.updated":
		// Check and update default payment method information.
		h.logger.Info("Handling customer.updated event for default payment method updates.")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			return fmt.Errorf("Error unmarshalling customer.updated webhook")
		}

		infoMsg := fmt.Sprintf("Customer (%s) updated", customer.ID)
		h.logger.Info(infoMsg)
		w.WriteHeader(http.StatusOK)
		return nil

	case "customer.tax_id.created", "customer.tax_id.deleted", "customer.tax_id.updated":
		// Handle tax ID related events.
		h.logger.Info("Handling tax ID related event")

		var customer stripe.Customer

		err = json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			return fmt.Errorf("Error unmarshalling customer.tax_id.created webhook")
		}

		log.Printf("Customer (%s) added tax ID", customer.Email)
		w.WriteHeader(http.StatusOK)
		return nil

	case "billing_portal.configuration.created", "billing_portal.configuration.updated":
		// Handle billing portal configuration events.
		h.logger.Info("Handling billing portal configuration event")
		w.WriteHeader(http.StatusOK)
		return nil

	case "billing_portal.session.created":
		// Handle billing portal session creation.
		h.logger.Info("Handling billing portal session created event.")
		w.WriteHeader(http.StatusOK)
		return nil

	default:
		warningMsg := fmt.Sprintf("Unknown event type sent to webhook endpoint: %s", event.Type)
		h.logger.Warning(warningMsg)
		w.WriteHeader(http.StatusBadRequest)
		return nil
		// something else happened
	}

	/*
		The minimum event types to monitor:
		Event name	Description
		checkout.session.completed	Sent when a customer clicks the Pay or Subscribe button in Checkout, informing you of a new purchase.
		invoice.paid	Sent each billing interval when a payment succeeds.
		invoice.payment_failed	Sent each billing interval if there is an issue with your customer’s payment method.*/
}
