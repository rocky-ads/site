// Package facet defines the category-specific ad attributes (price, mileage,
// hours, year, condition, ...) in one typed registry. Each facet's kind,
// unit handling, allowed values, validation, and display formatting live in a
// single Def so that adding a facet is a one-line registry entry rather than a
// schema/SQL/UI change.
package facet

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/location"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type Kind int

const (
	Int   Kind = iota // numeric, range-filterable
	Money             // amount + currency (Value.Text), FREE when amount == 0
	Enum              // fixed string set, exact-match filter
)

// FormWidget controls the new-ad form control for a facet.
type FormWidget int

const (
	FormNumber FormWidget = iota // plain number input (+ unit select when Units set)
	FormSelect                   // <select>; options from Enum (Kind Enum) or SelectMin..SelectMax (Kind Int)
	FormRadio                    // exclusive choices for Kind Enum
	FormMoney                    // amount + currency + FREE
)

// FilterWidget controls the search filter panel for a facet.
type FilterWidget int

const (
	FilterRange      FilterWidget = iota // min/max inputs
	FilterExact                          // single-select (Enum dropdown)
	FilterCheckboxes                     // multi-select (Enum checkboxes)
)

// FormDefaults holds per-request values for new-ad form widgets.
type FormDefaults struct {
	Currency string // default selected currency for money facets
	Unit     string // preferred unit for facets with Units (e.g. mi/km)
}

// Def is the definition of a single facet.
type Def struct {
	Key        string
	Label      string
	Kind       Kind
	Form       FormWidget
	Filter     FilterWidget
	Units      []string // Int-only: allowed units stored in Value.Text; empty = no unit
	Suffix     string   // Int-only: constant suffix when Units is empty (e.g. "hrs")
	Compact    bool     // Int-only: render large values as e.g. "45K" on cards
	Enum       []string // enum options for FormSelect and FormRadio
	SelectMin  int      // FormSelect lower bound when Kind == Int (inclusive)
	SelectMax  int      // FormSelect upper bound when Kind == Int; 0 = current calendar year + 1
	Filterable bool     // show in the search filter panel
	Required   bool     // required on the new-ad form
}

// ValidUnit reports whether unit is allowed for this facet.
func (d Def) ValidUnit(unit string) bool {
	unit = strings.TrimSpace(unit)
	for _, u := range d.Units {
		if u == unit {
			return true
		}
	}
	return false
}

// SupportedCurrencies returns currency dropdown options for money facets.
func (d Def) SupportedCurrencies() []string {
	if d.Form != FormMoney {
		return nil
	}
	return currency.Supported
}

// FormDefaultCurrency returns the selected currency for a money facet.
func (d Def) FormDefaultCurrency(defaults FormDefaults) string {
	if d.Form != FormMoney {
		return ""
	}
	code := currency.Normalize(defaults.Currency)
	if currency.IsSupported(code) {
		return code
	}
	return currency.Default
}

// FormDefaultUnit returns the selected unit for a facet with Units.
func (d Def) FormDefaultUnit(defaults FormDefaults) string {
	if len(d.Units) == 0 {
		return ""
	}
	return d.NormalizeUnit(defaults.Unit)
}

// NormalizeUnit returns unit if allowed, otherwise the first allowed unit.
func (d Def) NormalizeUnit(unit string) string {
	unit = strings.TrimSpace(unit)
	if d.ValidUnit(unit) {
		return unit
	}
	if len(d.Units) > 0 {
		return d.Units[0]
	}
	return unit
}

// SelectBounds returns inclusive min/max for FormSelect on Kind Int.
func (d Def) SelectBounds() (min, max int) {
	min = d.SelectMin
	if d.SelectMax != 0 {
		max = d.SelectMax
	} else {
		max = time.Now().Year() + 1
	}
	if max < min {
		max = min
	}
	return min, max
}

// FormOptions returns dropdown choices for FormSelect.
func (d Def) FormOptions() []string {
	switch {
	case d.Kind == Enum:
		return d.Enum
	case d.Form == FormSelect && d.Kind == Int:
		min, max := d.SelectBounds()
		opts := make([]string, 0, max-min+1)
		for v := max; v >= min; v-- {
			opts = append(opts, strconv.Itoa(v))
		}
		return opts
	default:
		return nil
	}
}

// Value is a stored facet value for one ad. Num holds Int/Money amounts; Text
// holds the currency (Money), unit (Int with Units), or selection (Enum).
type Value struct {
	Num  *int
	Text *string
}

