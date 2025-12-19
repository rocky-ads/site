package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sqlx.DB
	once sync.Once
)

func Init(dbPath string) error {
	var err error
	once.Do(func() {
		db, err = sqlx.Open("sqlite3", dbPath+"?_foreign_keys=1")
		if err != nil {
			return
		}

		if err = db.Ping(); err != nil {
			return
		}
	})
	return err
}

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

func formatArgs(args []interface{}) []interface{} {
	formatted := make([]interface{}, len(args))
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

			preview := make([]interface{}, previewCount)
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

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	/*
		fmt.Println("")
		fmt.Println("=== SQL Query ===")
		fmt.Println(prettyPrintSQL(query))
		fmt.Println("Args:", formatArgs(args))
		fmt.Println("================")
		fmt.Println("")
	*/
	return db.Query(query, args...)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	/*
		fmt.Println("")
		fmt.Println("=== SQL QueryRow ===")
		fmt.Println(prettyPrintSQL(query))
		fmt.Println("Args:", formatArgs(args))
		fmt.Println("===================")
		fmt.Println("")
	*/
	return db.QueryRow(query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	/*
		fmt.Println("")
		fmt.Println("=== SQL Exec ===")
		fmt.Println(prettyPrintSQL(query))
		fmt.Println("Args:", args)
		fmt.Println("===============")
		fmt.Println("")
	*/
	return db.Exec(query, args...)
}

func Begin() (*sql.Tx, error) {
	return db.Begin()
}

func Select(dest interface{}, query string, args ...interface{}) error {
	/*
		fmt.Println("")
		fmt.Println("=== SQL Select ===")
		fmt.Println(prettyPrintSQL(query))
		fmt.Println("Args:", formatArgs(args))
		fmt.Println("=================")
		fmt.Println("")
	*/
	return db.Select(dest, query, args...)
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}

// For SQLite, JSON functions return TEXT, so we scan as string
func QueryJSON(dst interface{}, query string, args ...interface{}) error {
	var jsonResult string
	/*
		fmt.Println("")
		fmt.Println("=== SQL QueryJSON ===")
		fmt.Println(prettyPrintSQL(query))
		fmt.Println("Args:", formatArgs(args))
		fmt.Println("===============")
		fmt.Println("")
	*/
	err := db.QueryRow(query, args...).Scan(&jsonResult)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonResult), dst)
}

// DB returns the underlying sqlx.DB instance
func DB() *sqlx.DB {
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
