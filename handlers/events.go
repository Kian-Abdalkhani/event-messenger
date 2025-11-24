package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"event-messenger.com/models"
	"event-messenger.com/utils"
)

func CreateEventForm(w http.ResponseWriter, r *http.Request) {

	events, err := models.GetAllActiveEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error retreiving events %v", err)
		return
	}

	data := struct {
		Events []models.Event
	}{
		Events: events,
	}

	renderTemplate(w, "./templates/create_event_form.html", data)

}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// grab values from the form
	name := r.FormValue("name")
	slug := utils.GenerateSlug(name) // Auto-generate from name
	description := r.FormValue("description")
	coordinator := r.FormValue("coordinator")
	coordinatorContact := strings.TrimSpace(r.FormValue("coordinator_contact"))
	recipientName := r.FormValue("recipientName")
	recipientContact := strings.TrimSpace(r.FormValue("recipientContact"))

	// Validate required fields
	if name == "" || recipientName == "" || recipientContact == "" {
		http.Error(w, "Name, recipient name, and recipient contact are required", http.StatusBadRequest)
		return
	}

	if utils.ValidateEmail(recipientContact); err != nil {
		http.Error(w, "Invalid recipient Email", http.StatusBadRequest)
		return
	}

	if coordinatorContact != "" {
		if utils.ValidateEmail(coordinatorContact); err != nil {
			http.Error(w, "Invalid recipient Email", http.StatusBadRequest)
			return
		}

	}

	eventDate, err := time.Parse("2006-01-02", r.FormValue("event_date"))
	if err != nil {
		http.Error(w, "Invalid event date format", http.StatusBadRequest)
		return
	}

	// Fix: Event date should be in the future
	if eventDate.Before(time.Now().Truncate(24 * time.Hour)) {
		http.Error(w, "Event date must be in the future", http.StatusBadRequest)
		return
	}

	// finds funnel url, returns blank string if funnel is inactive
	funnel_url, err := utils.GetFunnelURL(slug)
	if err != nil {
		http.Error(w, "failed to retreive funnel active status", http.StatusInternalServerError)
		return
	}

	if funnel_url == "" {
		err = utils.CreateFunnel()
		if err != nil {
			http.Error(w, "Error funneling server to public internet", http.StatusInternalServerError)
			return
		}

		// retry fetching url after creating tunnel:
		funnel_url, err = utils.GetFunnelURL(slug)
		if err != nil {
			http.Error(w, "failed to retreive funnel active status", http.StatusInternalServerError)
			return
		}

	}
	event := models.NewEvent(
		name,
		slug,
		eventDate,
		models.WithDescription(description),
		models.WithCoordinator(coordinator, coordinatorContact),
		models.WithRecipient(recipientName, recipientContact),
		models.WithFunnelURL(funnel_url),
	)

	err = event.SaveEvent()
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/events/"+slug+"/created", http.StatusSeeOther)
}

// New handler for event creation success
func EventCreatedSuccess(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path: /events/{slug}/created
	slug := strings.TrimPrefix(r.URL.Path, "/events/")
	slug = strings.TrimSuffix(slug, "/created")
	slug = strings.Trim(slug, "/")

	event, err := models.GetEventBySlug(slug)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	data := struct {
		Event     models.Event
		FunnelURL string
	}{
		Event:     *event,
		FunnelURL: event.FunnelURL,
	}

	renderTemplate(w, "./templates/event_created_success.html", data)
}
