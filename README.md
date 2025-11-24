# Event Messenger

A Go web application for collecting congratulatory messages and photos for special events like graduations, birthdays, weddings, and retirements. Built as a LAN-first service with an automated email notification system.

## Features

- **Event Creation**: Create events with shareable URLs for collecting submissions
- **Message & Photo Collection**: Users submit congratulatory messages and images to event-specific pages
- **Automated Notifications**: On the event date, recipients automatically receive an email with all submissions at 8AM
- **Image Optimization**: Automatic resizing and conversion of uploaded images (max 800px width, JPEG format)
- **Auto-Cleanup**: Events are automatically deleted 30 days after the notification email is sent
- **Tailscale Integration**: Built-in Tailscale support with automatic funnel management for secure remote access
- **Email Validation**: Server-side email validation for recipient and coordinator emails
- **No Authentication Required**: Designed for trusted LAN environments

## Quick Start

### Prerequisites

- Go 1.21+ (for development)
- Docker & Docker Compose (for production deployment)
- Tailscale account and auth key (optional, for remote access)

### Development Setup

1. Clone the repository:

```bash
git clone <repository-url>
cd event_messenger
```

2. Create a `.env` file in the project root:

```bash
BASE_URL=http://localhost:8080
WEB_PORT=8080
DB_PATH=./data/app.db
GO_ENV=development

# SMTP Configuration (required for email notifications)
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_EMAIL=noreply@example.com

# Tailscale Configuration (optional, for remote access)
TS_AUTHKEY=your_tailscale_auth_key
```

3. Install dependencies:

```bash
go mod download
```

4. Run the application:

```bash
go run main.go
```

5. Open your browser to `http://localhost:8080`

### Production Deployment (Docker)

1. Set environment variables in your system or create a `.env` file

2. Start the application:

```bash
docker compose up -d
```

3. Data persists in the `./data` directory via Docker volume

## Project Structure

```
event_messenger/
├── main.go                 # Application entry point
├── config/                 # Configuration management
│   └── config.go
├── db/                     # Database initialization
│   └── db.go
├── handlers/               # HTTP request handlers
│   ├── admin.go
│   ├── events.go
│   ├── home.go
│   ├── render.go
│   ├── submissions.go
│   └── view.go
├── models/                 # Data models and database queries
│   ├── event.go
│   └── submission.go
├── routes/                 # URL routing
│   └── routes.go
├── scheduler/              # Background job schedulers
│   ├── cleanup.go         # Auto-deletion of old events
│   ├── notification.go    # Email sending on event dates
│   └── scheduler.go       # Cron orchestration
├── utils/                  # Utility functions
│   ├── email.go           # SMTP email sending
│   ├── email_test.go      # Email validation tests
│   ├── funnel.go          # Tailscale funnel management
│   ├── slugs.go           # URL slug generation
│   └── url.go             # URL helpers
├── templates/              # HTML templates
│   ├── create_event_form.html
│   ├── email_notification.html
│   ├── home.html
│   ├── submission_form.html
│   └── success.html
└── data/                   # Application data (gitignored)
    ├── app.db             # SQLite database
    └── uploads/           # Uploaded images
```

## Usage Flow

1. **Create an Event**: Navigate to the home page and click "Create New Event"

   - Enter event name, date, recipient details, and coordinator information
   - System generates a unique shareable URL

2. **Share the URL**: Send the event URL to friends, family, or colleagues

3. **Collect Submissions**: People visit the URL and submit messages with photos

   - Images are automatically optimized (resized and converted to JPEG)
   - Maximum 10MB per image upload

4. **Automatic Email**: On the event date at 8AM, the recipient receives an email with:

   - All submitted messages and images
   - Up to 150 submissions (SMTP size limit protection)
   - Images embedded as base64 data URIs

5. **Auto-Cleanup**: 30 days after the email is sent, the event is automatically deleted

## Configuration

### Environment Variables

| Variable          | Required | Default         | Description                                    |
| ----------------- | -------- | --------------- | ---------------------------------------------- |
| `BASE_URL`        | Yes      | -               | Base URL for generating shareable links        |
| `WEB_PORT`        | Yes      | `8080`          | Port for the web server                        |
| `DB_PATH`         | Yes      | `./data/app.db` | Path to SQLite database                        |
| `GO_ENV`          | No       | `development`   | Environment mode (development/production)      |
| `SMTP_SERVER`     | Yes\*    | -               | SMTP server hostname                           |
| `SMTP_PORT`       | Yes\*    | -               | SMTP server port                               |
| `SMTP_USERNAME`   | Yes\*    | -               | SMTP authentication username                   |
| `SMTP_PASSWORD`   | Yes\*    | -               | SMTP authentication password                   |
| `SMTP_FROM_EMAIL` | Yes\*    | -               | From address for notification emails           |
| `TS_AUTHKEY`      | No       | -               | Tailscale authentication key for remote access |

\*Required for email notifications to work

### Gmail SMTP Setup

To use Gmail for sending emails:

1. Enable 2-factor authentication on your Google account
2. Generate an App Password: https://myaccount.google.com/apppasswords
3. Use these settings:
   ```
   SMTP_SERVER=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your_email@gmail.com
   SMTP_PASSWORD=your_app_password
   ```

## API Endpoints

