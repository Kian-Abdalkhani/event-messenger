# Event Messenger - AI Agent Instructions

## Project Overview

A Go web application for collecting congratulatory messages/photos for events (graduations, birthdays, etc.). Built as a LAN-first service with automated email notification system. Users submit messages and images to event-specific pages; on the event date, recipients receive an email with all submissions.

**Core Flow**: Create event → Get shareable URL → Collect submissions → Automated email on event date → Auto-cleanup after 30 days

## Architecture

### Package Structure

- `main.go` - Server initialization with `init()` pattern: loads config → initializes SQLite DB → starts schedulers → registers routes
- `config/` - Environment-based configuration (dev uses `.env`, production uses system env vars via `GO_ENV`)
- `db/` - SQLite database initialization with global `db.DB` connection (max 10 open, 5 idle)
- `models/` - Data layer (Events, Submissions) with embedded SQL queries
- `handlers/` - HTTP handlers split by concern (events, submissions, home, rendering)
- `routes/` - Custom router with slug-based routing: `/events/{slug}/{action}`
- `scheduler/` - **PRODUCTION**: Two background schedulers (notification at 8AM, cleanup weekly at 2AM)
  - `notification.go` - Sends emails with embedded images on event date
  - `cleanup.go` - Deletes events 30 days after email sent
  - `scheduler.go` - Cron orchestration
- `utils/` - Slug generation, email sending (SMTP), URL helpers
- `templates/` - HTML templates rendered server-side via `text/template`

### Database Schema

**SQLite** at `./data/app.db` (configurable via `DB_PATH`):

- `events` table: event metadata with slug-based URLs, recipient info for planned email feature
- `submissions` table: messages + image filenames (files stored in `./data/uploads/`)
- Indexes on: `events.slug`, `events.active`, `submissions.event_id`, `submissions.created_at`

### Routing Pattern

Custom URL router in `routes/routes.go`:

```go
/                              → List all active events (home.go)
/events/create                 → Event creation form (events.go)
/events/create/submit          → POST: Create new event (events.go)
/events/{slug}                 → Submission form for event (submissions.go)
/events/{slug}/submit          → POST: Submit message/photo (submissions.go)
/uploads/*                     → Static file serving for uploaded images
/static/*                      → Static assets (CSS, JS)
```

The `eventRouteHandler` parses paths using `parseEventPath()` to extract slug and action from `/events/{slug}/{action}` patterns. Routes are registered via `http.ServeMux` (no third-party router).

## Key Development Patterns

### Functional Options Pattern

Event creation uses functional options (see `models/event.go`):

```go
event := models.NewEvent(name, slug, date,
    models.WithDescription(desc),
    models.WithCoordinator(name, contact),
)
```

Use this pattern for optional fields to maintain clean constructors.

### Global Database Connection

Database connection is initialized once in `init()` and accessed via `db.DB` throughout. All models use this global connection for queries.

### Template Rendering

All handlers use `renderTemplate(w, path, data)` from `handlers/render.go`. Templates are parsed on each request (no caching) for development simplicity.

### File Upload Validation & Processing

Submissions require image uploads with strict validation and automatic optimization:

- Max 10MB file size (enforced via `ParseMultipartForm`)
- MIME type detection via first 512 bytes (`http.DetectContentType`)
- Allowed types: JPEG, PNG, GIF, WebP
- **Auto-resize**: Images wider than 800px are resized maintaining aspect ratio (uses `golang.org/x/image/draw` with CatmullRom scaling)
- **Format conversion**: All images converted to JPEG at 85% quality to reduce email size
- Files saved as `{unix_timestamp}_{original_filename}.jpg` in `./data/uploads/`
- **Critical for email**: Images embedded as base64 data URIs in notification emails (see `scheduler/notification.go`)

### Slug Generation

Event URLs use auto-generated slugs from event names via `utils.GenerateSlug()`:

- Lowercase, ASCII-only, hyphen-separated
- Removes special chars, collapses multiple hyphens
- **Must be unique** in database (enforced by UNIQUE constraint)

## Environment Configuration

### Required Environment Variables

```bash
BASE_URL=http://localhost:8080  # Used for generating shareable event links
WEB_PORT=8080
DB_PATH=./data/app.db
GO_ENV=production              # Controls .env loading (development|production)
```

### Optional (For Future Email Feature)

```bash
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email
SMTP_PASSWORD=your_password
SMTP_FROM_EMAIL=noreply@example.com
```

## Running the Application

### Development

```bash
# Create .env file with variables above
go run main.go
# Server starts on http://localhost:8080
```

### Production (Docker)

```bash
docker compose up -d
# Persists data via volume: ./data:/app/data
# Timezone set to America/Los_Angeles in compose file
```

### Email System

**SMTP Configuration** (required in `.env` or system environment):

```bash
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email
SMTP_PASSWORD=your_app_password  # Use app-specific password for Gmail
SMTP_FROM_EMAIL=noreply@example.com
```

