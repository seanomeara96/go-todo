package handlers

import (
	"fmt"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) UpgradePage(w http.ResponseWriter, r *http.Request) error {

	user, err := h.getUserFromSession(h.store.Get(r, USER_SESSION))
	if err != nil {
		return fmt.Errorf("could not get user from session in upgrade handler")
	}

	if user == nil {
		return h.Logout(w, r)
	}

	defer func() {
		infoMsg := fmt.Sprintf("User (%s) started upgrade flow", user.ID)
		h.logger.Info(infoMsg)
	}()

	basePageProps := renderer.NewBasePageProps(user)
	upgradePageProps := renderer.NewUpgradePageProps(basePageProps)
	upgradePageBytes, err := h.render.Upgrade(upgradePageProps)
	if err != nil {
		return err
	}

	_, err = w.Write(upgradePageBytes)
	if err != nil {
		return err
	}

	return nil
}
