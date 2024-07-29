package router

import (
	"fmt"
	"go-todo/internal/handlers"
	"net/http"
)

func adminSubRouter() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("## subrouter called %s ###\n", r.URL.Path)
		w.Write([]byte(`hello admin`))
	})

	return r
}

func NewRouter(handler *handlers.Handler) *http.ServeMux {
	r := http.NewServeMux()

	r.Handle("/admin/", adminSubRouter())

	fs := http.FileServer(http.Dir("assets"))
	r.Handle("/assets/", http.StripPrefix("/assets/", fs))

	middleware := []handlers.MiddleWareFunc{}

	handle := func(path string, fn handlers.HandleFunc) {

		fn = fn.Use(middleware...)

		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if err := fn(w, r); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}

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
	get("/upgrade", handler.UpgradePage)
	post("/login", handler.Login)

	get("/logout", handler.Logout.Use(handler.MustBeLoggedIn))

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
