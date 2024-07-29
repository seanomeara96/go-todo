package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

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
