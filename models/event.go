package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"event-messenger.com/db"
	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	ID                 int       `db:"id"`
	Name               string    `db:"name"`
	Slug               string    `db:"slug"`
	Description        string    `db:"description"`
	EventDate          time.Time `db:"event_date"`
	Active             bool      `db:"active"`
	Coordinator        string    `db:"coordinator"`
	CoordinatorContact string    `db:"coordinator_contact"`
	// New fields for email notification feature
	RecipientName  string       `db:"recipient_name"`
	RecipientEmail string       `db:"recipient_email"`
	EmailSent      bool         `db:"email_sent"`
	EmailSentAt    sql.NullTime `db:"email_sent_at"`
	WebsiteLink    string       `db:"website_link"` // Link to send recipient
	CreatedAt      time.Time    `db:"created_at"`
}

type EventOption func(*Event)

// NewEvent creates a new Event with required fields and optional configuration
func NewEvent(name, slug string, eventDate time.Time, opts ...EventOption) *Event {
	event := &Event{
		Name:      name,
		Slug:      slug,
		EventDate: eventDate,
		Active:    true,
		CreatedAt: time.Now(),
	}

	// Apply optional configurations
	for _, opt := range opts {
		opt(event)
	}

	return event
}

// Option functions
func WithDescription(desc string) EventOption {
	return func(e *Event) {
		e.Description = desc
	}
}

func WithCoordinator(name, contact string) EventOption {
	return func(e *Event) {
		e.Coordinator = name
		e.CoordinatorContact = contact
	}
}

func WithRecipient(name, email string) EventOption {
	return func(e *Event) {
		e.RecipientName = name
		e.RecipientEmail = email
	}
}

func WithWebsiteLink(link string) EventOption {
	return func(e *Event) {
		e.WebsiteLink = link
	}
}

func WithActive(active bool) EventOption {
	return func(e *Event) {
		e.Active = active
	}
}

func (e *Event) ArchiveEvent() error {
	// Only archive if email was sent successfully
	if !e.EmailSent {
		return fmt.Errorf("cannot archive event: email not sent")
	}

	query := `
    UPDATE events
    SET active = FALSE
    WHERE id = ?
    `

	_, err := db.DB.Exec(query, e.ID)
	if err != nil {
		return fmt.Errorf("error archiving event: %v", err)
	}

	return nil
}

