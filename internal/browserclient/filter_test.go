package browserclient

import "testing"

func TestFilterSensitiveForms(t *testing.T) {
	p := PageAffordances{
		Forms: []Form{{
			Action: "/auth/user/settings/password",
			Method: "POST",
		}, {
			Action: "/auth/user/settings/notifications",
			Method: "POST",
		}},
	}
	out := FilterSensitiveForms(p)
	if len(out.Forms) != 1 || out.Forms[0].Action != "/auth/user/settings/notifications" {
		t.Fatalf("forms %v", out.Forms)
	}
}

func TestFilterHTMXPrefix(t *testing.T) {
	p := PageAffordances{
		HTMX: []HTMXAction{
			{Path: "/auth/user/myads/tab/active"},
			{Path: "/auth/user/myads/tab/deleted"},
		},
	}
	out := FilterHTMXPrefix(p, "/auth/user/myads/tab/")
	if len(out.HTMX) != 0 {
		t.Fatalf("htmx %v", out.HTMX)
	}
}
