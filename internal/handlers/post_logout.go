package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	session, err := h.store.Get(r, USER_SESSION)
	user, err := h.getUserFromSession(session, err)
	if err != nil {
		return err
	}

	err = h.store.Delete(r, w, session)
	if err != nil {
		h.logger.Error("Could not delete user session")
		h.logger.Debug(err.Error())
		return err
	}

	noCacheRedirect("/", w, r)

	infoMsg := fmt.Sprintf("Session deleted for user (%s)", user.ID)
	h.logger.Info(infoMsg)
	return nil
}
