package router

import (
	"go-todo/internal/handlers"
	"net/http"
)

func registerUserRoutes(handler *handlers.Handler, r *http.ServeMux) {
	middleware := []handlers.MiddleWareFunc{
		handler.UserFromSession,
		handler.PathLogger,
	}
	handle := newHandleFunc(r, middleware)
	method := newMethodFunc("", handle)
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
}

func registerAdminRoutes(handler *handlers.Handler, r *http.ServeMux) {
	adminMiddleware := []handlers.MiddleWareFunc{
		handler.PathLogger,
	}
	handle := newHandleFunc(r, adminMiddleware)
	method := newMethodFunc("/admin", handle)
	get := method(http.MethodGet)
	//adminPost := adminMethod(http.MethodPost)

	get("/hello", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("hello!"))
		return nil
	})

	get("/manage", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("manage!"))
		return nil
	})
}

func newHandleFunc(r *http.ServeMux, middleware []handlers.MiddleWareFunc) func(path string, fn handlers.HandleFunc) {
	return func(path string, fn handlers.HandleFunc) {

		for i := range middleware {
			fn = middleware[i](fn)
		}

		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if err := fn(w, r); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}
}

func newMethodFunc(prefix string, handle func(path string, fn handlers.HandleFunc)) func(method string) func(path string, fn handlers.HandleFunc) {
	return func(method string) func(path string, fn handlers.HandleFunc) {
		return func(path string, fn handlers.HandleFunc) {
			handle(method+" "+prefix+path, fn)
		}
	}
}

func NewRouter(handler *handlers.Handler) *http.ServeMux {
	r := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	r.Handle("/assets/", http.StripPrefix("/assets/", fs))

	registerUserRoutes(handler, r)

	registerAdminRoutes(handler, r)

	return r
}
