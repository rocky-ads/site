package field

import "fmt"

func CategoryFieldsOptions(categoryID int) (CategoryChains, map[int]map[string][]string, error) {
	chains, err := GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return CategoryChains{}, nil, err
	}

	optionsMap := make(map[int]map[string][]string)
	for _, chain := range chains.Chains {
		if !IsSpecChain(chain) {
			continue
		}
		for _, cf := range chain.Fields {
			opts, err := ListSpecOptions(chain, cf.FieldName, categoryID, ParentValues{})
			if err != nil {
				return CategoryChains{}, nil, err
			}
			if len(opts) == 0 {
				continue
			}
			if optionsMap[chain.ChainID] == nil {
				optionsMap[chain.ChainID] = make(map[string][]string)
			}
			optionsMap[chain.ChainID][cf.FieldName] = opts
			break
		}
	}

	return chains, optionsMap, nil
}

func FilterFieldsOptions(categoryID int, filters AdFilters) (CategoryChains, map[int]map[string][]string, error) {
	chains, err := GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return CategoryChains{}, nil, err
	}

	optionsMap := make(map[int]map[string][]string)
	for _, chain := range chains.Chains {
		if !IsSpecChain(chain) {
			continue
		}

		parents := ParentValues{}
		firstField := ""
		for _, f := range chain.Fields {
			if !IsFilterExcluded(f.FieldName) {
				firstField = f.FieldName
				break
			}
		}

		for _, cf := range chain.Fields {
			if IsFilterExcluded(cf.FieldName) {
				continue
			}

			if cf.FieldName != firstField && !ParentsReady(chain, cf.FieldName, parents) {
				break
			}

			opts, err := ListAdFilterOptions(categoryID, chain, cf.FieldName, parents)
			if err != nil {
				return CategoryChains{}, nil, fmt.Errorf("filter options for %s: %w", cf.FieldName, err)
			}

			if len(opts) == 0 {
				if len(parents) > 0 {
					if optionsMap[chain.ChainID] == nil {
						optionsMap[chain.ChainID] = make(map[string][]string)
					}
					optionsMap[chain.ChainID][cf.FieldName] = nil
				}
				break
			}

			if optionsMap[chain.ChainID] == nil {
				optionsMap[chain.ChainID] = make(map[string][]string)
			}
			optionsMap[chain.ChainID][cf.FieldName] = opts

			selected := filters.Values(cf.FieldName)
			if len(selected) == 0 {
				break
			}
			parents[cf.FieldName] = selected
		}
	}

	return chains, optionsMap, nil
}

func ParentValuesFromQuery(values Values, fields []ChainField) ParentValues {
	parents := make(ParentValues)
	for _, f := range fields {
		if vals := values[f.FieldName]; len(vals) > 0 {
			parents[f.FieldName] = vals
		}
	}
	return parents
}
