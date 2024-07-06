package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) UpgradePage(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		h.logger.Error("Could not get user from session in upgrade handler")
		h.logger.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.Logout(w, r)
		return
	}

	basePageProps := renderer.NewBasePageProps(user)
	upgradePageProps := renderer.NewUpgradePageProps(basePageProps)
	upgradePageBytes, err := h.render.Upgrade(upgradePageProps)
	if err != nil {
		h.logger.Error("Could not render upgrade page")
		h.logger.Debug(err.Error())
		http.Error(w, "could not render upgrade page", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(upgradePageBytes)
	if err != nil {
		h.logger.Error("could not write upgrade page")
		h.logger.Debug(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	infoMsg := fmt.Sprintf("User (%s) started upgrade flow", user.ID)
	h.logger.Info(infoMsg)
}