// GetEventsReadyForDeletion returns events that have been inactive for a grace period
func GetEventsReadyForDeletion(graceDays int) ([]Event, error) {
	query := `
    SELECT id, name, slug, description, event_date, active,
           coordinator, coordinator_contact, 
           recipient_name, recipient_email, email_sent, email_sent_at,
           website_link, created_at
    FROM events
    WHERE email_sent = TRUE 
      AND active = FALSE
      AND email_sent_at IS NOT NULL
      AND DATE(email_sent_at) <= DATE(?, ?)
    `

	cutoffDate := time.Now().UTC().AddDate(0, 0, -graceDays)
	rows, err := db.DB.Query(query, cutoffDate)
	if err != nil {
		return nil, fmt.Errorf("error querying events for deletion: %v", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Description,
			&e.EventDate, &e.Active,
			&e.Coordinator, &e.CoordinatorContact,
			&e.RecipientName, &e.RecipientEmail, &e.EmailSent,
			&e.EmailSentAt, &e.WebsiteLink, &e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		events = append(events, e)
	}

	return events, nil
}

func (e *Event) DeleteEvent() error {
	// Safety check: only delete if email was sent AND sufficient time has passed
	if !e.EmailSent {
		return fmt.Errorf("cannot delete event: email not sent")
	}

	if e.EmailSentAt.Valid {
		daysSinceSent := time.Since(e.EmailSentAt.Time).Hours() / 24
		if daysSinceSent < 30 { // 30-day grace period
			return fmt.Errorf("cannot delete event: grace period not elapsed (%.0f days remaining)", 30-daysSinceSent)
		}
	}

	query := `
	DELETE FROM events
	WHERE id = ?
	`

	_, err := db.DB.Exec(query, e.ID)
	if err != nil {
		return fmt.Errorf("error deleting event: %v", err)
	}

	log.Printf("Event deleted: %s (ID: %d)", e.Name, e.ID)
	return nil
}

func (e *Event) SaveEvent() error {
	// converts datetimes to UTC datetime for consistency
	eventDateUTC := e.EventDate.UTC()
	createdAtUTC := e.CreatedAt.UTC()

	insertSQL := `INSERT INTO events (
        name, slug, description, event_date, active, 
        coordinator, coordinator_contact, 
        recipient_name, recipient_email, website_link, 
        created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(
		insertSQL,
		e.Name, e.Slug, e.Description, eventDateUTC, e.Active,
		e.Coordinator, e.CoordinatorContact,
		e.RecipientName, e.RecipientEmail, e.WebsiteLink,
		createdAtUTC,
	)
	if err != nil {
		return err
	}

	eventId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	e.ID = int(eventId)
	return err
}

// GetSubmissionCount returns the number of submissions for this event
func (e *Event) GetSubmissionCount() (int, error) {
	query := `SELECT COUNT(*) FROM submissions WHERE event_id = ?`

	var count int
	err := db.DB.QueryRow(query, e.ID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting submissions: %v", err)
	}

	return count, nil
}

func (e *Event) MarkEmailSent() error {
	query := `
	UPDATE events
	SET email_sent = TRUE, active = FALSE, email_sent_at = ?
	WHERE id = ?;
	`

	_, err := db.DB.Exec(query, time.Now().UTC(), e.ID)
	if err != nil {
		return fmt.Errorf("error updating event: %v", err)
	}

	return nil
}

func GetEventsForToday() ([]Event, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	startOfDayUTC := startOfDay.UTC()
	endOfDayUTC := endOfDay.UTC()

	query := `
	SELECT * FROM events
	WHERE event_date >= ? AND event_date < ?
	`

	rows, err := db.DB.Query(query, startOfDayUTC, endOfDayUTC)
	if err != nil {
		return nil, fmt.Errorf("error querying event previews: %v", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err = rows.Scan(&event.ID, &event.Name, &event.Slug, &event.Description,
			&event.EventDate, &event.Active,
			&event.Coordinator, &event.CoordinatorContact,
			&event.RecipientName, &event.RecipientEmail, &event.EmailSent,
			&event.EmailSentAt, &event.WebsiteLink, &event.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		events = append(events, event)
	}
	return events, nil

}

type EventPreview struct {
	ID              int       `db:"id"`
	Name            string    `db:"name"`
	Slug            string    `db:"slug"`
	Description     string    `db:"description"`
	EventDate       time.Time `db:"event_date"`
	RecipientName   string    `db:"recipient_name"`
	SubmissionCount int       `db:"submission_count"`
}

// GetActiveEventPreviews returns lightweight event data for list/preview displays
func GetActiveEventPreviews() ([]EventPreview, error) {
	query := `SELECT 
        e.id, e.name, e.slug, e.description, e.event_date, 
        e.recipient_name,
        COUNT(s.id) as submission_count
    FROM events e
    LEFT JOIN submissions s ON e.id = s.event_id
    WHERE e.active = true
    GROUP BY e.id
    ORDER BY e.event_date DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying event previews: %v", err)
	}
	defer rows.Close()

	var events []EventPreview
	for rows.Next() {
		var ep EventPreview
		err := rows.Scan(
			&ep.ID, &ep.Name, &ep.Slug, &ep.Description,
			&ep.EventDate, &ep.RecipientName,
			&ep.SubmissionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		events = append(events, ep)
	}

	return events, nil
}

func GetAllActiveEvents() ([]Event, error) {
	query := `SELECT id, name, slug, description, event_date, active,
              coordinator, coordinator_contact, 
              recipient_name, recipient_email, website_link, created_at 
              FROM events WHERE active = true ORDER BY event_date DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying events: %v", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Description,
			&e.EventDate, &e.Active,
			&e.Coordinator, &e.CoordinatorContact,
			&e.RecipientName, &e.RecipientEmail, &e.WebsiteLink,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		events = append(events, e)
	}

	return events, nil
}

type EventWithCount struct {
	Event
	SubmissionCount int
}

// GetAllActiveEventsWithCounts returns all active events with their submission counts
func GetAllActiveEventsWithCounts() ([]EventWithCount, error) {
	query := `SELECT 
        e.id, e.name, e.slug, e.description, e.event_date, e.active,
        e.coordinator, e.coordinator_contact, 
        e.recipient_name, e.recipient_email, e.website_link, e.created_at,
        COUNT(s.id) as submission_count
    FROM events e
    LEFT JOIN submissions s ON e.id = s.event_id
    WHERE e.active = true
    GROUP BY e.id
    ORDER BY e.event_date DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying events with counts: %v", err)
	}
	defer rows.Close()

	var events []EventWithCount
	for rows.Next() {
		var ewc EventWithCount
		err := rows.Scan(
			&ewc.ID, &ewc.Name, &ewc.Slug, &ewc.Description,
			&ewc.EventDate, &ewc.Active,
			&ewc.Coordinator, &ewc.CoordinatorContact,
			&ewc.RecipientName, &ewc.RecipientEmail, &ewc.WebsiteLink,
			&ewc.CreatedAt,
			&ewc.SubmissionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		events = append(events, ewc)
	}

	return events, nil
}

func GetEventBySlug(slug string) (*Event, error) {
	query := `SELECT id, name, slug, description, event_date, active, 
              coordinator, coordinator_contact, 
              recipient_name, recipient_email, website_link, created_at 
              FROM events WHERE slug = ? AND active = true`

	var e Event
	err := db.DB.QueryRow(query, slug).Scan(
		&e.ID, &e.Name, &e.Slug, &e.Description,
		&e.EventDate, &e.Active,
		&e.Coordinator, &e.CoordinatorContact,
		&e.RecipientName, &e.RecipientEmail, &e.WebsiteLink,
		&e.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("event not found: %v", err)
	}

	return &e, nil
}

func (e *Event) Update() error {
	eventDateUTC := e.EventDate.UTC()

	updateSQL := `UPDATE events SET 
        name = ?, description = ?, event_date = ?, active = ?,
        coordinator = ?, coordinator_contact = ?,
        recipient_name = ?, recipient_email = ?, website_link = ?
        WHERE id = ?`

	_, err := db.DB.Exec(
		updateSQL,
		e.Name, e.Description, eventDateUTC, e.Active,
		e.Coordinator, e.CoordinatorContact,
		e.RecipientName, e.RecipientEmail, e.WebsiteLink,
		e.ID,
	)
	return err
}
