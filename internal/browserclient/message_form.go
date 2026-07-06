package browserclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func messageInputSelectors(postPath string) []string {
	var sels []string
	if id, ok := ConversationIDFromSendPath(postPath); ok {
		sels = append(sels, fmt.Sprintf(`#conversation-%d-content-input:not([disabled])`, id))
	}
	if _, ok := AdIDFromSendPath(postPath); ok {
		sels = append(sels,
			fmt.Sprintf(`form[hx-post="%s"] [name="content"]:not([disabled])`, postPath),
			`#conversation-0-content-input:not([disabled])`,
		)
	}
	return sels
}

func adMessageInputSelectors(adID int) []string {
	sendPath := fmt.Sprintf("/auth/ad/%d/send", adID)
	return append(messageInputSelectors(sendPath),
		`[id^="conversation-"][id$="-content-input"]:not([disabled])`,
	)
}

func (c *Client) actPostMessage(postPath string, fields map[string]string) error {
	content, ok := fields["content"]
	if !ok || strings.TrimSpace(content) == "" {
		return fmt.Errorf("missing content field")
	}
	input, err := c.waitVisibleMessageInput(messageInputSelectors(postPath), 10*time.Second)
	if err != nil {
		return fmt.Errorf("field content: %w", err)
	}
	wait := c.page.Timeout(formPostTimeout).WaitRequestIdle(
		1*time.Second, []string{postPath}, nil, nil)
	_, err = input.Eval(`(v) => {
		this.value = v;
		this.dispatchEvent(new Event("input", {bubbles: true}));
		const form = this.closest("form");
		if (!form) throw new Error("no parent form");
		const btn = form.querySelector("button[type=submit]");
		if (!btn) throw new Error("no submit button");
		btn.click();
	}`, content)
	if err != nil {
		return fmt.Errorf("field content: %w", err)
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

func (c *Client) waitVisibleMessageInput(selectors []string, timeout time.Duration) (*rod.Element, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if input, err := c.lastVisibleElement(selectors); err == nil {
			return input, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("cannot find element")
}

func (c *Client) lastVisibleElement(selectors []string) (*rod.Element, error) {
	for _, sel := range selectors {
		els, err := c.page.Elements(sel)
		if err != nil || len(els) == 0 {
			continue
		}
		for i := len(els) - 1; i >= 0; i-- {
			vis, err := els[i].Visible()
			if err == nil && vis {
				return els[i], nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find element")
}

// WaitForConversationForm blocks until the reply input for a conversation exists.
func (c *Client) WaitForConversationForm(convID int) error {
	sels := []string{fmt.Sprintf(`#conversation-%d-content-input:not([disabled])`, convID)}
	_, err := c.waitVisibleMessageInput(sels, 10*time.Second)
	return err
}

// WaitForAdMessageForm blocks until a message input is available on an ad page.
func (c *Client) WaitForAdMessageForm(adID int) error {
	_, err := c.waitVisibleMessageInput(adMessageInputSelectors(adID), 10*time.Second)
	return err
}
