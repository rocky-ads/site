package browserclient

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Link is a navigable href on the page.
type Link struct {
	Href  string `json:"href"`
	Label string `json:"label"`
}

// FormField is an input/select/textarea name on a form.
type FormField struct {
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// Form is a submit target on the page.
type Form struct {
	Action string      `json:"action"`
	Method string      `json:"method"`
	Fields []FormField `json:"fields"`
}

// HTMXAction is an hx-get/post/delete target.
type HTMXAction struct {
	Kind  string `json:"kind"`
	Path  string `json:"path"`
	Label string `json:"label,omitempty"`
}

// AdCard is a detected ad link.
type AdCard struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

// PageAffordances is a compact view of a page for agent planning.
type PageAffordances struct {
	URL                     string                 `json:"url"`
	Title                   string                 `json:"title,omitempty"`
	Error                   string                 `json:"error,omitempty"`
	LoggedIn                bool                   `json:"logged_in,omitempty"`
	HasUnreadMessages       bool                   `json:"has_unread_messages,omitempty"`
	CurrentCategory         string                 `json:"current_category,omitempty"`
	Links                   []Link                 `json:"links,omitempty"`
	Forms                   []Form                 `json:"forms,omitempty"`
	HTMX                    []HTMXAction           `json:"htmx,omitempty"`
	AdCards                 []AdCard               `json:"ad_cards,omitempty"`
	UnreadConversationPaths []string               `json:"unread_conversation_paths,omitempty"`
	OpenConversationForms   []int                  `json:"open_conversation_forms,omitempty"`
	OpenAdMessageSendPaths  []string               `json:"open_ad_message_send_paths,omitempty"`
	Conversations           []ConversationSnapshot `json:"conversations,omitempty"`
	Conversation            *ConversationSnapshot  `json:"conversation,omitempty"`
}

var adPathRe = regexp.MustCompile(`^/ad/(\d+)`)

// ParsePage extracts affordances from live DOM HTML.
func ParsePage(html []byte, currentURL string) PageAffordances {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return PageAffordances{URL: currentURL, Error: "parse failed"}
	}

	p := PageAffordances{URL: currentURL}
	p.Title = strings.TrimSpace(doc.Find("title").First().Text())
	p.LoggedIn = doc.Find("#user-avatar-container").Length() > 0 ||
		doc.Find(`[sse-connect="/auth/sse"]`).Length() > 0
	p.HasUnreadMessages = doc.Find("#message-unread-indicator .bg-green-500").Length() > 0
	doc.Find("[id^='conversation-item-']").Each(func(_ int, s *goquery.Selection) {
		if s.Find(".bg-green-500").Length() == 0 {
			return
		}
		path, ok := s.Attr("hx-get")
		if ok && path != "" {
			p.UnreadConversationPaths = append(p.UnreadConversationPaths, path)
		}
	})
	doc.Find("[id^='conversation-'][id$='-form']").Each(func(_ int, s *goquery.Selection) {
		idAttr, ok := s.Attr("id")
		if !ok {
			return
		}
		var id int
		if _, err := fmt.Sscanf(idAttr, "conversation-%d-form", &id); err != nil {
			return
		}
		if s.Find(`[name="content"]:not([disabled])`).Length() > 0 {
			p.OpenConversationForms = append(p.OpenConversationForms, id)
		}
	})
	doc.Find(`form[hx-post^="/auth/ad/"][hx-post$="/send"]`).Each(func(_ int, s *goquery.Selection) {
		path, ok := s.Attr("hx-post")
		if !ok || path == "" {
			return
		}
		if s.Find(`[name="content"]:not([disabled])`).Length() > 0 {
			p.OpenAdMessageSendPaths = append(p.OpenAdMessageSendPaths, path)
		}
	})
	if btn := doc.Find(`button[hx-get*="category-select"]`).First(); btn.Length() > 0 {
		p.CurrentCategory = strings.TrimSpace(btn.Find("span").First().Text())
	}

	errText := strings.TrimSpace(doc.Find("#error, .error, [role=alert], .text-red-500").First().Text())
	if errText != "" {
		p.Error = errText
	}

	seenLinks := map[string]bool{}
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || href == "" || strings.HasPrefix(href, "#") {
			return
		}
		if !strings.HasPrefix(href, "/") {
			return
		}
		if seenLinks[href] {
			return
		}
		seenLinks[href] = true
		label := strings.TrimSpace(s.Text())
		p.Links = append(p.Links, Link{Href: href, Label: label})

		if m := adPathRe.FindStringSubmatch(href); len(m) == 2 {
			p.AdCards = append(p.AdCards, AdCard{ID: m[1], Title: label})
		}
	})

	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		if action == "" {
			action, _ = s.Attr("hx-post")
		}
		method, _ := s.Attr("method")
		if method == "" {
			method = "POST"
		}
		f := Form{Action: action, Method: strings.ToUpper(method)}
		s.Find("input, select, textarea").Each(func(_ int, el *goquery.Selection) {
			name, ok := el.Attr("name")
			if !ok || name == "" {
				return
			}
			typ, _ := el.Attr("type")
			val, _ := el.Attr("value")
			f.Fields = append(f.Fields, FormField{Name: name, Type: typ, Value: val})
		})
		if f.Action != "" || len(f.Fields) > 0 {
			p.Forms = append(p.Forms, f)
		}
	})

	addHTMX := func(kind, path, label string) {
		if path == "" || !strings.HasPrefix(path, "/") {
			return
		}
		p.HTMX = append(p.HTMX, HTMXAction{Kind: kind, Path: path, Label: label})
	}

	doc.Find("[hx-get]").Each(func(_ int, s *goquery.Selection) {
		path, _ := s.Attr("hx-get")
		addHTMX("get", path, strings.TrimSpace(s.Text()))
	})
	doc.Find("[hx-post]").Each(func(_ int, s *goquery.Selection) {
		path, _ := s.Attr("hx-post")
		addHTMX("post", path, strings.TrimSpace(s.Text()))
	})
	doc.Find("[hx-delete]").Each(func(_ int, s *goquery.Selection) {
		path, _ := s.Attr("hx-delete")
		addHTMX("delete", path, strings.TrimSpace(s.Text()))
	})

	return p
}

// AllowedPaths returns all relative paths from affordances.
func (p PageAffordances) AllowedPaths() map[string]bool {
	out := map[string]bool{}
	for _, l := range p.Links {
		out[l.Href] = true
	}
	for _, f := range p.Forms {
		if f.Action != "" {
			out[f.Action] = true
		}
	}
	for _, h := range p.HTMX {
		out[h.Path] = true
	}
	return out
}
