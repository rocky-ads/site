package field

import (
	"fmt"
	"net/url"

	"github.com/rocky-ads/site/db"
	g "maragu.dev/gomponents"
)

type Field struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	CategoryID  int    `json:"category_id"`
	IsRequired  bool   `json:"is_required"`
	NextFieldID int    `json:"next_field_id"`
	PrevFieldID int    `json:"prev_field_id"`
}

// Values represents field values keyed by field name
type Values = url.Values // map[string][]string

type Fielder interface {
	FilterNode(fv Values) g.Node
	GetField() Field
}

var (
	categoryFields = make(map[int][]Fielder) // key: category ID, value: fields
)

func Init() error {
	query := `
		SELECT COALESCE(json_group_array(json_object(
			'category_id', category_id,
			'id', id,
			'name', name,
			'display_name', display_name,
			'spec_table', spec_table,
			'is_required', json(CASE WHEN is_required = 1 THEN 'true' ELSE 'false' END),
			'is_first', json(CASE WHEN is_first = 1 THEN 'true' ELSE 'false' END),
			'is_last_overall', json(CASE WHEN is_last_overall = 1 THEN 'true' ELSE 'false' END),
			'next_field_id', next_field_id,
			'prev_field_id', prev_field_id
		)), '[]')
		FROM (
			SELECT 
				c.category_id,
				f.id,
				f.name,
				f.display_name,
				COALESCE(c.spec_table, '') as spec_table,
				cf.is_required,
				CASE WHEN ROW_NUMBER() OVER (PARTITION BY c.category_id ORDER BY c.chain_index, cf.field_order) = 1 THEN 1 ELSE 0 END as is_first,
				CASE WHEN ROW_NUMBER() OVER (PARTITION BY c.category_id ORDER BY c.chain_index DESC, cf.field_order DESC) = 1 THEN 1 ELSE 0 END as is_last_overall,
				LEAD(f.id) OVER (PARTITION BY c.category_id ORDER BY c.chain_index, cf.field_order) as next_field_id,
				LAG(f.id) OVER (PARTITION BY c.category_id ORDER BY c.chain_index, cf.field_order) as prev_field_id
			FROM chain_fields cf
			JOIN chains c ON cf.chain_id = c.id
			JOIN fields f ON cf.field_id = f.id
			ORDER BY c.category_id, c.chain_index, cf.field_order
		)
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
				fd.Name, fd.DisplayName, fd.SpecTable, fd.IsRequired,
				fd.NextFieldID, fd.PrevFieldID, fd.IsFirst, fd.IsLastOverall)
			if err != nil {
				return fmt.Errorf("creating field %s for category %d: %w", fd.Name, categoryID, err)
			}
			fielders = append(fielders, fielder)
		}
		categoryFields[categoryID] = fielders
	}

	// Initialize spec fields cache
	InitSpecFields()

	return nil
}

func newField(id, categoryID int, name, displayName, specTable string, isRequired bool, nextFieldID, prevFieldID int, isFirst, isLastOverall bool) (Fielder, error) {

	field := Field{
		ID:          id,
		CategoryID:  categoryID,
		Name:        name,
		DisplayName: displayName,
		IsRequired:  isRequired,
		NextFieldID: nextFieldID,
		PrevFieldID: prevFieldID,
	}
	specField := SpecField{
		Field:         field,
		SpecTable:     specTable,
		IsFirst:       isFirst,
		IsLastOverall: isLastOverall,
	}

	switch name {
	case "location":
		return &LocationField{Field: field}, nil
	case "price":
		return &PriceField{Field: field}, nil
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

func (f Field) FilterNode(fv Values) g.Node {
	return g.Group{}
}

func (f Field) GetField() Field {
	return f
}
