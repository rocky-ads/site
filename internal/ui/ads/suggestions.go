package ads

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

const (
	adSuggestionsID     = "ad-suggestions"
	suggestionFormSep   = "\x1f"
	suggestionsFormName = "suggestion"
)

func descriptionWithSuggestionsBox() g.Node {
	return Div(
		Class("w-full border rounded-md overflow-hidden"),
		descriptionInput(),
		Div(
			Class("flex items-start gap-2 p-2 border-t"),
			Div(
				Class("shrink-0 flex flex-col gap-2"),
				suggestionsButton(),
				suggestionsIndicator(),
			),
			Div(
				ID(adSuggestionsID),
				Class("flex flex-wrap gap-2 flex-1 min-w-0"),
			),
		),
	)
}

func suggestionsButton() g.Node {
	return Button(
		Type("button"),
		Class("inline-flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors text-sm whitespace-nowrap"),
		hx.Get("/auth/ad/new/suggestions"),
		hx.Target("#"+adSuggestionsID),
		hx.Swap("innerHTML"),
		hx.Include("#new-ad-form"),
		hx.Indicator("#"+adSuggestionsIndicatorID),
		Img(
			Src("/images/wand.svg"),
			Alt(""),
			Class("w-4 h-4 brightness-0 invert"),
		),
		g.Text("Suggestions"),
	)
}

const adSuggestionsIndicatorID = "ad-suggestions-indicator"

func suggestionsIndicator() g.Node {
	return Div(
		ID(adSuggestionsIndicatorID),
		Class("htmx-indicator flex items-center gap-2 text-blue-600 text-sm whitespace-nowrap"),
		Div(
			Class("w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"),
		),
		g.Text("Thinking..."),
	)
}

// SuggestionsPartial renders toggleable suggestion pills for HTMX swap.
func SuggestionsPartial(opts []SuggestionOption) g.Node {
	if len(opts) == 0 {
		return Span(
			Class("text-sm text-zinc-500 dark:text-zinc-400"),
			g.Text("Add more content to Title and Description and try again"),
		)
	}
	nodes := make([]g.Node, len(opts))
	for i, o := range opts {
		nodes[i] = suggestionPill(o)
	}
	return g.Group(nodes)
}

func suggestionPill(o SuggestionOption) g.Node {
	display := suggestionDisplay(o.Label, o.Value)
	attrs := []g.Node{
		Type("checkbox"),
		Name(suggestionsFormName),
		Value(encodeSuggestionFormValue(o.Label, o.Value)),
		Class("sr-only"),
	}
	if o.Selected {
		attrs = append(attrs, g.Attr("checked", "checked"))
	}
	return Label(
		Class("inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-zinc-300 dark:border-zinc-600 cursor-pointer text-sm has-[:checked]:bg-blue-100 has-[:checked]:border-blue-500 dark:has-[:checked]:bg-blue-900 dark:has-[:checked]:border-blue-400"),
		Input(attrs...),
		Span(g.Text(display)),
	)
}

func encodeSuggestionFormValue(label, value string) string {
	return label + suggestionFormSep + value
}

// SuggestionsFormName is the checkbox name for selected suggestion pills.
func SuggestionsFormName() string {
	return suggestionsFormName
}

func suggestionDisplay(label, value string) string {
	if value == "yes" {
		return label
	}
	return label + ": " + value
}
