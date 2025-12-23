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
			Type("tel"),
			Name("phone"),
			MinLength("10"),
			MaxLength("20"),
			g.Attr("placeholder", "+12025550123 or 202-555-0123"),
			g.Attr("pattern", "^\\+?[\\d\\s\\-\\(\\)\\.]{10,20}$"),
			g.Attr("autocomplete", "tel"),
			Required(),
		),
		Span(
			Class("text-xs text-gray-500 dark:text-gray-400 mt-1"),
			g.Text("Enter your phone in international format (e.g. +12025550123) or US/Canada format (e.g. 503-523-8780)."),
		),
	)
}

func offers() g.Node {
	return Div(
		Class("space-y-3"),
		checkbox("offers", "true", "I agree to receive informational text messages", false, false, Required()),
		Div(
			Class("text-xs text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-800 p-3 rounded border dark:border-gray-700 space-y-2"),
			P(
				g.Text("By providing your phone number you agree to receive informational text messages from "+config.ServerName+"."),
			),
			P(
				g.Text("Message frequency will vary. Msg & data rates may apply. Reply HELP for help or STOP to cancel. We only use your phone for essential communications and verification."),
			),
		),
	)
}

func RegisterForm() g.Node {
	return Form(
		Class("space-y-8"),
		ID("registerForm"),
		hx.Post("/api/register/step1"),
		hx.Swap("none"),
		userNameInput(true, "username"),
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

func hiddenUserName(username string) g.Node {
	return Input(
		Type("hidden"),
		Name("username"),
		Value(username),
	)
}

func hiddenPhone(phone string) g.Node {
	return Input(
		Type("hidden"),
		Name("phone"),
		Value(phone),
	)
}

func verificationCodeInput() g.Node {
	return Div(
		Class("text-center mb-6"),
		P(
			Class("text-gray-600"),
			g.Text("We've sent a verification code to your phone number. "+
				"Please enter the code below to complete your registration."),
		),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("text"),
			Name("code"),
			g.Attr("autocomplete", "one-time-code"),
			g.Attr("inputmode", "numeric"),
			g.Attr("pattern", "[0-9]*"),
		),
	)
}

func passwordInput2() g.Node {
	return Div(
		label("Confirm Password"),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("password"),
			Name("password2"),
			MaxLength("32"),
			g.Attr("autocomplete", "off"),
		),
	)
}

func terms() g.Node {
	return Div(
		Class("flex items-center gap-2"),
		checkbox("terms", "accepted", "I accept the ", false, false, Required()),
		A(
			Href("/terms"),
			Class("text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline"),
			g.Text("Terms of Service"),
		),
		g.Text(" & "),
		A(
			Href("/privacy"),
			Class("text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline"),
			g.Text("Privacy Policy"),
		),
		g.Text("."),
	)
}

func RegisterVerify(username, phoneE64 string) g.Node {
	return Form(
		Class("space-y-8"),
		ID("registerForm"),
		hx.Post("/api/register/step2"),
		hx.Swap("none"),
		hx.SwapOOB("true"),
		hiddenUserName(username),
		hiddenPhone(phoneE64),
		verificationCodeInput(),
		passwordInput("new-password"),
		passwordInput2(),
		terms(),
		Div(
			Class("flex items-center gap-4"),
			standardButton(buttonProps{
				Type: "submit",
				Text: "Complete Registration",
			}),
			ErrorDiv(""),
		),
	)
}
