package ads

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// LocationInput renders the shared location text field (search and new-ad forms).
func LocationInput(id, name, value, placeholder string) g.Node {
	attrs := []g.Node{
		Type("text"),
		Name(name),
		ID(id),
		Class("w-full p-2 border rounded-md"),
		g.Attr("placeholder", placeholder),
	}
	if value != "" {
		attrs = append(attrs, Value(value))
	}
	return Input(attrs...)
}
