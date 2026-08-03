package handler

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/nyaruka/phonenumbers"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/password"
	"github.com/rocky-ads/site/internal/phoneverification"
	"github.com/rocky-ads/site/internal/registrationticket"
	"github.com/rocky-ads/site/internal/service/grok"
	"github.com/rocky-ads/site/internal/service/sms"
	"github.com/rocky-ads/site/internal/ui"
	"github.com/rocky-ads/site/internal/user"
)

// RegistrationRateLimiter is a strict rate limiter for registration (per IP)
var RegistrationRateLimiter = limiter.New(limiter.Config{
	Max:        config.EffectiveRegistrationRateLimitMax(),
	Expiration: config.RegistrationRateLimitExp,
	KeyGenerator: func(c *fiber.Ctx) string {
		// Rate limit per IP address
		return c.IP()
	},
	LimitReached: func(c *fiber.Ctx) error {
		minutes := int(config.RegistrationRateLimitExp.Minutes())
		errorMsg := fmt.Sprintf("Too many registration attempts. "+
			"Please try again in %d minutes.", minutes)
		return showError(c, errorMsg)
	},
})

func RegisterHandler(c *fiber.Ctx) error {
	logout(c)
	cookie.ClearRegisterTicket(c)
	return renderPage(c, "Register", ui.RegisterPage(c.Query("username")))
}

