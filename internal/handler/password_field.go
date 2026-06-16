package handler

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ui"
)

var passwordFieldNames = []string{
	"password",
	"password2",
	"current_password",
	"new_password",
	"confirm_password",
}

var passwordFieldAutocompletes = []string{
	"current-password",
	"new-password",
	"off",
}

func PasswordFieldHandler(c *fiber.Ctx) error {
	name := c.Query("name")
	if !slices.Contains(passwordFieldNames, name) {
		return fiber.ErrBadRequest
	}
	autocomplete := c.Query("autocomplete")
	if !slices.Contains(passwordFieldAutocompletes, autocomplete) {
		return fiber.ErrBadRequest
	}
	visible := c.Query("visible") == "true"
	return render(c, ui.PasswordField(ui.PasswordFieldView{
		Name:         name,
		Autocomplete: autocomplete,
		Value:        c.FormValue(name),
		Visible:      visible,
	}))
}
