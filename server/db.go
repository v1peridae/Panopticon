package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // blank import the sqlite3 driver to register it
)

func connectToDatabase(path string) (*sql.DB, error) {
	fmt.Println("Attempting to execute database connection")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		fmt.Println("Connection secured, but pinging database failed - closing connection")
		_ = db.Close()
		return nil, err
	}
	fmt.Println("Connected successfully to SQLite database")
	return db, nil
}
