package models

import (
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"event-messenger.com/db"
)

type Submission struct {
	ID        int
	EventID   int
	Name      string
	Message   string
	Filename  string
	CreatedAt time.Time
}

func (s *Submission) Save() error {
	insertSQL := `INSERT INTO submissions (event_id, name, message, filename, created_at) VALUES (?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(insertSQL, s.EventID, s.Name, s.Message, s.Filename, time.Now().UTC())
	if err != nil {
		return err
	}

	submissionId, err := result.LastInsertId()

	s.ID = int(submissionId)

	return err
}

func GetAllSubmissions() ([]Submission, error) {
	query := `SELECT id, event_id, name, message, filename, created_at FROM submissions ORDER BY created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying submissions: %v", err)
	}

	defer rows.Close()

	var submissions []Submission
	for rows.Next() {
		var s Submission
		err := rows.Scan(&s.ID, &s.EventID, &s.Name, &s.Message, &s.Filename, &s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		submissions = append(submissions, s)
	}

	return submissions, nil
}

func GetSubmissionsByEventSlug(slug string) ([]Submission, error) {
	query := `SELECT s.id, s.event_id, s.name, s.message, s.filename, s.created_at 
              FROM submissions s
              JOIN events e ON s.event_id = e.id
              WHERE e.slug = ?
              ORDER BY s.created_at DESC`

	rows, err := db.DB.Query(query, slug)
	if err != nil {
		return nil, fmt.Errorf("error querying submissions: %v", err)
	}
	defer rows.Close()

	var submissions []Submission
	for rows.Next() {
		var s Submission
		err := rows.Scan(&s.ID, &s.EventID, &s.Name, &s.Message, &s.Filename, &s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		submissions = append(submissions, s)
	}

	return submissions, nil
}
