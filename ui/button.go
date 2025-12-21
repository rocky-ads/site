package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// buttonProps contains properties for creating a button
type buttonProps struct {
	Text     string
	Href     string
	Type     string // "button", "submit", "reset" - only used when Href is empty
	Class    string // Additional classes to append
	Attrs    []g.Node
	Children []g.Node // If provided, Text is ignored
}

// standardButton creates a standardized button element.
// If Href is provided, returns an anchor element styled as a button.
// Otherwise, returns a button element.
func standardButton(props buttonProps) g.Node {
	baseClasses := "inline-block px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"

	var classes string
	if props.Class != "" {
		classes = baseClasses + " " + props.Class
	} else {
		classes = baseClasses
	}

	allAttrs := []g.Node{Class(classes)}
	allAttrs = append(allAttrs, props.Attrs...)

	if props.Href != "" {
		// Return anchor element
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

	// Return button element
	if props.Type == "" {
		props.Type = "button"
	}
	allAttrs = append(allAttrs, Type(props.Type))
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
