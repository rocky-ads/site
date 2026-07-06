package browserclient

import (
	"fmt"
	"strings"
)

const authCookieName = "auth_token"

func (c *Client) hasAuthCookie() bool {
	if c.page == nil || c.browser == nil {
		return false
	}
	cookies, err := c.browser.GetCookies()
	if err != nil {
		return false
	}
	for _, ck := range cookies {
		if ck.Name == authCookieName && ck.Value != "" {
			return true
		}
	}
	return false
}

func (c *Client) visibleError() string {
	obs, err := c.Observe()
	if err != nil {
		return ""
	}
	return PageError(obs.Page)
}

// SessionActive reports whether the browser has an active login session.
func (c *Client) SessionActive() bool {
	if c.hasAuthCookie() {
		return true
	}
	if strings.HasPrefix(c.CurrentPath(), "/auth/") {
		return true
	}
	obs, err := c.Observe()
	if err != nil {
		return c.hasSelector("#user-avatar-container")
	}
	if PageLooksLoggedIn(obs.Page) {
		return true
	}
	return !PageLooksLoggedOut(obs.Page)
}

func (c *Client) hasSelector(sel string) bool {
	_, err := c.page.Timeout(queryTimeout).Element(sel)
	return err == nil
}

// requireAuthPage navigates to path and fails if redirected to login.
func (c *Client) requireAuthPage(path string) error {
	if err := c.Navigate(path); err != nil {
		return err
	}
	if strings.HasPrefix(c.CurrentPath(), "/login") {
		if msg := c.visibleError(); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("authentication failed")
	}
	return nil
}

// Login submits the login form in the browser.
func (c *Client) Login(username, password string) error {
	if c.SessionActive() {
		return nil
	}
	if err := c.Navigate("/login"); err != nil {
		return err
	}
	if c.SessionActive() {
		return c.Navigate("/")
	}
	if err := c.ActPostForm("/api/login", map[string]string{
		"username": username,
		"password": password,
	}); err != nil {
		if c.SessionActive() {
			return c.Navigate("/")
		}
		return err
	}
	if msg := c.visibleError(); msg != "" {
		return fmt.Errorf("login: %s", msg)
	}
	return c.requireAuthPage("/")
}

// RegisterTestUser completes registration for a +1555010xxxx phone.
func (c *Client) RegisterTestUser(username, phone, password string) error {
	if err := c.Navigate("/register"); err != nil {
		return err
	}
	if err := c.ActPostForm("/api/register/step1", map[string]string{
		"username": username,
		"phone":    phone,
		"offers":   "true",
	}); err != nil {
		return err
	}
	if msg := c.visibleError(); msg != "" {
		return fmt.Errorf("register step1: %s", msg)
	}
	if err := c.waitForForm("/api/register/step3"); err != nil {
		return fmt.Errorf("register step1: password form not shown (%w)", err)
	}
	if err := c.ActPostForm("/api/register/step3", map[string]string{
		"password":  password,
		"password2": password,
		"terms":     "accepted",
	}); err != nil {
		return err
	}
	if msg := c.visibleError(); msg != "" {
		return fmt.Errorf("register step3: %s", msg)
	}
	return c.requireAuthPage("/auth/welcome")
}

// PageError returns a user-visible error from the current DOM, if any.
func PageError(page PageAffordances) string {
	return strings.TrimSpace(page.Error)
}
