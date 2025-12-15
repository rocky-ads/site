package models

// Category represents an ad category
type Category struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	SeedAdFile string  `json:"seed_ad_file"`
	ImageFile  string  `json:"image_file"`
	Chains     []Chain `json:"chains"`
}

// Chain represents a chain within a category
type Chain struct {
	SpecTable string  `json:"spec_table,omitempty"`
	ChainFile string  `json:"chain_file,omitempty"`
	Fields    []Field `json:"fields"`
}
