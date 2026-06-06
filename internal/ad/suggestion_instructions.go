package ad

import (
	"fmt"
	"sort"
	"strings"
)

// suggestionsSharedHeader is identical for every category so it can be cached
// as the leading part of the per-category system prompt (see x-grok-conv-id).
const suggestionsSharedHeader = `You help sellers fill in optional details missing
from their classified ad. Return JSON only: an array of
{"label":"...","value":"..."}. No markdown, no prose. At most 12 items.

Goal: suggest clickable choices for specs a buyer would ask about that are
NOT covered by the formal form fields and NOT already clear from the ad.

Rules:
- Do NOT restate facts already obvious from the title or description
- Do NOT suggest anything in the "Already selected" list
- If "Already selected" includes ANY value for an attribute (label), do NOT
  suggest that attribute again — no alternate values for the same label
- Focus on missing specs; infer from year, make, model, and category
- When an attribute has multiple plausible values and is NOT in "Already
  selected", emit a separate entry for EACH realistic option (same label,
  different values)
- For binary features (present or not: navigation, backup camera, heated
  seats, towing package, etc.), emit ONE entry with value "yes":
  {"label":"heated seats","value":"yes"}. Never emit yes/no pairs.
- Only include options that could apply to this make/model/year; omit
  choices that never existed for this item
- label = short attribute name; value = one specific choice when needed
- when value is non-empty, it MUST differ from label`

// categorySuggestionInstructions holds the category-specific guidance block,
// keyed by category name. Keep each value constant so the assembled system
// prompt stays byte-identical across requests for the same category.
var categorySuggestionInstructions = map[string]string{
	"Cars & Trucks": `Buyers of cars and trucks care about drivetrain and
usability specs that are not formal fields, such as:
- transmission: manual | automatic
- fuel: gas | diesel | electric | hybrid
- drivetrain: FWD | RWD | AWD | 4WD
- cylinders, engine size, exterior color, doors, seating
Tailor options to the specific make, model, and year.`,

	"Car & Truck Parts": `Buyers of car and truck parts care about fitment
and condition details, such as:
- fits make/model/year or generation
- part type (OEM | aftermarket)
- side (left | right | front | rear), position
- material/finish, whether the part is new, used, or rebuilt
Tailor options to the specific part described.`,

	"Motorcycles": `Buyers of motorcycles care about:
- engine size (cc), engine type (2-stroke | 4-stroke)
- cylinders, cooling (air | liquid)
- type (sport | cruiser | touring | dirt | adventure)
- transmission, aftermarket parts, ownership/title extras
Tailor options to the specific make, model, and year.`,

	"Motorcycle Parts": `Buyers of motorcycle parts care about fitment and
condition, such as:
- fits make/model/year
- part type (OEM | aftermarket)
- position/side, material/finish
- new, used, or rebuilt
Tailor options to the specific part described.`,

	"Bicycles": `Buyers of bicycles care about:
- type (road | mountain | gravel | hybrid | BMX | e-bike)
- frame size, frame material (aluminum | carbon | steel | titanium)
- wheel size, drivetrain/speeds, brake type (disc | rim)
- suspension (none | front | full)
Tailor options to the specific make, model, and year.`,

	"Bicycle Parts": `Buyers of bicycle parts care about compatibility and
condition, such as:
- component group/standard, wheel or tire size
- material, mounting standard
- new or used
Tailor options to the specific part described.`,

	"Agricultural Equipment": `Buyers of agricultural equipment care about:
- equipment type, drive (2WD | 4WD)
- horsepower (PTO/engine), fuel (diesel | gas)
- hours of use, attachments/implements included
- cab vs open station
Tailor options to the specific make, model, and year.`,

	"Agricultural Equipment Parts": `Buyers of agricultural equipment parts
care about fitment and condition, such as:
- fits make/model/series
- part type (OEM | aftermarket)
- new, used, or rebuilt
Tailor options to the specific part described.`,
}

// suggestionsSystemPrompt builds the constant-per-category system prompt:
// shared header + formal fields + category guidance.
func suggestionsSystemPrompt(categoryName string, facets map[string]string) string {
	guidance := categorySuggestionInstructions[categoryName]
	if guidance == "" {
		guidance = `Suggest specs a buyer in this category would ask about
that are not already on the form. Tailor options to the specific item.`
	}

	fields := formalFacetLabelList(facets)

	return fmt.Sprintf(`%s
- label at most %d characters; value at most %d characters
- keep labels and values short; abbreviate when needed (e.g. "transmission" ->
  "trans", "automatic" -> "auto", "exterior color" -> "ext color")

Category: %s
Formal form fields already collected (never suggest these): %s

Category guidance:
%s`, suggestionsSharedHeader, maxSuggestionLabelLen, maxSuggestionValueLen, categoryName, fields, guidance)
}

func formalFacetLabelList(facets map[string]string) string {
	if len(facets) == 0 {
		return "(none)"
	}
	keys := make([]string, 0, len(facets))
	for k := range facets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	labels := make([]string, len(keys))
	for i, k := range keys {
		labels[i] = facets[k]
	}
	return strings.Join(labels, ", ")
}

func suggestionsConvID(categoryID int) string {
	return fmt.Sprintf("ad-suggestions-cat-%d", categoryID)
}
