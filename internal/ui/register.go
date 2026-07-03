package ui

import (
	"github.com/rocky-ads/site/internal/config"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func phoneInput() g.Node {
	return Div(
		Div(
			Class("flex items-baseline justify-between mb-1"),
			Label(
				Class("text-base font-medium"),
				For("phone"),
				g.Text("Phone Number"),
			),
			A(
				Href("/faq/phone-number"),
				Class("text-sm text-blue-600 dark:text-blue-400 hover:underline shrink-0"),
				g.Text("Why do you need my phone number?"),
			),
		),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("tel"),
			Name("phone"),
			ID("phone"),
			MinLength("10"),
			MaxLength("20"),
			g.Attr("placeholder", "+12025550123 or 202-555-0123"),
			g.Attr("pattern", "^\\+?[\\d\\s\\-\\(\\)\\.]{10,20}$"),
			g.Attr("autocomplete", "tel"),
			Required(),
		),
		Span(
			Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1"),
			g.Text("Enter your phone in international format (e.g. +12025550123) or US/Canada format (e.g. 503-523-8780)."),
		),
	)
}

func offers() g.Node {
	return Div(
		Class("space-y-3"),
		checkbox("offers", "true", "I agree to receive informational text messages", false, false, Required()),
		Div(
			Class("text-xs text-zinc-600 dark:text-zinc-300 bg-zinc-50 dark:bg-zinc-800 p-3 rounded border dark:border-zinc-700 space-y-2"),
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
		Class("space-y-8 mt-8"),
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
		Class("text-center mb-6 space-y-6 py-4"),
		P(
			Class("text-base leading-relaxed"),
			g.Text("We've sent a verification code to your phone number. "+
				"Please enter the code below to continue."),
		),
		Label(
			Class("block text-base font-medium mb-2"),
			g.Attr("for", "verification-code"),
			g.Text("Verification Code"),
		),
		Input(
			ID("verification-code"),
			Class("max-w-xs w-full p-4 border-2 border-zinc-300 dark:border-zinc-600 rounded-md text-center text-2xl font-mono tracking-widest focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none dark:bg-zinc-800 dark:text-zinc-200 mx-auto block my-2"),
			Type("text"),
			Name("code"),
			g.Attr("autocomplete", "one-time-code"),
			g.Attr("inputmode", "numeric"),
			g.Attr("pattern", "[0-9]*"),
			g.Attr("maxlength", "6"),
			g.Attr("placeholder", "000000"),
		),
	)
}

func passwordInput2() g.Node {
	return labeledPasswordField("Confirm Password", "password2", "off")
}

func terms() g.Node {
	linkClass := "text-blue-600 dark:text-blue-400 " +
		"hover:text-blue-800 dark:hover:text-blue-300 underline"
	return Label(
		Class("flex items-start gap-2"),
		For("terms-accepted"),
		Input(
			Type("checkbox"),
			Name("terms"),
			Value("accepted"),
			ID("terms-accepted"),
			Class("mt-1 shrink-0"),
			Required(),
		),
		Span(
			g.Text("I accept the "),
			A(
				Href("/terms"),
				Class(linkClass),
				g.Text("Terms of Service"),
			),
			g.Text(" & "),
			A(
				Href("/privacy"),
				Class(linkClass),
				g.Text("Privacy Policy"),
			),
			g.Text("."),
		),
	)
}

func RegisterVerify(username, phoneE64 string) g.Node {
	return Form(
		Class("space-y-12"),
		ID("registerForm"),
		hx.Post("/api/register/step2"),
		hx.Swap("none"),
		hx.SwapOOB("true"),
		hiddenUserName(username),
		hiddenPhone(phoneE64),
		verificationCodeInput(),
		Div(
			Class("flex flex-col items-center gap-6"),
			standardButton(buttonProps{
				Type: "submit",
				Text: "Verify Code",
			}),
			ErrorDiv(""),
		),
	)
}

func RegisterPassword(username, phoneE64 string) g.Node {
	return Form(
		Class("space-y-8"),
		ID("registerForm"),
		hx.Post("/api/register/step3"),
		hx.Swap("none"),
		hx.SwapOOB("true"),
		hiddenUserName(username),
		hiddenPhone(phoneE64),
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
