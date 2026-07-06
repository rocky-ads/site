package ui

import (
	"github.com/rocky-ads/site/internal/config"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type faqSectionData struct {
	title string
	body  func() []g.Node
}

var faqOrder = []string{
	"rocks",
	"phone-number",
	"sms-notifications",
}

var faqSections = map[string]faqSectionData{
	"rocks": {
		title: "What are rocks for?",
		body: func() []g.Node {
			return []g.Node{
				faqParagraph(g.Textf(
					"When you join %s, you receive 3 rocks.",
					config.ServerName,
				)),
				faqParagraph(g.Text(
					"Throw a rock on an ad that violates our policies " +
						"or has problems. That starts a conversation " +
						"with the ad owner so you can work together to " +
						"resolve the issue.",
				)),
				faqParagraph(g.Text(
					"Once the issue is resolved, the seller can return " +
						"your rock. Rocks are limited — use them wisely.",
				)),
			}
		},
	},
	"phone-number": {
		title: "Why do you need my phone number?",
		body: func() []g.Node {
			return []g.Node{
				faqParagraph(g.Textf(
					"%s connects buyers and sellers through "+
						"messaging. We need your phone number mainly "+
						"so we can notify you when someone sends you "+
						"a message.",
					config.ServerName,
				)),
				faqParagraph(g.Textf(
					"Ideally we would not collect personal "+
						"information at all, but we do need a "+
						"reliable way to reach you. %s is built to "+
						"collect as little as possible while still "+
						"working.",
					config.ServerName,
				)),
				faqParagraph(g.Text(
					"A text-capable phone number is a practical " +
						"choice: texting is everywhere, and it helps " +
						"tie each account to a real person. We do not " +
						"ask for your email, mailing address, payment " +
						"details, or other profile information — your " +
						"phone number is the only contact info we " +
						"collect.",
				)),
				faqParagraph(
					g.Text(
						"During signup we send a one-time "+
							"verification code to confirm the number "+
							"can receive texts. After that, we text "+
							"you when other users message you. You "+
							"can turn text notifications on or off "+
							"anytime in ",
					),
					faqLink("/auth/user/settings", "Settings"),
					g.Text("."),
				),
				faqParagraph(g.Text(
					"We do not sell your phone number, share it " +
						"with other users, or use it for marketing " +
						"unrelated to your account.",
				)),
			}
		},
	},
	"sms-notifications": {
		title: "Why am I not getting text messages?",
		body: func() []g.Node {
			fromNumber := config.TwilioFromNumber
			return []g.Node{
				faqParagraph(
					g.Text("First, open "),
					faqLink("/auth/user/settings", "Settings"),
					g.Text(" and confirm text messages are turned on."),
				),
				smsResumeNote(fromNumber),
				faqParagraph(
					g.Text(
						"You can also turn notifications off or on "+
							"anytime from ",
					),
					faqLink("/auth/user/settings", "Settings"),
					g.Text(" without texting STOP or START."),
				),
			}
		},
	},
}

func ValidFAQSection(id string) bool {
	_, ok := faqSections[id]
	return ok
}

func faqLink(href, text string) g.Node {
	return A(
		Href(href),
		Class("text-blue-600 dark:text-blue-400 hover:underline"),
		g.Text(text),
	)
}

func faqParagraph(children ...g.Node) g.Node {
	attrs := []g.Node{
		Class("text-sm text-zinc-600 dark:text-zinc-400"),
	}
	attrs = append(attrs, children...)
	return P(attrs...)
}

func faqNote(children ...g.Node) g.Node {
	attrs := []g.Node{
		Class("text-sm text-zinc-600 dark:text-zinc-400 " +
			"border-l-2 border-zinc-300 dark:border-zinc-600 pl-3"),
	}
	attrs = append(attrs, children...)
	return Div(attrs...)
}

func smsResumeNote(fromNumber string) g.Node {
	if fromNumber == "" {
		return faqNote(g.Text(
			"If you previously replied STOP to a text message from us, " +
				"your carrier may have blocked further texts. " +
				"Text START to our number to resume delivery.",
		))
	}
	return faqNote(
		g.Text(
			"If you previously replied STOP to a text message "+
				"from us, your carrier may have blocked further "+
				"texts. Text START to ",
		),
		Span(Class("font-mono"), g.Text(fromNumber)),
		g.Text(" to resume delivery."),
	)
}

func faqIconButton(expanded bool) g.Node {
	icon := "/images/expand.svg"
	label := "Expand"
	visibility := "inline-flex group-open:hidden"
	if expanded {
		icon = "/images/collapse.svg"
		label = "Collapse"
		visibility = "hidden group-open:inline-flex"
	}
	return Span(
		Class("p-1 border border-zinc-300 dark:border-zinc-600 rounded-md shrink-0 "+visibility),
		Img(
			Class("w-5 h-5 dark:invert dark:opacity-80"),
			Src(icon),
			Alt(label),
		),
	)
}

func faqSection(id, title string, expanded bool, body ...g.Node) g.Node {
	attrs := []g.Node{
		ID(id),
		Class("group scroll-mt-24"),
	}
	if expanded {
		attrs = append(attrs, Open())
	}
	attrs = append(attrs,
		Summary(
			Class("flex items-center gap-3 text-xl font-semibold cursor-pointer list-none py-2 [&::-webkit-details-marker]:hidden"),
			faqIconButton(false),
			faqIconButton(true),
			g.Text(title),
		),
		Div(
			Class("space-y-2 pt-1 pl-10"),
			g.Group(body),
		),
	)
	return Details(attrs...)
}

func FAQPage(expandedSection string) []g.Node {
	sections := make([]g.Node, 0, len(faqOrder))
	for _, id := range faqOrder {
		data := faqSections[id]
		sections = append(sections, faqSection(
			id,
			data.title,
			expandedSection == id,
			data.body()...,
		))
	}
	return []g.Node{
		pageTitle("FAQ"),
		Div(
			Class("mt-8 space-y-2 text-zinc-700 dark:text-zinc-300"),
			g.Group(sections),
		),
	}
}