**Email Sending Logic** (`utils/email.go`):

- `SendEmailNotification(toEmail, subject, htmlContent)` - Sends via SMTP with TLS
- Automatically triggered by `scheduler/notification.go` on event dates
- Emails capped at 150 submissions to stay under SMTP limits (25MB)
- Images embedded as base64 data URIs (encoded in `scheduler/notification.go`)

## Important Constraints

1. **No Authentication**: Designed for trusted LAN environments. Admin routes like `/events/create` are unprotected.

2. **File Upload Required**: Submissions must include an image (enforced in `handlers/submissions.go`). Returns 400 if missing.

3. **Event Date Validation**: Events must have future dates (`handlers/events.go:48`).

4. **Slug Uniqueness**: Database enforces unique slugs. Manual slug editing not supported—duplicates will fail on insert.

5. **Timezone Handling**: Docker deployment uses `TZ=America/Los_Angeles`. Event dates stored as DATETIME without explicit timezone conversion.

## Production Features: Schedulers

### Email Notification Scheduler

**Daily Scheduler** (`scheduler/scheduler.go` + `scheduler/notification.go`):

- Runs at 8AM system time (configurable via `StartScheduler(hourToRun)`)
- Also executes immediately on app startup for testing
- Flow on each run:
  1. `GetEventsForToday()` - Queries events where `event_date` matches today
  2. For each event with `EmailSent = false`:
     - `GetSubmissionsByEventSlug()` - Retrieves all submissions
     - Caps at 150 submissions (SMTP 25MB limit protection)
     - Encodes images as base64 data URIs
     - Renders `templates/email_notification.html` with submission data
     - Sends email via `utils.SendEmailNotification()`
     - Calls `event.MarkEmailSent()` - Sets `EmailSent=true`, `Active=false`, records timestamp
  3. Logs all successes/failures

**Critical Implementation Details:**

- Email size is logged (warns if >15MB)
- Image encoding handles JPEG, PNG, GIF, WebP via MIME type detection
- Template data includes: `EventName`, `RecipientName`, `EventDate`, `Submissions[]`, `TotalCount`, `CoordinatorName`

### Cleanup Scheduler

**Weekly Scheduler** (`scheduler/cleanup.go`):

- Runs every 7 days at 2AM system time
- Deletes events where `EmailSent=true` AND 30+ days have passed since `EmailSentAt`
- Flow:
  1. `GetEventsReadyForDeletion(graceDays)` - Queries old archived events
  2. For each event: `event.DeleteEvent()` - Includes safety checks
  3. Cascade deletes submissions via `ON DELETE CASCADE` foreign key
- **Safety**: `DeleteEvent()` enforces 30-day grace period before allowing deletion

## Common Tasks

### Adding a New Route

1. Add handler to appropriate file in `handlers/`
2. Register in `routes/routes.go` via `mux.HandleFunc()`
3. For event-specific routes, extend `eventRouteHandler()` switch statement

### Database Changes

1. Modify schema in `db/createTables()`
2. Update model structs in `models/`
3. Adjust SQL queries in model methods
4. **Note**: No migrations—delete `./data/app.db` for fresh schema in dev

### Adding Model Methods

Follow existing pattern: receiver methods for single-entity operations, package functions for queries:

```go
func (e *Event) SaveEvent() error { ... }     // Instance method (saves new event)
func GetEventBySlug(slug string) (*Event, error) { ... }  // Package function (query)
```

**Model Query Patterns:**

- `models.GetActiveEventPreviews()` - Lightweight joins with submission counts for listings
- `models.GetAllActiveEventsWithCounts()` - Full event data + counts for admin views
- Models use `LEFT JOIN` with `COUNT()` aggregations for submission tallies

### Working with Time

Event dates use `time.Time` throughout:

- Database stores as SQLite `DATETIME` (no timezone suffix)
- Docker deployment sets `TZ=America/Los_Angeles` via environment
- Date parsing: `time.Parse("2006-01-02", dateString)` for form inputs
- Date queries: `models.GetEventsForToday()` compares `event_date` to `time.Now().Date`

**Important:** Date comparisons in queries need review—current implementation may have timezone offset bugs.

## Code Style Notes

- Error handling: Return errors up to handlers, log and return HTTP errors there
- SQL queries: Inline SQL in model methods (no ORM)
- No middleware: Direct handler registration in `routes/routes.go`
- Template data: Use anonymous structs in handlers for type safety
- Logging: Use `log.Printf()` for errors, `log.Println()` for info, `log.Fatal()` for initialization failures

## Dependencies

Core dependencies in `go.mod`:

- `github.com/mattn/go-sqlite3` - SQLite database driver
- `github.com/joho/godotenv` - Environment variable loading from `.env`
- `golang.org/x/image` - Image processing (resize, format conversion)

No web frameworks used - relies on `net/http` standard library.
