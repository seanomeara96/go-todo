package handlers

import (
	"errors"
	"fmt"
	"go-todo/internal/models"
	"go-todo/internal/server/renderer"
	"net/http"
)

func (h *Handler) UpgradePage(w http.ResponseWriter, r *http.Request) error {

	user, ok := r.Context().Value(userIDKey).(*models.User)
	if !ok {
		return errors.New("user not found")
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
