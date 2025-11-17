package routes

import (
	"net/http"
	"strings"

	"congrats-project.com/handlers"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public routes - Home/Landing
	mux.HandleFunc("/", handlers.HomeHandler) // Landing page listing active events

	// Event management routes (no auth required for LAN service)
	mux.HandleFunc("/events/create", handlers.CreateEventForm)
	mux.HandleFunc("/events/create/submit", handlers.CreateEvent)

	// Event-specific public routes
	mux.HandleFunc("/events/", eventRouteHandler) // Handles all /events/* routes

	// Static file serving
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./data/uploads"))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return mux
}

// eventRouteHandler routes all event-specific requests
func eventRouteHandler(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path: /events/{slug}/{action}
	path := r.URL.Path[len("/events/"):]

	// Handle /events/create separately (already handled above)
	if strings.HasPrefix(path, "create") {
		return
	}

	if path == "" || path == "/" {
		// List all events
		handlers.HomeHandler(w, r)
		return
	}

	// Parse slug and action
	slug, action := parseEventPath(path)

	switch action {
	case "":
		// GET /events/graduation-2025 - Show event page with form
		handlers.SubmissionFormHandler(w, r, slug)
	case "submit":
		// POST /events/graduation-2025/submit - Submit to event
		if r.Method == http.MethodPost {
			handlers.SubmissionHandler(w, r, slug)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	default:
		http.NotFound(w, r)
	}
}

// parseEventPath extracts slug and action from path
// Examples:
//
//	"graduation-2025" -> ("graduation-2025", "")
//	"graduation-2025/submit" -> ("graduation-2025", "submit")
//	"graduation-2025/messages" -> ("graduation-2025", "messages")
func parseEventPath(path string) (slug string, action string) {
	parts := strings.SplitN(path, "/", 2)
	slug = parts[0]
	if len(parts) > 1 {
		action = parts[1]
	}
	return slug, action
}
