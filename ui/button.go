package ui

import (
	g "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

// ButtonProps contains properties for creating a button
type ButtonProps struct {
	Text     string
	Href     string
	Type     string // "button", "submit", "reset" - only used when Href is empty
	Class    string // Additional classes to append
	Attrs    []g.Node
	Children []g.Node // If provided, Text is ignored
}

// StandardButton creates a standardized button element.
// If Href is provided, returns an anchor element styled as a button.
// Otherwise, returns a button element.
func StandardButton(props ButtonProps) g.Node {
	baseClasses := "inline-block px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"

	var classes string
	if props.Class != "" {
		classes = baseClasses + " " + props.Class
	} else {
		classes = baseClasses
	}

	allAttrs := []g.Node{html.Class(classes)}
	allAttrs = append(allAttrs, props.Attrs...)

	if props.Href != "" {
		// Return anchor element
		allAttrs = append(allAttrs, html.Href(props.Href))
		var content g.Node
		if len(props.Children) > 0 {
			content = g.Group(props.Children)
		} else {
			content = g.Text(props.Text)
		}
		allNodes := append(allAttrs, content)
		return html.A(allNodes...)
	}

	// Return button element
	if props.Type == "" {
		props.Type = "button"
	}
	allAttrs = append(allAttrs, html.Type(props.Type))
	var content g.Node
	if len(props.Children) > 0 {
		content = g.Group(props.Children)
	} else {
		content = g.Text(props.Text)
	}
	allNodes := append(allAttrs, content)
	return html.Button(allNodes...)
}
