package router

import (
	"go-todo/internal/handlers"
	"net/http"
)

func NewRouter(handler *handlers.Handler) *http.ServeMux {
	r := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	r.Handle("/assets/", http.StripPrefix("/assets/", fs))

	middleware := []handlers.MiddleWareFunc{
		handler.UserFromSession,
		handler.PathLogger,
	}

	handle := newHandleFunc(r, middleware)

	method := func(method string) func(path string, fn handlers.HandleFunc) {
		return func(path string, fn handlers.HandleFunc) {
			handle(method+" "+path, fn)
		}
	}

	get := method(http.MethodGet)
	post := method(http.MethodPost)

	get("/signup", handler.SignupPage)
	post("/signup", handler.CreateUser)
	get("/success", handler.SuccessPage)
	get("/cancel", handler.CancelPage)
	get("/upgrade", handler.MustBeLoggedIn(handler.UpgradePage))

	post("/login", handler.Login)
	get("/logout", handler.Logout)

	post("/todo/add", handler.AddTodo)
	post("/todo/update/description", handler.UpdateTodoDescription)
	post("/todo/update/status/{id}", handler.UpdateTodoStatus)
	post("/todo/remove/{id}", handler.RemoveTodo)

	post("/create-checkout-session", handler.CreateCheckoutSession)
	get("/manage-subscription", handler.CreateCustomerPortalSession)
	post("/webhook", handler.HandleStripeWebhook)

	handle("/", handler.HomePage)
	return r
}

func newHandleFunc(r *http.ServeMux, use []handlers.MiddleWareFunc) func(path string, fn handlers.HandleFunc) {
	return func(path string, fn handlers.HandleFunc) {

		fn = fn.Use(use...)

		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if err := fn(w, r); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}
}
