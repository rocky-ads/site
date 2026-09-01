package ui

import (
	"bytes"
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

func TestUserProfilePageShareButton(t *testing.T) {
	d := UserProfileData{
		ID:                1,
		Name:              "test",
		MemberSince:       "January 2, 2006",
		AdsViewPathPrefix: "/auth/user/1/view/",
		ShowShare:         true,
	}
	var buf bytes.Buffer
	if err := g.Group(UserProfilePage(d, ViewList, nil)).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	if !strings.Contains(html, "/auth/user/share") {
		t.Fatal("missing share button")
	}
	if !strings.Contains(html, "/auth/user/1/view/") {
		t.Fatal("missing auth view prefix")
	}

	d.ShowShare = false
	d.AdsViewPathPrefix = "/u/tokentokentokentoke/view/"
	buf.Reset()
	if err := g.Group(UserProfilePage(d, ViewList, nil)).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html = buf.String()
	if strings.Contains(html, "/auth/user/share") {
		t.Fatal("guest page should not share")
	}
	if !strings.Contains(html, "/u/tokentokentokentoke/view/") {
		t.Fatal("missing shared view prefix")
	}
}
