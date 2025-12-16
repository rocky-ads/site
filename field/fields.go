package field

import (
	"fmt"

	"github.com/rocky-ads/site/db"
)

var (
	categoryFields = make(map[int][]Fielder)
)

func Init() error {
	query := `
		SELECT COALESCE(json_group_array(json_object(
			'category_id', c.category_id,
			'id', f.id,
			'name', f.name,
			'display_name', f.display_name,
			'spec_table', COALESCE(c.spec_table, ''),
			'is_required', cf.is_required
		)), '[]')
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		JOIN fields f ON cf.field_id = f.id
		ORDER BY c.category_id, c.chain_index, cf.field_order
	`

	var allFields []SpecField
	if err := db.QueryJSON(&allFields, query); err != nil {
		return fmt.Errorf("loading fields: %w", err)
	}

	// Group fields by category_id
	fieldsByCategory := make(map[int][]SpecField)
	for _, fd := range allFields {
		fieldsByCategory[fd.CategoryID] = append(fieldsByCategory[fd.CategoryID], fd)
	}

	// Build categoryFields map
	for categoryID, fields := range fieldsByCategory {
		var fielders []Fielder
		for _, fd := range fields {
			fielder, err := newField(fd.ID, fd.CategoryID,
				fd.Name, fd.DisplayName, fd.SpecTable, fd.IsRequired)
			if err != nil {
				return fmt.Errorf("creating field %s for category %d: %w", fd.Name, categoryID, err)
			}
			fielders = append(fielders, fielder)
		}
		categoryFields[categoryID] = fielders
	}

	return nil
}

func newField(id, categoryID int, name, displayName, specTable string, isRequired bool) (Fielder, error) {

	field := Field{ID: id, CategoryID: categoryID, Name: name, DisplayName: displayName, IsRequired: isRequired}
	specField := SpecField{Field: field, SpecTable: specTable}

	switch name {
	case "location":
		return &LocationField{Field: field}, nil
	case "price":
		return &PriceField{Field: Field{Name: name}}, nil
	case "make":
		return &MakeField{SpecField: specField}, nil
	case "year":
		return &YearField{SpecField: specField}, nil
	case "model":
		return &ModelField{SpecField: specField}, nil
	case "engine":
		return &EngineField{SpecField: specField}, nil
	case "part_category":
		return &PartCategoryField{SpecField: specField}, nil
	case "part_subcategory":
		return &PartSubcategoryField{SpecField: specField}, nil
	default:
		return nil, fmt.Errorf("unknown field name: %s", name)
	}
}

func GetFields(categoryID int) ([]Fielder, error) {
	if fields, ok := categoryFields[categoryID]; ok {
		return fields, nil
	}
	return nil, fmt.Errorf("fields not found for category %d", categoryID)
}