// Filter is a search constraint. Min/Max apply to Int/Money; Value to a single
// Enum; Values to multiple Enum selections (OR within the facet).
type Filter struct {
	Min    *int     `json:"min,omitempty"`
	Max    *int     `json:"max,omitempty"`
	Value  *string  `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

func (v Value) Present() bool {
	return v.Num != nil || v.Text != nil
}

func (f Filter) Active() bool {
	return f.Min != nil || f.Max != nil || f.Value != nil || len(f.Values) > 0
}

var defs = []Def{
	{
		Key:        "price",
		Label:      "Price",
		Kind:       Money,
		Form:       FormMoney,
		Filter:     FilterRange,
		Filterable: true,
		Required:   true,
	},
	{
		Key:        "mileage",
		Label:      "Mileage",
		Kind:       Int,
		Form:       FormNumber,
		Filter:     FilterRange,
		Units:      []string{location.UnitMiles, location.UnitKm},
		Compact:    true,
		Filterable: true,
		Required:   true,
	},
	{
		Key:        "hours",
		Label:      "Hours",
		Kind:       Int,
		Form:       FormNumber,
		Filter:     FilterRange,
		Suffix:     "hrs",
		Filterable: true,
		Required:   true,
	},
	{
		Key:        "year",
		Label:      "Year",
		Kind:       Int,
		Form:       FormSelect,
		Filter:     FilterRange,
		SelectMin:  1900,
		Filterable: true,
		Required:   true,
	},
	{
		Key:    "condition",
		Label:  "Condition",
		Kind:   Enum,
		Form:   FormRadio,
		Filter: FilterCheckboxes,
		Enum: []string{
			"New",
			"Used - Like new",
			"Used - Good",
			"Used - Fair",
			"Used - Poor",
		},
		Filterable: true,
		Required:   false,
	},
	{
		Key:    "title_status",
		Label:  "Title Status",
		Kind:   Enum,
		Form:   FormRadio,
		Filter: FilterCheckboxes,
		Enum: []string{
			"Clean",
			"Salvage",
			"Rebuilt",
			"Parts Only",
		},
		Filterable: true,
		Required:   false,
	},
}

var byKey = func() map[string]Def {
	m := make(map[string]Def, len(defs))
	for _, d := range defs {
		m[d.Key] = d
	}
	return m
}()

// Get returns the definition for a facet key.
func Get(key string) (Def, bool) {
	d, ok := byKey[key]
	return d, ok
}

// All returns every registered facet definition.
func All() []Def {
	return defs
}

// Validate checks a value against the definition. Optional facets may be absent;
// Required facets must be present and pass kind-specific rules.
func (d Def) Validate(v Value) error {
	switch d.Kind {
	case Int:
		if v.Num == nil {
			if d.Required {
				return fmt.Errorf("%s is required", d.Label)
			}
			return nil
		}
		if *v.Num < 0 {
			return fmt.Errorf("%s must be zero or greater", d.Label)
		}
		if len(d.Units) > 0 && (v.Text == nil || !d.ValidUnit(*v.Text)) {
			return fmt.Errorf("%s requires a valid unit", d.Label)
		}
		if d.Form == FormSelect && d.Kind == Int {
			min, max := d.SelectBounds()
			if *v.Num < min || *v.Num > max {
				return fmt.Errorf("%s must be between %d and %d", d.Label, min, max)
			}
		}
	case Money:
		if v.Num == nil {
			if d.Required {
				return fmt.Errorf("%s is required", d.Label)
			}
			return nil
		}
		if *v.Num < 0 {
			return fmt.Errorf("%s must be zero or greater", d.Label)
		}
		if v.Text == nil || !currency.IsSupported(*v.Text) {
			return fmt.Errorf("%s requires a valid currency", d.Label)
		}
	case Enum:
		if v.Text == nil || strings.TrimSpace(*v.Text) == "" {
			if d.Required {
				return fmt.Errorf("%s is required", d.Label)
			}
			return nil
		}
		for _, e := range d.Enum {
			if e == *v.Text {
				return nil
			}
		}
		return fmt.Errorf("invalid %s", d.Label)
	}
	return nil
}

// FormatCompact renders a value for dense contexts (listing cards).
func (d Def) FormatCompact(v Value) string {
	return d.format(v, d.Compact)
}

// FormatFull renders a value with full precision (ad detail page).
func (d Def) FormatFull(v Value) string {
	return d.format(v, false)
}

func (d Def) format(v Value, compact bool) string {
	switch d.Kind {
	case Money:
		if v.Num == nil {
			return ""
		}
		code := ""
		if v.Text != nil {
			code = *v.Text
		}
		return currency.Format(*v.Num, code)
	case Int:
		if v.Num == nil {
			return ""
		}
		var s string
		if d.Form == FormSelect && d.Kind == Int {
			s = strconv.Itoa(*v.Num)
		} else {
			s = formatCount(*v.Num, compact)
		}
		if len(d.Units) > 0 && v.Text != nil {
			return s + " " + *v.Text
		}
		if d.Suffix != "" {
			return s + " " + d.Suffix
		}
		return s
	case Enum:
		if v.Text == nil {
			return ""
		}
		return *v.Text
	}
	return ""
}

func formatCount(n int, compact bool) string {
	if compact && n >= 1000 {
		return strconv.Itoa(n/1000) + "K"
	}
	p := message.NewPrinter(language.English)
	return p.Sprint(number.Decimal(int64(n), number.Scale(0)))
}
