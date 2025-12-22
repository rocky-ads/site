package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func userNameInput(showHelp bool) g.Node {
	return Div(
		label("Username"),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("text"),
			Name("username"),
			MinLength("3"),
			MaxLength("20"),
			g.Attr("pattern", "^[a-zA-Z][a-zA-Z0-9]{2,19}$"),
			g.Attr("placeholder", "JohnDoe"),
			Required(),
		),
		g.If(showHelp, Span(
			Class("text-xs text-gray-500 dark:text-gray-400 mt-1"),
			g.Text("3-20 characters, letters and digits only. Must start with a letter."),
		)),
	)
}

func passwordInput() g.Node {
	return Div(
		label("Password"),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("password"),
			Name("password"),
			MaxLength("32"),
		),
	)
}

func LoginForm() g.Node {
	return Form(
		Class("space-y-8"),
		hx.Post("/api/login"),
		hx.Swap("none"),
		userNameInput(false),
		passwordInput(),
		Div(
			Class("flex items-center gap-4"),
			standardButton(buttonProps{
				Type: "submit",
				Text: "Login",
			}),
			ErrorDiv(""),
		),
	)
}

func LoginPage() []g.Node {
	return []g.Node{
		pageTitle("Login"),
		LoginForm(),
	}
}

func LogoutPage() []g.Node {
	return []g.Node{
		Div(
			Class("text-center py-16"),
			H2(
				Class("text-2xl font-semibold text-gray-700 dark:text-gray-300 mb-8"),
				g.Text("You have been logged out"),
			),
			standardButton(buttonProps{
				Href: "/",
				Text: "Go Home",
			}),
		),
	}
}
