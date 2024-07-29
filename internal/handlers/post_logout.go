package handlers

import (
	"fmt"
	"net/http"
)

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
