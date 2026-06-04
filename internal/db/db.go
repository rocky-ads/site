package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

func Query(query string, args ...any) (*sql.Rows, error) {
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

func QueryRow(query string, args ...any) *sql.Row {
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

func Exec(query string, args ...any) (sql.Result, error) {
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

func Select(dest any, query string, args ...any) error {
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
	return db.Ping()
}

// For SQLite, JSON functions return TEXT, so we scan as string
func QueryJSON(dst any, query string, args ...any) error {
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

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// HashString creates a SHA256 hash of a value for database lookups
func HashString(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
