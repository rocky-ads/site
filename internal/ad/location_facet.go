package ad

import (
	"strings"

	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/location"
)

const addressLoginPrompt = "(log in to see address)"

// LocationFacetKey returns the facet key used for geo lookup on this category.
func LocationFacetKey(category Category) string {
	for _, key := range category.FacetKeys {
		if key == "address" || key == "location" {
			return key
		}
	}
	return ""
}

// HasLocationFacet reports whether the category uses a location facet.
func HasLocationFacet(category Category) bool {
	return LocationFacetKey(category) != ""
}

// UsesFullAddressDisplay reports whether ads show the full address text.
func UsesFullAddressDisplay(category Category) bool {
	for _, key := range category.FacetKeys {
		if key == "address" {
			return true
		}
	}
	return false
}

// LocationTextFromFacets returns trimmed location/address text from facets.
func LocationTextFromFacets(
	category Category,
	facets map[string]facet.Value,
) string {
	key := LocationFacetKey(category)
	if key == "" {
		return ""
	}
	v, ok := facets[key]
	if !ok || v.Text == nil {
		return ""
	}
	return strings.TrimSpace(*v.Text)
}

// fullAddressText returns the stored address text for an ad.
func fullAddressText(a Ad) string {
	if v, ok := a.Facets["address"]; ok && v.Text != nil {
		if s := strings.TrimSpace(*v.Text); s != "" {
			return s
		}
	}
	return strings.TrimSpace(a.RawLocation)
}

// AdLocationDisplay returns the location line for cards and detail pages.
// Full street addresses require a logged-in viewer.
func AdLocationDisplay(a Ad, viewerUserID int) string {
	cat, err := GetCategory(a.CategoryID)
	if err != nil {
		return location.DisplayText(a.City, a.AdminArea, a.Country)
	}
	return locationDisplayForCategory(a, cat, viewerUserID)
}

func locationDisplayForCategory(a Ad, cat Category, viewerUserID int) string {
	if UsesFullAddressDisplay(cat) {
		if s := fullAddressText(a); s != "" {
			if !local.IsLoggedIn(viewerUserID) {
				return addressLoginPrompt
			}
			return s
		}
	}
	return location.DisplayText(a.City, a.AdminArea, a.Country)
}
