package models

type Ad struct {
	ID          int          `json:"id,omitempty" db:"id"`
	CategoryID  int          `json:"category_id,omitempty" db:"category_id"`
	SpecValues  FieldValues  `json:"spec_values,omitempty"` // Spec field values: key is field name
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description,omitempty" db:"description"`
	Price       float64      `json:"price" db:"price"`
	CreatedAt   string       `json:"created_at" db:"created_at"`
	UserID      int          `json:"user_id" db:"user_id"`
	ImageCount  int          `json:"image_count" db:"image_count"`
	Location    LocationData `json:"location" db:"location"`
}

// LocationData represents location information (matches JSON structure)
type LocationData struct {
	City      string  `json:"city"`
	AdminArea string  `json:"admin_area"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
