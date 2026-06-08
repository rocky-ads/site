package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func userNameInput(showHelp bool, autocomplete string) g.Node {
	return Div(
		label("Username"),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("text"),
			Name("username"),
			MinLength("3"),
			MaxLength("20"),
			g.Attr("pattern", "^[a-zA-Z][a-zA-Z0-9]{2,19}$"),
			g.Attr("autocomplete", autocomplete),
			Required(),
		),
		g.If(showHelp, Span(
			Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1"),
			g.Text("3-20 characters, letters and digits only. Must start with a letter."),
		)),
	)
}

func passwordInput(autocomplete string) g.Node {
	return Div(
		label("Password"),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("password"),
			Name("password"),
			MaxLength("32"),
			g.Attr("autocomplete", autocomplete),
			Required(),
		),
	)
}

func LoginForm() g.Node {
	return Form(
		Class("space-y-8 mt-8"),
		hx.Post("/api/login"),
		hx.Swap("none"),
		userNameInput(false, "username"),
		passwordInput("current-password"),
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
