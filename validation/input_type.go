package validation

import (
	"net/url"
	"strings"
)

const (
	InputText        = "text"
	InputNumber      = "number"
	InputSelect      = "select"
	InputSelectMulti = "select_multi"
)

// InputSpec is a parsed input_type value (base type plus validation query params).
type InputSpec struct {
	Type   string
	Params url.Values
}

type patternDef struct {
	regex string
	msg   string
}

var patternDefs = map[string]patternDef{
	"ascii": {
		regex: `[\x20-\x7E]+`,
		msg:   "Please enter printable ASCII characters only",
	},
	"ascii-multiline": {
		regex: `[\x20-\x7E\n\r]+`,
		msg:   "Please enter printable ASCII characters only (line breaks allowed)",
	},
	"nonneg-int": {
		regex: `^(0|[1-9][0-9]*)$`,
		msg:   "Please enter a whole number (0 or greater)",
	},
}

func ParseInputType(raw string) InputSpec {
	spec := InputSpec{Type: InputText, Params: make(url.Values)}
	if raw == "" {
		return spec
	}

	var base string
	switch {
	case strings.Contains(raw, "?"):
		u, err := url.Parse("/" + raw)
		if err != nil {
			return spec
		}
		base = strings.TrimPrefix(u.Path, "/")
		spec.Params = u.Query()
	case strings.Contains(raw, "="):
		params, err := url.ParseQuery(raw)
		if err != nil {
			return spec
		}
		spec.Params = params
	default:
		spec.Type = raw
		return spec
	}

	switch {
	case base != "":
		spec.Type = base
	case spec.Params.Get("type") != "":
		spec.Type = spec.Params.Get("type")
		spec.Params.Del("type")
	default:
		spec.Type = InputText
	}

	return spec
}

func (s InputSpec) resolvePattern() (patternDef, bool) {
	key := s.Param("pattern")
	if key == "" {
		return patternDef{}, false
	}
	if def, ok := patternDefs[key]; ok {
		return def, true
	}
	return patternDef{regex: key}, true
}

func (s InputSpec) HTMLPattern() string {
	if def, ok := s.resolvePattern(); ok {
		return def.regex
	}
	return ""
}

func AnchoredPattern(regex string) string {
	if regex == "" {
		return ""
	}
	if strings.HasPrefix(regex, "^") || strings.HasSuffix(regex, "$") {
		return regex
	}
	return "^(?:" + regex + ")$"
}

func (s InputSpec) PatternMessage() string {
	if def, ok := s.resolvePattern(); ok {
		return def.msg
	}
	return ""
}

func (s InputSpec) Param(key string) string {
	if v := s.Params[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}
