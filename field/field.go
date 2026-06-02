package field

import "net/url"

type Field struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	InputType   string `json:"input_type"`
	CategoryID  int    `json:"category_id"`
	IsRequired  bool   `json:"is_required"`
}

type Values = url.Values
