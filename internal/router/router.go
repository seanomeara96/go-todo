package router

import (
	"go-todo/internal/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler *handlers.Handler) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	fs := http.FileServer(http.Dir("assets"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))
	r.HandleFunc("/", handler.HomePage).Methods(http.MethodGet)
	r.HandleFunc("/signup", handler.SignupPage).Methods(http.MethodGet)
	r.HandleFunc("/signup", handler.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/success", handler.SuccessPage).Methods(http.MethodGet)
	r.HandleFunc("/cancel", handler.CancelPage).Methods(http.MethodGet)
	r.HandleFunc("/upgrade", handler.UpgradePage).Methods(http.MethodGet)
	r.HandleFunc("/login", handler.Login).Methods(http.MethodPost)
	r.HandleFunc("/logout", handler.Logout).Methods(http.MethodGet)
	r.HandleFunc("/todo/add", handler.AddTodo).Methods(http.MethodPost)
	r.HandleFunc("/todo/update/description", handler.UpdateTodoDescription).Methods(http.MethodPost)
	r.HandleFunc("/todo/update/status/{id}", handler.UpdateTodoStatus).Methods(http.MethodPost)
	r.HandleFunc("/todo/remove/{id}", handler.RemoveTodo).Methods(http.MethodPost)
	r.HandleFunc("/create-checkout-session", handler.CreateCheckoutSession).Methods(http.MethodPost)
	r.HandleFunc("/manage-subscription", handler.CreateCustomerPortalSession).Methods(http.MethodGet)
	r.HandleFunc("/webhook", handler.HandleStripeWebhook).Methods(http.MethodPost)
	return r
}
