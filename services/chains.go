package services

import (
	"sort"

	"github.com/rocky-ads/site/db"
)

var (
	categoryChainsCache = make(map[int][]ChainInfo)
)

// ChainFieldInfo represents a field in a chain with database metadata for API responses
type ChainFieldInfo struct {
	Name        string `json:"Name"`
	DisplayName string `json:"DisplayName"`
	Order       int    `json:"Order"`
	NextInChain int    `json:"NextInChain"`
}

// ChainInfo represents chain information for API responses
type ChainInfo struct {
	ChainIndex int              `json:"ChainIndex"`
	Fields     []ChainFieldInfo `json:"Fields"`
}

func GetCategoryChains(categoryID int) ([]ChainInfo, error) {
	if cached, ok := categoryChainsCache[categoryID]; ok {
		return cached, nil
	}

	type fieldData struct {
		ChainIndex  int    `json:"chain_index"`
		FieldOrder  int    `json:"field_order"`
		NextInChain int    `json:"next_in_chain"`
		FieldName   string `json:"name"`
		DisplayName string `json:"display_name"`
	}
	var allFields []fieldData
	query := `
		SELECT COALESCE(json_group_array(json_object(
			'chain_index', c.chain_index,
			'field_order', cf.field_order,
			'next_in_chain', cf.next_in_chain,
			'name', f.name,
			'display_name', f.display_name
		)), '[]')
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		JOIN fields f ON cf.field_id = f.id
		WHERE c.category_id = ? AND c.spec_table IS NOT NULL AND c.spec_table != ''
		ORDER BY c.chain_index, cf.field_order
	`
	if err := db.QueryJSON(&allFields, query, categoryID); err != nil {
		return nil, err
	}

	chainFieldCounts := make(map[int]int)
	for _, fd := range allFields {
		chainFieldCounts[fd.ChainIndex]++
	}

	var chainIndices []int
	for ci := range chainFieldCounts {
		chainIndices = append(chainIndices, ci)
	}
	sort.Ints(chainIndices)

	chainOffsets := make(map[int]int)
	offset := 0
	for _, actualIdx := range chainIndices {
		chainOffsets[actualIdx] = offset
		offset += chainFieldCounts[actualIdx]
	}

	chains := make(map[int][]ChainFieldInfo)
	for _, fd := range allFields {
		absoluteOrder := chainOffsets[fd.ChainIndex] + fd.FieldOrder + 1
		absoluteNextInChain := 0
		if fd.NextInChain > 0 {
			absoluteNextInChain = chainOffsets[fd.ChainIndex] + fd.NextInChain + 1
		}

		chains[fd.ChainIndex] = append(chains[fd.ChainIndex], ChainFieldInfo{
			Name:        fd.FieldName,
			DisplayName: fd.DisplayName,
			Order:       absoluteOrder,
			NextInChain: absoluteNextInChain,
		})
	}

	var result []ChainInfo
	for normalizedIdx, actualIdx := range chainIndices {
		result = append(result, ChainInfo{
			ChainIndex: normalizedIdx,
			Fields:     chains[actualIdx],
		})
	}

	categoryChainsCache[categoryID] = result
	return result, nil
}
