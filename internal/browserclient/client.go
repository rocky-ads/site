package browserclient

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	queryTimeout = 500 * time.Millisecond
	observeGap   = 200 * time.Millisecond
	sseSettleGap = 400 * time.Millisecond
)

// Client drives a headless browser for one agent.
type Client struct {
	baseURL string
	root    *url.URL
	browser *rod.Browser
	page    *rod.Page
}

// New creates a client; call Start before use.
func New(baseURL string) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	return &Client{baseURL: parsed.String(), root: parsed}, nil
}

// Start launches Chromium and opens the site home page.
func (c *Client) Start() error {
	l := launcher.New().Headless(true).Set("disable-images")
	controlURL, err := l.Launch()
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("connect browser: %w", err)
	}
	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		browser.Close()
		return fmt.Errorf("new page: %w", err)
	}
	c.browser = browser
	c.page = page
	return c.Navigate("/")
}

// Close shuts down the browser.
func (c *Client) Close() {
	if c.browser != nil {
		c.browser.Close()
		c.browser = nil
		c.page = nil
	}
}

// CurrentPath returns the browser location path (with query string).
func (c *Client) CurrentPath() string {
	if c.page == nil {
		return "/"
	}
	info, err := c.page.Timeout(3 * time.Second).Info()
	if err != nil {
		return "/"
	}
	u, err := url.Parse(info.URL)
	if err != nil || u.Host != c.root.Host {
		return "/"
	}
	p := u.EscapedPath()
	if p == "" {
		p = "/"
	}
	if u.RawQuery != "" {
		p += "?" + u.RawQuery
	}
	return p
}

// Observation is the composed DOM state for planning.
type Observation struct {
	Path string
	Page PageAffordances
}

// Observe reads the live DOM and parses affordances.
func (c *Client) Observe() (Observation, error) {
	path := c.CurrentPath()
	html, err := c.pageHTML()
	if err != nil {
		return Observation{Path: path, Page: PageAffordances{URL: path}}, nil
	}
	body := []byte(html)
	page := ParsePage(body, path)
	page.Conversations = ParseAllConversationsHTML(body)
	if conv := BestConversation(page.Conversations); conv != nil {
		page = EnrichWithConversation(page, *conv)
	}
	time.Sleep(observeGap)
	return Observation{Path: path, Page: page}, nil
}

func (c *Client) pageHTML() (string, error) {
	type result struct {
		html string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		depth := -1
		doc, err := proto.DOMGetDocument{Depth: &depth, Pierce: true}.Call(c.page)
		if err != nil {
			ch <- result{"", err}
			return
		}
		html, err := proto.DOMGetOuterHTML{NodeID: doc.Root.NodeID}.Call(c.page)
		if err != nil {
			ch <- result{"", err}
			return
		}
		ch <- result{html.OuterHTML, nil}
	}()
	select {
	case r := <-ch:
		return r.html, r.err
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("page html: timed out")
	}
}

// Navigate loads a full page by path.
func (c *Client) Navigate(path string) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if err := c.page.Timeout(15 * time.Second).Navigate(c.baseURL + path); err != nil {
		return err
	}
	if err := c.page.Timeout(15 * time.Second).WaitLoad(); err != nil {
		return err
	}
	return c.waitSettle()
}

func (c *Client) waitSettle() error {
	wait := c.page.Timeout(3 * time.Second)
	_, _ = wait.Eval(`() => new Promise(resolve => {
		const deadline = Date.now() + 3000;
		const check = () => {
			const busy = document.documentElement.classList.contains('htmx-request')
				|| document.querySelector('.htmx-request');
			if (!busy || Date.now() > deadline) {
				resolve(true);
				return;
			}
			setTimeout(check, 100);
		};
		check();
	})`)
	time.Sleep(sseSettleGap)
	return nil
}
