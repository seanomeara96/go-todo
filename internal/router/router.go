package router

import (
	"go-todo/internal/handlers"
	"net/http"
)

type router struct {
	Prefix     string
	Mux        *http.ServeMux
	middleware []handlers.MiddleWareFunc
}

func NewRouter(handler *handlers.Handler) *http.ServeMux {
	r := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	r.Handle("/assets/", http.StripPrefix("/assets/", fs))

	app := newRouter(r)

	app.Use(handler.AddUserToContext)
	app.Use(handler.PathLogger)

	app.Handle("/", handler.HomePage)

	app.Get("/signup", handler.SignupPage)
	app.Post("/signup", handler.CreateUser)
	app.Get("/success", handler.SuccessPage)
	app.Get("/subscription/cancel", handler.CancelPage)
	app.Get("/subscription/upgrade", handler.UserMustBeLoggedIn(handler.UpgradePage))

	app.Post("/login", handler.Login)
	app.Get("/logout", handler.Logout)

	app.Post("/todo/add", handler.AddTodo)
	app.Post("/todo/update/description", handler.UpdateTodoDescription)
	app.Post("/todo/update/status/{id}", handler.UpdateTodoStatus)
	app.Post("/todo/remove/{id}", handler.RemoveTodo)

	app.Post("/create-checkout-session", handler.CreateCheckoutSession)
	app.Get("/manage-subscription", handler.CreateCustomerPortalSession)
	app.Post("/webhook", handler.HandleStripeWebhook)

	admin := app.SubRouter("/admin", false)

	admin.Use(handler.UserMustBeAdmin)
	admin.Use(handler.AddUserToContext)
	admin.Use(handler.UserMustBeLoggedIn)
	admin.Use(handler.PathLogger)

	admin.Get("/dashboard", handler.AdminDashboard)
	admin.Get("/analytics", handler.AnalyticsDashboard)

	users := admin.SubRouter("/users", true)

	users.Get("", handler.UsersPage)
	users.Get("/{user_id}", handler.UserProfilePage)
	users.Put("/{user_id}", handler.UpdateUser)
	users.Delete("/{user_id}", handler.DeleteUser)

	return r
}

func newRouter(mux *http.ServeMux) *router {
	return &router{"", mux, []handlers.MiddleWareFunc{}}
}

func (parent *router) SubRouter(prefix string, carryMiddleware bool) *router {
	var middleware []handlers.MiddleWareFunc
	if carryMiddleware {
		middleware = parent.Middleware()
	}
	return &router{prefix, parent.Mux, middleware}
}

func (s *router) Handle(path string, fn handlers.HandleFunc) {

	// wrap fn in middleware
	for i := range s.middleware {
		fn = s.middleware[i](fn)
	}

	s.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

}
func (s *router) Get(path string, fn handlers.HandleFunc) {
	path = s.Prefix + path
	s.Handle(http.MethodGet+" "+path, fn)
}
func (s *router) Post(path string, fn handlers.HandleFunc) {
	path = s.Prefix + path
	s.Handle(http.MethodPost+" "+path, fn)
}
func (s *router) Put(path string, fn handlers.HandleFunc) {
	path = s.Prefix + path
	s.Handle(http.MethodPut+" "+path, fn)
}
func (s *router) Delete(path string, fn handlers.HandleFunc) {
	path = s.Prefix + path
	s.Handle(http.MethodDelete+" "+path, fn)
}

func (s *router) Use(fn handlers.MiddleWareFunc) {
	s.middleware = append(s.middleware, fn)
}

func (s *router) Middleware() []handlers.MiddleWareFunc {
	return s.middleware
}
