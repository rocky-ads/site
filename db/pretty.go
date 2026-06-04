package db

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func prettyPrintSQL(query string) string {
	query = strings.TrimSpace(query)

	// Order matters: longer keywords first to avoid partial matches
	keywords := []string{
		"LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL OUTER JOIN",
		"GROUP BY", "ORDER BY", "INSERT INTO", "DELETE FROM",
		"SELECT", "FROM", "JOIN", "WHERE", "HAVING", "UNION",
		"UPDATE", "SET", "VALUES", "ON", "AND", "OR",
	}

	formatted := query
	for _, keyword := range keywords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
		formatted = re.ReplaceAllStringFunc(formatted, func(match string) string {
			return "\n" + match
		})
	}

	lines := strings.Split(formatted, "\n")
	var result []string
	indentLevel := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(line), "WHERE") ||
			strings.HasPrefix(strings.ToUpper(line), "ORDER BY") ||
			strings.HasPrefix(strings.ToUpper(line), "GROUP BY") ||
			strings.HasPrefix(strings.ToUpper(line), "HAVING") {
			indentLevel = 0
		}

		indented := strings.Repeat("  ", indentLevel) + line
		result = append(result, indented)

		if strings.HasPrefix(strings.ToUpper(line), "SELECT") ||
			strings.HasPrefix(strings.ToUpper(line), "FROM") ||
			strings.HasPrefix(strings.ToUpper(line), "JOIN") ||
			strings.HasPrefix(strings.ToUpper(line), "WHERE") {
			indentLevel++
		}

		if i < len(lines)-1 {
			nextLine := strings.ToUpper(strings.TrimSpace(lines[i+1]))
			if strings.HasPrefix(nextLine, "WHERE") ||
				strings.HasPrefix(nextLine, "ORDER BY") ||
				strings.HasPrefix(nextLine, "GROUP BY") ||
				strings.HasPrefix(nextLine, "HAVING") {
				indentLevel = 0
			}
		}
	}

	return strings.Join(result, "\n")
}

func formatArgs(args []any) []any {
	formatted := make([]any, len(args))
	for i, arg := range args {
		if arg == nil {
			formatted[i] = arg
			continue
		}

		val := reflect.ValueOf(arg)

		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			length := val.Len()
			if length == 0 {
				formatted[i] = "[]"
				continue
			}

			previewCount := 2
			if length < previewCount {
				previewCount = length
			}

			preview := make([]any, previewCount)
			for j := range previewCount {
				preview[j] = val.Index(j).Interface()
			}

			if length > previewCount {
				formatted[i] = fmt.Sprintf("[%v, %v, ...] (%d elements)", preview[0], preview[1], length)
			} else if length == 2 {
				formatted[i] = fmt.Sprintf("[%v, %v]", preview[0], preview[1])
			} else {
				formatted[i] = fmt.Sprintf("[%v]", preview[0])
			}
			continue
		}

		formatted[i] = arg
	}
	return formatted
}
