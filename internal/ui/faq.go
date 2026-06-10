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
	"phone-number",
	"sms-notifications",
}

var faqSections = map[string]faqSectionData{
	"phone-number": {
		title: "Why do you need my phone number?",
		body: func() []g.Node {
			return []g.Node{
				faqParagraph(g.Textf(
					"%s is a local marketplace where buyers and "+
						"sellers message each other directly. We ask "+
						"for your phone number during registration "+
						"for two reasons: to verify your account and "+
						"to reach you when something important "+
						"happens.",
					config.ServerName,
				)),
				faqParagraph(g.Text(
					"During signup we send a one-time verification " +
						"code to confirm the number is yours. This " +
						"helps keep fake accounts and spam off the " +
						"site.",
				)),
				faqParagraph(
					g.Text(
						"After you register, we may text you about "+
							"messages on your ads, replies to inquiries "+
							"you sent, and other account activity you "+
							"would reasonably want to know about. You "+
							"can turn text notifications on or off "+
							"anytime in ",
					),
					faqLink("/auth/user/settings", "Settings"),
					g.Text("."),
				),
				faqParagraph(g.Text(
					"We do not sell your phone number, use it for " +
						"marketing unrelated to your account, or " +
						"share it with other users. It is used only " +
						"for verification and essential " +
						"communications about your account.",
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