| Method | Path                    | Description                       |
| ------ | ----------------------- | --------------------------------- |
| `GET`  | `/`                     | List all active events            |
| `GET`  | `/events/create`        | Show event creation form          |
| `POST` | `/events/create/submit` | Create a new event                |
| `GET`  | `/events/{slug}`        | Show submission form for an event |
| `POST` | `/events/{slug}/submit` | Submit a message and photo        |
| `GET`  | `/uploads/*`            | Serve uploaded images             |

## Tailscale Integration

The application includes built-in Tailscale support for secure remote access:

### Features

- **Automatic Funnel Creation**: When an event is created, Tailscale funnel is automatically enabled
- **Automatic Funnel Removal**: When all events become inactive, the funnel is automatically disabled
- **Public URLs**: Event creation displays both local and Tailscale funnel URLs for sharing
- **Zero Configuration**: Funnel management happens automatically in the background

### Setup

1. Get a Tailscale auth key from https://login.tailscale.com/admin/settings/keys
2. Add `TS_AUTHKEY` to your `.env` file or environment variables
3. Start the application with `docker compose up -d`
4. The Tailscale container will connect to your tailnet automatically
5. Funnels are created/removed based on active events

### How It Works

- **Event Creation**: Calls `utils.CreateFunnel()` to enable public HTTPS access via Tailscale
- **Event Deactivation**: Calls `utils.RemoveFunnel()` when no active events remain
- **URL Generation**: Success page shows both LAN and public funnel URLs
- **Network Sharing**: Main app container shares Tailscale's network namespace

### Docker Configuration

The `docker-compose.yml` includes two services:

- `event-messenger`: Main application (shares Tailscale's network)
- `tailscale`: Tailscale client with funnel support

The application container has access to Docker socket for funnel management via `docker exec` commands.

## Database Schema

### Events Table

- `id` - Primary key
- `name` - Event name
- `slug` - URL-friendly identifier (unique)
- `description` - Optional event description
- `event_date` - Date when email should be sent
- `recipient_name` - Name of the person receiving the email
- `recipient_email` - Email address for notifications
- `coordinator_name` - Name of event organizer
- `coordinator_contact` - Contact info for organizer
- `active` - Boolean flag (inactive after email sent)
- `email_sent` - Boolean flag
- `email_sent_at` - Timestamp of email delivery
- `created_at` - Creation timestamp

### Submissions Table

- `id` - Primary key
- `event_id` - Foreign key to events table
- `submitter_name` - Name of person submitting
- `message` - Congratulatory message
- `image_filename` - Filename of uploaded image
- `created_at` - Submission timestamp

## Background Schedulers

### Email Notification Scheduler

- Runs daily at 8AM system time
- Also runs immediately on application startup (for testing)
- Sends emails for events matching today's date
- Marks events as inactive after sending
- Logs email size (warns if >15MB)

### Cleanup Scheduler

- Runs weekly at 2AM system time
- Deletes events 30+ days after email was sent
- Cascade deletes all associated submissions and images
- Safety check enforces 30-day grace period

## Development Notes

- **No authentication**: Designed for trusted LAN environments
- **SQLite database**: Lightweight, no separate database server needed
- **No ORM**: Direct SQL queries in model methods
- **No web framework**: Built with Go's `net/http` standard library
- **Template rendering**: HTML templates parsed on each request (no caching in dev)
- **Timezone handling**: All timestamps stored in system timezone, converted from UTC
- **Email validation**: Server-side validation using regex for recipient and coordinator emails
- **Multi-stage Docker build**: Optimized image size using builder pattern
- **Tailscale integration**: Automatic funnel management for secure remote access

## Image Processing

Uploaded images are automatically processed:

- Max file size: 10MB
- Allowed formats: JPEG, PNG, GIF, WebP
- Auto-resize: Images wider than 800px are scaled down (maintains aspect ratio)
- Format conversion: All images converted to JPEG at 85% quality
- Saved as: `{unix_timestamp}_{original_filename}.jpg`

## Dependencies

- `github.com/mattn/go-sqlite3` - SQLite database driver (CGO required)
- `github.com/joho/godotenv` - Load environment variables from `.env`
- `golang.org/x/image` - Image processing and manipulation

**Note**: The Go module path is `event-messenger.com` (see `go.mod`)

## Recent Improvements

### Technical Enhancements

- **Multi-stage Docker Build**: Reduced image size and improved build speed using separate build and runtime stages
- **Timezone Awareness**: Database queries properly handle system timezone when storing and filtering event dates
- **Email Validation**: Server-side regex validation for recipient and coordinator emails with unit tests in `utils/email_test.go`
- **Tailscale Auto-Funnel**: Automatic public HTTPS URL generation via Tailscale funnel, with automatic cleanup when events become inactive
- **Improved Logging**: Enhanced logging throughout the application with sensitive data (submission messages, image filenames) excluded from logs
- **Template Path Fixes**: Corrected template loading for both development (`go run`) and production (compiled binary) environments

### Known Limitations

- **No Pagination**: Homepage lists all active events - could scale poorly with many events
- **No Orphan File Cleanup**: If database records are deleted but image files remain, no automatic cleanup occurs
- **Email Size Constraints**: Base64 encoding inflates image size by ~33%, necessitating 150 submission per email cap
- **Docker Socket Requirement**: Tailscale funnel management requires Docker socket access in container
- **CGO Dependency**: SQLite driver requires C compiler, making cross-compilation more complex

## License

MIT License

Copyright (c) 2025 Kian Abdalkhani

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Contributing

[Add contribution guidelines here]
