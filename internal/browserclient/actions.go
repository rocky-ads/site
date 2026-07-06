package browserclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

const formPostTimeout = 45 * time.Second

// ActGet loads a full page. Use ActClick for in-page HTMX controls.
func (c *Client) ActGet(path string) error {
	return c.Navigate(path)
}

// ActClick triggers an in-page HTMX control (tabs, swaps, modals).
func (c *Client) ActClick(path string) error {
	if id, ok := ConversationIDFromOpenPath(path); ok {
		itemSel := fmt.Sprintf(`#conversation-item-%d`, id)
		if el, err := c.firstVisible(itemSel); err == nil {
			return c.clickElement(el, path)
		}
	}
	for _, attr := range []string{"hx-get", "hx-post", "hx-delete"} {
		if el, err := c.firstVisible(attrSelector("", attr, path)); err == nil {
			return c.clickElement(el, path)
		}
	}
	if el, err := c.firstVisible(attrSelector("a", "href", path)); err == nil {
		if hx, _ := el.Attribute("hx-get"); hx != nil && *hx != "" {
			return c.clickElement(el, path)
		}
	}
	return fmt.Errorf("no HTMX control for %s", path)
}

func (c *Client) clickElement(el *rod.Element, path string) error {
	wait := c.page.Timeout(10*time.Second).WaitRequestIdle(
		1*time.Second, []string{path}, nil, nil)
	if _, err := el.Eval(`() => this.click()`); err != nil {
		return err
	}
	wait()
	return c.waitSettle()
}

// ActPostForm fills fields and submits a matching form.
func (c *Client) ActPostForm(path string, fields map[string]string) error {
	if IsMessageSendPath(path) {
		return c.actPostMessage(path, fields)
	}
	form, err := c.findForm(path)
	if err != nil {
		return err
	}
	for name, value := range fields {
		if err := c.setField(form, name, value); err != nil {
			if isMissingField(err) {
				continue
			}
			return fmt.Errorf("field %s: %w", name, err)
		}
	}
	submit, err := form.Timeout(queryTimeout).Element("button[type=submit], input[type=submit]")
	if err != nil {
		submit, err = form.Timeout(queryTimeout).Element("button:not([type=button])")
	}
	if err != nil {
		return fmt.Errorf("submit button: %w", err)
	}
	wait := c.page.Timeout(formPostTimeout).WaitRequestIdle(
		1*time.Second, []string{path}, nil, nil)
	if _, err := submit.Eval(`() => this.click()`); err != nil {
		return err
	}
	wait()
	if err := c.waitSettle(); err != nil {
		return err
	}
	if msg := c.visibleError(); msg != "" {
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func (c *Client) findForm(path string) (*rod.Element, error) {
	if id, ok := ConversationIDFromSendPath(path); ok {
		sel := fmt.Sprintf(`#conversation-%d-form`, id)
		if form, err := c.page.Timeout(10 * time.Second).Element(sel); err == nil {
			return form, nil
		}
		modalSel := fmt.Sprintf(`#conversation-%d-modal form[hx-post="%s"]`, id, path)
		if form, err := c.firstVisible(modalSel); err == nil {
			return form, nil
		}
	}
	sel := attrSelector("form", "hx-post", path)
	if form, err := c.firstVisible(sel); err == nil {
		return form, nil
	}
	return c.firstVisible(attrSelector("form", "action", path))
}

func (c *Client) formField(form *rod.Element, name string) (*rod.Element, error) {
	return form.Timeout(queryTimeout).Element(fmt.Sprintf(`[name="%s"]`, name))
}

func (c *Client) setField(form *rod.Element, name, value string) error {
	el, err := c.formField(form, name)
	if err != nil {
		return err
	}
	tag, err := el.Property("tagName")
	if err != nil {
		return err
	}
	switch strings.ToUpper(tag.String()) {
	case "SELECT":
		return el.Select([]string{value}, true, rod.SelectorTypeText)
	case "INPUT":
		inputType, _ := el.Attribute("type")
		if inputType != nil {
			switch *inputType {
			case "checkbox":
				return c.setCheckbox(el, checkboxChecked(value))
			case "hidden", "file":
				return nil
			}
		}
	case "TEXTAREA":
		_, err := el.Eval(fmt.Sprintf(`(v) => { this.value = %q; this.dispatchEvent(new Event("input", {bubbles:true})); }`, value))
		return err
	}
	return el.Input(value)
}

func checkboxChecked(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "false", "0", "off":
		return false
	default:
		return true
	}
}

func (c *Client) setCheckbox(el *rod.Element, checked bool) error {
	prop, err := el.Property("checked")
	if err != nil {
		return err
	}
	isChecked := prop.Bool()
	if isChecked == checked {
		return nil
	}
	_, err = el.Eval(`() => this.click()`)
	return err
}

func (c *Client) firstVisible(selector string) (*rod.Element, error) {
	els, err := c.page.Elements(selector)
	if err != nil {
		return nil, err
	}
	for _, el := range els {
		visible, err := el.Visible()
		if err == nil && visible {
			return el, nil
		}
	}
	if len(els) == 0 {
		return nil, fmt.Errorf("no elements for %s", selector)
	}
	return els[0], nil
}

func isMissingField(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "not found") ||
		strings.Contains(s, "no elements") ||
		strings.Contains(s, "context deadline exceeded")
}

func attrSelector(tag, attr, path string) string {
	if tag == "" {
		return fmt.Sprintf(`[%s="%s"]`, attr, path)
	}
	return fmt.Sprintf(`%s[%s="%s"]`, tag, attr, path)
}

func (c *Client) waitForForm(postPath string) error {
	sel := attrSelector("form", "hx-post", postPath)
	_, err := c.page.Timeout(10 * time.Second).Element(sel)
	return err
}

func (c *Client) waitForPath(pathPrefix string) error {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if strings.HasPrefix(c.CurrentPath(), pathPrefix) {
			return c.waitSettle()
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("expected %s, at %s", pathPrefix, c.CurrentPath())
}
