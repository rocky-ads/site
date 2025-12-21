package ui

import (
	"github.com/rocky-ads/site/config"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func phoneInput() g.Node {
	return Div(
		label("Phone Number"),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("text"),
			Name("phone"),
			MinLength("10"),
			MaxLength("12"),
			g.Attr("placeholder", "+12025550123 or 202-555-0123"),
			g.Attr("pattern", "(\\+[0-9]{10,15}|[0-9]{3}-[0-9]{3}-[0-9]{4})"),
			Required(),
		),
		Span(
			Class("text-xs text-gray-500 mt-1"),
			g.Text("Enter your phone in international format (e.g. +12025550123) or US/Canada format (e.g. 503-523-8780)."),
		),
	)
}

func offers() g.Node {
	return Div(
		Div(
			Class("space-y-3"),
			checkbox("offers", "true", "I agree to receive informational text messages", false, false, Required()),
		),
		Div(
			Class("text-xs text-gray-600 bg-gray-50 p-3 rounded border"),
			g.Text("By providing your phone number you agree to receive informational text messages from "+config.ServerName+". Message frequency will vary. Msg & data rates may apply. Reply HELP for help or STOP to cancel. We only use your phone for essential communications and verification."),
		),
	)
}

func RegisterForm() g.Node {
	return Form(
		Class("space-y-8"),
		hx.Post("/api/register/step1"),
		hx.Swap("none"),
		userNameInput(),
		phoneInput(),
		offers(),
		Div(
			Class("flex items-center gap-4"),
			standardButton(buttonProps{
				Type: "submit",
				Text: "Register",
			}),
			ErrorDiv(""),
		),
	)
}

func RegisterPage() []g.Node {
	return []g.Node{
		pageTitle("Register"),
		RegisterForm(),
	}
}
