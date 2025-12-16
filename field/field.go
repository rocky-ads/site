package field

import (
	"net/url"

	g "maragu.dev/gomponents"
)

type Field struct {
	ID          int
	Name        string
	DisplayName string
	IsRequired  bool
	CategoryID  int
}

// Values represents field values keyed by field name
type Values = url.Values // map[string][]string

// ValuesByIDs represents field values keyed by field ID
type ValuesByIDs map[int][]string

type Fielder interface {
	FilterNode(fv Values) g.Node
}
