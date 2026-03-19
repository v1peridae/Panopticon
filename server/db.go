package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

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

// create tables if they don't exist
func initSchema(db *sql.DB) {
	schema, _ := os.ReadFile("server/database/setup.sql")
	for _, stmt := range strings.Split(string(schema), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			db.Exec(stmt)
		}
	}
}

// gets pending notifs from db
func GetPendingNotifs(db *sql.DB) ([]Notification, error) {
	rows, err := db.Query("SELECT id, header, description, status FROM notifications WHERE status = 'pending' ORDER BY created_at DESC")
	if err != nil { return nil, err}
	defer rows.Close()
	var notifs []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Header, &n.Description, &n.Status); err != nil {return nil, err}
		notifs = append(notifs, n)
	}
	return notifs, rows.Err()
}

// get completed notifs from db
func GetCompletedNotifs(db *sql.DB) ([]Notification, error) {
	rows, err := db.Query("SELECT id, header, description, status FROM notifications WHERE status = 'complete' ORDER BY created_at DESC")
	if err != nil { return nil, err}
	defer rows.Close()
	var notifs []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Header, &n.Description, &n.Status); err != nil {return nil, err}
		notifs = append(notifs, n)
	}
	return notifs, rows.Err()
}

// create a new notif in the db
func CreateNotif(db *sql.DB, header, description string) (int64, error) {
	result, err:= db.Exec("INSERT INTO notifications (header, description) VALUES (?, ?)", header, description)
	if err != nil { return 0, err }
	return result.LastInsertId()
}

//updates
func UpdateNotif(db *sql.DB, id int, header, description, status string) error {
	_, err:= db.Exec("UPDATE notifications SET header = ?, description = ?, status = ? WHERE id = ?", header, description, status, id)
	return err
}

//deletes
func DeleteNotif(db *sql.DB, id int) error {
	_, err:= db.Exec("DELETE FROM notifications WHERE id = ?", id)
	return err
}