// validateUsername validates that a username follows conventions
// Rules:
// - 3 to 20 characters
// - Only letters (a-z, A-Z) and digits (0-9)
// - First character must be a letter (a-z, A-Z)
func validateUsername(username string) error {
	validPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{2,19}$`)
	if !validPattern.MatchString(username) {
		return fmt.Errorf("username must be 3-20 characters, start with a letter, and contain only letters and digits")
	}
	return nil
}

func validatePhone(phone string) (string, error) {

	// Try parsing as international format first (handles formatting automatically)
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		// If parsing fails and number doesn't start with +, try US region
		// Note: "US" region also validates Canadian numbers (both share +1 and NANP)
		if !strings.HasPrefix(phone, "+") {
			num, err = phonenumbers.Parse(phone, "US")
			if err != nil {
				return "", fmt.Errorf("invalid phone number format")
			}
		} else {
			return "", fmt.Errorf("phone must be in international format, e.g. +12025550123")
		}
	}

	// Format in E.164 (e.g., +15551234567)
	e164 := phonenumbers.Format(num, phonenumbers.E164)

	// +1555010xxxx are fictional test numbers; libphonenumber rejects them but
	// registration accepts them when ALLOW_TEST_REGISTRATION is on.
	if allowTestRegistration(e164) {
		return e164, nil
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number")
	}

	return e164, nil
}

func allowTestRegistration(phoneE64 string) bool {
	return config.AllowTestRegistration && user.IsTestPhoneE64(phoneE64)
}

func screenUsername(username string) (string, error) {

	systemPrompt := `Your job is to screen potential user names for a web site.
Reject user names that the general public would find offensive or inappropriate.
The user name is displayed on the site for other to see and interact with, so we
want polite names.

Unacceptable usernames:
- racial slurs
- hate speech
- explicit sexual content

If the user name is acceptable, return only: OK

If the user name is unacceptable, return a short, direct error message (1-2
sentences), and do not mention yourself, AI, or Grok in the response.`

	userPrompt := `Screen the following user name for the web site: ` + username

	resp, err := grok.CallGrok(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf(
			"unable to complete registration with these credentials. Please try different information.")
	}
	if resp != "OK" {
		return "", fmt.Errorf("%s", resp)
	}
	return resp, nil
}

// RegisterStep1Handler handles the first step submission and sends SMS
func RegisterStep1Handler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	phone := c.FormValue("phone")

	if err := validateUsername(username); err != nil {
		return showError(c, err.Error())
	}

	phoneE64, err := validatePhone(phone)
	if err != nil {
		return showError(c, err.Error())
	}

	if user.IsTestPhoneE64(phoneE64) && !config.AllowTestRegistration {
		return showError(c,
			"Unable to complete registration with these credentials. Please try different information.")
	}

	// Validate required checkbox
	offers := c.FormValue("offers")
	if offers != "true" {
		return showError(c, "You must agree to receive informational text messages to continue.")
	}

	// TODO timing attack protection

	taken, err := user.UsernameTaken(username)
	if err != nil {
		logger.Error("Failed to check username", "error", err)
		return showError(c, "Unable to complete registration. Please try again.")
	}
	if taken {
		return showError(c,
			"Unable to complete registration with these credentials. Please try different information.")
	}

	avail, err := user.CheckPhoneAvailability(phoneE64, 0)
	if err != nil {
		logger.Error("Failed to check phone", "error", err)
		return showError(c, "Unable to complete registration. Please try again.")
	}
	switch avail.Status {
	case user.PhoneActive:
		return showError(c,
			"Unable to complete registration with these credentials. Please try different information.")
	case user.PhoneHeld:
		return showError(c, (&user.PhoneHoldError{
			DaysRemaining: avail.DaysRemaining,
		}).Error())
	}

	resp, err := screenUsername(username)
	if err != nil {
		return showError(c, err.Error())
	}
	if resp != "OK" {
		return showError(c, resp)
	}

	if allowTestRegistration(phoneE64) {
		if err := registrationticket.Issue(c, username, phoneE64); err != nil {
			logger.Error("Failed to issue registration ticket",
				"error", err, "phone", phoneE64)
			return showError(c, "Unable to complete registration. Please try again.")
		}
		return render(c, ui.RegisterPassword(username, phoneE64))
	}

	code, err := phoneverification.GenerateCode()
	if err != nil {
		logger.Error("Failed to generate verification code",
			"error", err)
		return showError(c, "Unable to generate verification code. Please try again.")
	}

	err = phoneverification.StoreCode(phoneE64, code,
		phoneverification.PurposeRegister, nil)
	if err != nil {
		logger.Error("Failed to store verification code",
			"error", err, "phone", phoneE64)
		return showError(c, "Unable to create verification code. Please try again.")
	}

	// Send SMS
	message := fmt.Sprintf("Your %s verification code is: %s. "+
		"This code expires in 10 minutes. Reply STOP to unsubscribe.", config.ServerName, code)
	err = sms.SendMessage(phoneE64, message)
	if err != nil {
		logger.Error("Failed to send SMS", "error", err, "phone", phoneE64)
		// Check if this is a blocked number error
		if errors.Is(err, sms.ErrBlockedNumber) {
			unstopMessage := fmt.Sprintf(
				"This phone number was previously opted out of text messages. "+
					"To receive verification codes, please reply UNSTOP to %s from this phone number, then try registering again.",
				config.TwilioFromNumber)
			return showError(c, unstopMessage)
		}
		return showError(c, "Unable to send verification code. Please try again.")
	}

	return render(c, ui.RegisterVerify(username, phoneE64))
}

func RegisterStep2Handler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	phoneE64 := c.FormValue("phone")
	code := c.FormValue("code")

	if err := validateUsername(username); err != nil {
		return c.Redirect("/register")
	}

	_, err := validatePhone(phoneE64)
	if err != nil {
		return showError(c, err.Error())
	}

	if code == "" {
		return showError(c, "Please enter the verification code")
	}

	if allowTestRegistration(phoneE64) {
		if err := registrationticket.Issue(c, username, phoneE64); err != nil {
			logger.Error("Failed to issue registration ticket",
				"error", err, "phone", phoneE64)
			return showError(c, "Unable to complete registration. Please try again.")
		}
		return render(c, ui.RegisterPassword(username, phoneE64))
	}

	ok, err := phoneverification.ConsumeCode(phoneE64, code,
		phoneverification.PurposeRegister, nil)
	if err != nil || !ok {
		logger.Warn("Verification code consume failed",
			"error", err, "phone", phoneE64)
		return showError(c, "Invalid or expired verification code. Please request a new code.")
	}

	if err := registrationticket.Issue(c, username, phoneE64); err != nil {
		logger.Error("Failed to issue registration ticket",
			"error", err, "phone", phoneE64)
		return showError(c, "Unable to complete registration. Please try again.")
	}

	return render(c, ui.RegisterPassword(username, phoneE64))
}

func RegisterStep3Handler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	phoneE64 := c.FormValue("phone")
	passwd := c.FormValue("password")
	passwd2 := c.FormValue("password2")
	terms := c.FormValue("terms")

	if err := validateUsername(username); err != nil {
		return c.Redirect("/register")
	}

	_, err := validatePhone(phoneE64)
	if err != nil {
		return showError(c, err.Error())
	}

	if err := password.ValidatePasswordConfirmation(passwd, passwd2); err != nil {
		return showError(c, err.Error())
	}

	if err := password.ValidatePasswordStrength(passwd); err != nil {
		return showError(c, err.Error())
	}

	if terms != "accepted" {
		return showError(c, "You must accept the terms and conditions to continue.")
	}

	if err := registrationticket.Consume(c, username, phoneE64); err != nil {
		logger.Warn("Registration ticket consume failed",
			"error", err, "phone", phoneE64)
		return showError(c,
			"Invalid or expired registration. Please start over.")
	}

	u, err := user.CreateUser(username, phoneE64, passwd)
	if err != nil {
		var holdErr *user.PhoneHoldError
		if errors.As(err, &holdErr) {
			return showError(c, holdErr.Error())
		}
		if errors.Is(err, user.ErrUserAlreadyExists) {
			return showError(c,
				"Unable to complete registration with these credentials. Please try different information.")
		}
		logger.Error("Failed to create user",
			"error", err, "phone", phoneE64)
		return showError(c, "Unable to complete registration. Please try again.")
	}

	// Generate JWT token and log user in
	token, err := generateJWTToken(&u)
	if err != nil {
		logger.Error("Failed to generate token",
			"error", err, "userID", u.ID)
		return showError(c, "Registration successful, but login failed. Please try logging in.")
	}

	cookie.SetJWT(c, token)
	cookie.SetDistanceUnitForUser(c, u.PhoneE64)

	c.Set("HX-Redirect", "/auth/welcome")
	return c.SendStatus(fiber.StatusOK)
}

func WelcomeHandler(c *fiber.Ctx) error {
	rockCount := config.MaxOutstandingRocks
	return renderPage(c, "Welcome", ui.WelcomePage(rockCount))
}
