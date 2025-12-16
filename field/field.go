package field

import (
	"net/url"

	g "maragu.dev/gomponents"
)

type Field struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	IsRequired  bool   `json:"is_required"`
	CategoryID  int    `json:"category_id"`
}

// Values represents field values keyed by field name
type Values = url.Values // map[string][]string

// ValuesByIDs represents field values keyed by field ID
type ValuesByIDs map[int][]string

type Fielder interface {
	FilterNode(fv Values) g.Node
}

func (f Field) FilterNode(fv Values) g.Node {
	return g.Group{}
}
