package ad

import (
	"fmt"
	"strings"
)

func suggestionsUserPrompt(in SuggestInput) string {
	var b strings.Builder
	b.WriteString("AD COPY:\n")
	fmt.Fprintf(&b, "Title: %s\n", in.Title)
	fmt.Fprintf(&b, "Description: %s\n", in.Description)
	fmt.Fprintf(&b, "Location: %s\n", in.Location)

	if len(in.AlreadySelected) > 0 {
		b.WriteString("\nALREADY SELECTED (do not suggest again):\n")
		for _, s := range in.AlreadySelected {
			fmt.Fprintf(&b, "  %s\n", s.PromptDisplay())
		}
	}

	remaining := maxSuggestions - len(in.AlreadySelected)
	if remaining > 0 {
		fmt.Fprintf(&b, "\nReturn at most %d new suggestions.\n", remaining)
	}

	return b.String()
}
