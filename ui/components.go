package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// buttonProps contains properties for creating a button
type buttonProps struct {
	Text     string
	Href     string
	Type     string // "button", "submit", "reset" - only used when Href is empty
	Class    string // Additional classes to append
	Disabled bool   // If true, button will be disabled (for links, renders as disabled button instead)
	Attrs    []g.Node
	Children []g.Node // If provided, Text is ignored
}

// standardButton creates a standardized button element.
// If Href is provided, returns an anchor element styled as a button.
// Otherwise, returns a button element.
// If Disabled is true and Href is set, renders a disabled button instead of a link.
func standardButton(props buttonProps) g.Node {
	baseClasses := "inline-block px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"

	var classes string
	if props.Class != "" {
		classes = baseClasses + " " + props.Class
	} else {
		classes = baseClasses
	}

	// Add disabled styling if disabled
	if props.Disabled {
		classes += " opacity-50 cursor-not-allowed"
	}

	allAttrs := []g.Node{Class(classes)}
	allAttrs = append(allAttrs, props.Attrs...)

	if props.Href != "" && !props.Disabled {
		// Return anchor element (only if not disabled)
		allAttrs = append(allAttrs, Href(props.Href))
		var content g.Node
		if len(props.Children) > 0 {
			content = g.Group(props.Children)
		} else {
			content = g.Text(props.Text)
		}
		allNodes := append(allAttrs, content)
		return A(allNodes...)
	}

	// Return button element (or disabled button if Href was set but Disabled is true)
	if props.Type == "" {
		props.Type = "button"
	}
	allAttrs = append(allAttrs, Type(props.Type))
	if props.Disabled {
		allAttrs = append(allAttrs, Disabled())
	}
	var content g.Node
	if len(props.Children) > 0 {
		content = g.Group(props.Children)
	} else {
		content = g.Text(props.Text)
	}
	allNodes := append(allAttrs, content)
	return Button(allNodes...)
}

func checkbox(name string, value string, label string, checked bool, disabled bool, attrs ...g.Node) g.Node {
	inputAttrs := []g.Node{
		Type("checkbox"),
		Name(name),
		Value(value),
		ID(name + "-" + value),
	}

	if checked {
		inputAttrs = append(inputAttrs, Checked())
	}
	if disabled {
		inputAttrs = append(inputAttrs, Disabled())
		inputAttrs = append(inputAttrs, g.Attr("class", "opacity-50 cursor-not-allowed"))
	}

	inputAttrs = append(inputAttrs, attrs...)

	labelNode := Label(
		For(name+"-"+value),
		g.If(disabled, Class("text-gray-400")),
		g.Text(label),
	)

	return Div(
		Class("flex items-center space-x-2"),
		Input(inputAttrs...),
		labelNode,
	)
}

func label(text string) g.Node {
	return Label(
		Class("block text-base font-medium mb-1"),
		g.Text(text),
	)
}

// inputText creates a standardized text input element with dark mode support
func inputText(name, placeholder string, isRequired bool, attrs ...g.Node) g.Node {
	inputAttrs := []g.Node{
		Type("text"),
		Name(name),
		Class("w-full p-2 border rounded-md"),
		Placeholder(placeholder),
		g.If(isRequired, Required()),
	}
	inputAttrs = append(inputAttrs, attrs...)
	return Input(inputAttrs...)
}

func textArea(name, placeholder string, isRequired bool, attrs ...g.Node) g.Node {
	inputAttrs := []g.Node{
		Type("textarea"),
		Name(name),
		Class("w-full p-2 border rounded-md"),
		Placeholder(placeholder),
		g.If(isRequired, Required()),
	}
	inputAttrs = append(inputAttrs, attrs...)
	return Textarea(inputAttrs...)
}

// pageTitle creates a standardized H1 page title element
func pageTitle(text string) g.Node {
	return H1(
		Class("text-3xl font-bold"),
		g.Text(text),
	)
}

// RemoveModal returns swap-oob delete elements to remove a modal and its backdrop
func RemoveModal(name string) []g.Node {
	return []g.Node{
		Div(
			ID(name+"-modal-backdrop"),
			hx.SwapOOB("delete"),
		),
		Div(
			ID(name+"-modal"),
			hx.SwapOOB("delete"),
		),
	}
}
