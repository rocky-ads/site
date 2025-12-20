package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func userNameInput() g.Node {
	return Div(
		Label(
			Class("block text-base font-medium mb-1"),
			g.Text("Username"),
		),
		Input(
			Class("w-full p-2 border rounded-md"),
			Type("text"),
			Name("username"),
			MaxLength("16"),
		),
	)
}

func passwordInput() g.Node {
	return Div(
		Label(
			Class("block text-base font-medium mb-1"),
			g.Text("Password"),
		),
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
		Class("space-y-6"),
		hx.Post("/api/login"),
		hx.Swap("none"),
		userNameInput(),
		passwordInput(),
		Div(
			Class("flex items-center gap-4"),
			StandardButton(ButtonProps{
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
				g.Text("You've been logged out."),
			),
			StandardButton(ButtonProps{
				Href: "/",
				Text: "Go Home",
			}),
		),
	}
}
