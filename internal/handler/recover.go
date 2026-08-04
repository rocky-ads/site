package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/accountrecovery"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/ui"
)

func setNoStore(c *fiber.Ctx) {
	c.Set("Cache-Control", "no-store")
}

func RecoverHandler(c *fiber.Ctx) error {
	setNoStore(c)

	if config.TwilioFromNumber == "" && !config.AllowTestRegistration {
		return renderPage(c, "Recover account", ui.RecoverUnavailablePage())
	}

	prev := cookie.GetRecoverSession(c)
	if prev != "" {
		_ = accountrecovery.Cancel(prev)
		cookie.ClearRecoverSession(c)
	}

	session, err := accountrecovery.Start()
	if err != nil {
		logger.Error("Failed to start recovery session", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError,
			"Unable to start account recovery. Please try again.")
	}

	cookie.SetRecoverSession(c, session.Token)
	return renderPage(c, "Recover account",
		ui.RecoverPage(session.Code, config.TwilioFromNumber, session.ExpiresAt))
}

func RecoverStatusHandler(c *fiber.Ctx) error {
	setNoStore(c)

	token := cookie.GetRecoverSession(c)
	status, err := accountrecovery.GetStatus(token)
	if err != nil {
		logger.Error("Failed to get recovery status", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError,
			"Unable to check recovery status")
	}

	switch status.Kind {
	case accountrecovery.StatusVerified:
		return render(c, ui.RecoverResetForm(status.Username))
	case accountrecovery.StatusFailed:
		return render(c, ui.RecoverFailedPanel(status.Message))
	case accountrecovery.StatusExpired:
		cookie.ClearRecoverSession(c)
		return render(c, ui.RecoverExpiredPanel())
	default:
		// Leave waiting panel (code + countdown) in place.
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func RecoverPasswordHandler(c *fiber.Ctx) error {
	setNoStore(c)

	token := cookie.GetRecoverSession(c)
	newPassword := c.FormValue("password")
	confirmPassword := c.FormValue("password2")

	err := accountrecovery.ResetPassword(token, newPassword, confirmPassword)
	if err != nil {
		switch {
		case errors.Is(err, accountrecovery.ErrNotFound),
			errors.Is(err, accountrecovery.ErrExpired),
			errors.Is(err, accountrecovery.ErrNotVerified):
			cookie.ClearRecoverSession(c)
			return showErrorTo(c, ui.RecoverPasswordErrorID,
				"Recovery session expired. Please start again.")
		default:
			return showErrorTo(c, ui.RecoverPasswordErrorID, err.Error())
		}
	}

	cookie.ClearRecoverSession(c)
	c.Set("HX-Redirect", "/login")
	return c.SendStatus(fiber.StatusOK)
}
