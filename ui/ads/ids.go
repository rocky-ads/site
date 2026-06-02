package ads

import "fmt"

func fieldContainerID(chainID int, fieldName string) string {
	return fmt.Sprintf("field-%d-%s", chainID, fieldName)
}

func chainContainerID(chainID int) string {
	return fmt.Sprintf("chain-%d", chainID)
}

func chainNextID(chainID int, afterField string) string {
	return fmt.Sprintf("chain-%d-next-%s", chainID, afterField)
}

func filterFieldContainerID(chainID int, fieldName string) string {
	return fmt.Sprintf("filter-field-%d-%s", chainID, fieldName)
}

func filterChainContainerID(chainID int) string {
	return fmt.Sprintf("filter-chain-%d", chainID)
}

func filterChainNextID(chainID int, afterField string) string {
	return fmt.Sprintf("filter-chain-%d-next-%s", chainID, afterField)
}
