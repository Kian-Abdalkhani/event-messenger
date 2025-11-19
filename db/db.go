package db

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"event-messenger.com/config"
	_ "github.com/mattn/go-sqlite3"
)

// Global database variable
var DB *sql.DB

// initDB initializes the database and creates the table if it doesn't exist
func InitDB() {

	// Create a directory if it does not already exist
	dir := filepath.Dir(config.App.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("Could not create database directory: %v", err))
	}

	var err error
	DB, err = sql.Open("sqlite3", config.App.DBPath)

	if err != nil {
		panic("could not connect to database")
	}

	// maxiumum amount of DB connections and idle connections
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	createTables()

}

func createTables() {
	// Create events table (eventID, eventName)
	createEventsTable := `CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        slug TEXT UNIQUE NOT NULL,
        description TEXT,
        event_date DATETIME NOT NULL,
        active BOOLEAN DEFAULT 1,
        coordinator TEXT,
        coordinator_contact TEXT,
        recipient_name TEXT,
        recipient_email TEXT,
		email_sent BOOLEAN DEFAULT 0,
		email_sent_at DATETIME,
        website_link TEXT,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );`

	// Create submissions table
	createSubmissionsTable := `CREATE TABLE IF NOT EXISTS submissions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        message TEXT NOT NULL,
        filename TEXT,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
    );`

	// Create indexes for performance
	createIndexes := `
        CREATE INDEX IF NOT EXISTS idx_events_slug ON events(slug);
        CREATE INDEX IF NOT EXISTS idx_events_active ON events(active);
        CREATE INDEX IF NOT EXISTS idx_submissions_event_id ON submissions(event_id);
        CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON submissions(created_at);
    `

	_, err := DB.Exec(createEventsTable)
	if err != nil {
		panic("could not create Events table")
	}

	_, err = DB.Exec(createSubmissionsTable)
	if err != nil {
		panic("could not create Submissions table")
	}

	_, err = DB.Exec(createIndexes)
	if err != nil {
		log.Printf("Warning: could not create indexes: %v", err)
	}

	slog.Debug("Database initialized successfully")
}
