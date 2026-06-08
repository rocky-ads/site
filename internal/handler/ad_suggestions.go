package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/param"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
)

func SuggestionsHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)
	category, err := ad.GetCategory(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return renderSuggestions(c, categoryID, category)
}

func EditSuggestionsHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)

	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	loc := cookie.GetLocation(c)
	a, err := ad.GetAd(userID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}
	if a.IsDeleted() {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot edit a deleted ad")
	}

	category, err := ad.GetCategory(a.CategoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return renderSuggestions(c, a.CategoryID, category)
}

func renderSuggestions(c *fiber.Ctx, categoryID int, category ad.Category) error {
	selected := parseSuggestionFormValues(c)
	suggested, _ := ad.GenerateSuggestions(suggestInputFrom(c, categoryID, category, selected))
	merged := mergeSuggestionOptions(selected, suggested)
	return render(c, uiads.SuggestionsPartial(merged))
}

func suggestInputFrom(c *fiber.Ctx, categoryID int, category ad.Category, selected []ad.Suggestion) ad.SuggestInput {
	facets := category.Facets()
	formalFacets := make(map[string]string, len(facets))
	for _, d := range facets {
		formalFacets[d.Key] = d.Label
	}
	return ad.SuggestInput{
		CategoryID:      categoryID,
		CategoryName:    category.Name,
		Title:           c.FormValue("title"),
		Description:     c.FormValue("description"),
		Location:        c.FormValue("location"),
		Facets:          formalFacets,
		AlreadySelected: selected,
	}
}

func parseSuggestionFormValues(c *fiber.Ctx) []ad.Suggestion {
	var suggestions []ad.Suggestion
	c.Context().QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) != uiads.SuggestionsFormName() {
			return
		}
		if s, ok := ad.ParseFormSuggestion(string(v)); ok {
			suggestions = append(suggestions, s)
		}
	})
	return dedupeSuggestions(suggestions)
}

func dedupeSuggestions(suggestions []ad.Suggestion) []ad.Suggestion {
	seen := make(map[string]struct{}, len(suggestions))
	out := make([]ad.Suggestion, 0, len(suggestions))
	for _, s := range suggestions {
		key := s.Key()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
}

func mergeSuggestionOptions(
	selected []ad.Suggestion,
	suggested []ad.Suggestion,
) []uiads.SuggestionOption {
	seen := make(map[string]struct{}, len(selected)+len(suggested))
	out := make([]uiads.SuggestionOption, 0, len(selected)+len(suggested))

	for _, s := range selected {
		key := s.Key()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, uiads.SuggestionOption{
			Label:    s.Label,
			Value:    s.Value,
			Selected: true,
		})
	}
	for _, s := range suggested {
		key := s.Key()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, uiads.SuggestionOption{
			Label:    s.Label,
			Value:    s.Value,
			Selected: false,
		})
	}
	return out
}

func parseAdSuggestions(c *fiber.Ctx) []ad.Suggestion {
	var suggestions []ad.Suggestion
	c.Context().PostArgs().VisitAll(func(k, v []byte) {
		if string(k) != uiads.SuggestionsFormName() {
			return
		}
		if s, ok := ad.ParseFormSuggestion(string(v)); ok {
			suggestions = append(suggestions, s)
		}
	})
	return suggestions
}
