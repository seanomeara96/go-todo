package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) CancelPage(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return err
	}

	// redirect to homepage if not logged in
	if user == nil {
		return h.Logout(w, r)

	}

	basePageProps := renderer.NewBasePageProps(user)
	cancelPageProps := renderer.NewCancelPageProps(basePageProps)
	bytes, err := h.render.Cancel(cancelPageProps)
	if err != nil {
		return fmt.Errorf("cannot render cancel page, %w", err)
	}

	_, err = w.Write(bytes)
	if err != nil {
		return fmt.Errorf("could not write cancel page, %w", err)
	}

	infoMsg := fmt.Sprintf("User (%s) has started cancellation flow", user.ID)
	h.logger.Info(infoMsg)
	return nil
}
