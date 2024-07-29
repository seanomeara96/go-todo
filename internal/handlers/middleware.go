package handlers

import "net/http"

func (h *Handler) MustBeLoggedIn(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// do something

		return next(w, r)
	}
}
